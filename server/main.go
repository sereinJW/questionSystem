package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	_ "modernc.org/sqlite"
)

const sqliteDB = "questionSystem.db"

var Choices [4]string = [4]string{"", "单选题", "多选题", "编程题"}
var Difficulties [4]string = [4]string{"", "简单", "中等", "困难"}

// ai请求体
type Ask struct { //omitempty 允许客户端不传该字段,default=xxx 为零值时自动填充默认值,required 表示必填
	Number     int    `json:"number" validate:"required"`
	Language   string `json:"language" validate:"omitempty,oneof=go javascript java python c++" default:"go"`
	Type       int    `json:"type" validate:"omitempty,oneof=1 2 3" default:"1"`
	Difficulty int    `json:"difficulty" validate:"omitempty,oneof=1 2 3" default:"1"`
	Keyword    string `json:"keyword" validate:"required"`
}

// 题目
type Topic struct {
	Id         int      `json:"id"`                                                                //题目ID
	Title      string   `json:"title" validate:"required"`                                         //题干
	Answers    []string `json:"answers"`                                                           //选项
	Right      []string `json:"right"`                                                             //正确的选项
	Typeid     int      `json:"type_id" validate:"required"`                                       //题目类型
	Difficulty int      `json:"difficulty" validate:"required"`                                    //题目难度
	Isai       int      `json:"is_ai"`                                                             //ai还是手工
	Language   string   `json:"language" validate:"omitempty,oneof=go javascript java python c++"` //编译语言
	Keyword    string   `json:"keyword" validate:"required"`                                       //关键词
	Active     int      `json:"active"`                                                            //是否被删除
}

// ai模型配置
type Ai struct {
	api_key string
	url     string
	model   string
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 请求体字段验证错误处理
func handleAskError(c *gin.Context, err error) {
	// 处理验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Field() {
			case "Number":
				c.JSON(-1, Response{-101, "number不能为空，至少为1", nil})
			case "Language":
				c.JSON(-1, Response{-102, "language参数错误，language必须是go/javascript/java/python/c++", nil})
			case "Type":
				c.JSON(-1, Response{-103, "type参数错误，type必须是1/2/3", nil})
			case "Keyword":
				c.JSON(-1, Response{-104, "keyword不能为空", nil})
			case "Difficulty":
				c.JSON(-1, Response{-105, "Difficulty参数错误，Difficulty必须是1/2/3", nil})
			}
		}
	} else {
		c.JSON(-1, Response{-400, "请求格式错误", nil})
	}
}
func handleTopicError(c *gin.Context, err error) {
	// 处理验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Field() {
			case "Title":
				c.JSON(-1, Response{-101, "title不能为空", nil})
			case "Language":
				c.JSON(-1, Response{-102, "language参数错误，language必须是go/javascript/java/python/c++", nil})
			case "Typeid":
				c.JSON(-1, Response{-103, "type不能为空", nil})
			case "Keyword":
				c.JSON(-1, Response{-104, "keyword不能为空", nil})
			case "Difficulty":
				c.JSON(-1, Response{-105, "Difficulty不能为空", nil})
			}
		}
	} else {
		c.JSON(-1, Response{-400, "请求格式错误", nil})
	}
}

// 读取env文件
func envinit() (Ai, error) {
	if err := godotenv.Load(); err != nil {
		return Ai{}, fmt.Errorf("读取env文件失败：%w", err)
	}

	var ai Ai
	ai = Ai{
		os.Getenv("DEEPSEEK_API_KEY"),
		os.Getenv("DEEPSEEK_URL"),
		os.Getenv("DEEPSEEK_MODEL"),
	}

	if ai.api_key == "" || ai.url == "" || ai.model == "" {
		return Ai{}, fmt.Errorf("env配置不完整")
	}

	return ai, nil
}

// 生成对应message
func toString(ask Ask) string {
	if ask.Type == 3 {
		return fmt.Sprintf(`请生成 %d 道难度为 %s 的 %s 语言的关于 %s 知识点的 %s ，并以JSON数组格式输出，每个题目包含以下字段：title(题干)。示例格式：[{"title":"请用GO实现冒泡排序，代码编程实现"}]。只输出JSON数组，不需要其他内容。`, ask.Number, Difficulties[ask.Difficulty], ask.Language, ask.Keyword, Choices[ask.Type])
	}
	return fmt.Sprintf(`请生成 %d 道难度为 %s 的 %s 语言的关于 %s 知识点的 %s ，并以JSON数组格式输出，每个题目包含以下字段：title(题干),answers(选项数组),right(正确答案数组)。示例格式：[{"title":"Gin框架的作用是什么","answers":["A.Gin是一个用于前端开发的JavaScript框架，类似React或Vue","B.Gin是Go语言的高性能HTTP Web框架，支持路由分组、中间件等功能","C.Gin主要用于数据库操作，是ORM工具的一种","D.Gin是Python的异步Web框架，类似Django或Flask"],"right":["B"]}]。只输出JSON数组，不需要其他内容。`, ask.Number, Difficulties[ask.Difficulty], ask.Language, ask.Keyword, Choices[ask.Type])
}

// 访问AI并处理
func visitAi(ai Ai, ask Ask) ([]Topic, error) {
	//根据申请生成对应message
	message := toString(ask)

	//启动ai访问
	client := openai.NewClient(
		option.WithAPIKey(ai.api_key),
		option.WithBaseURL(ai.url),
	)

	//正式访问
	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(message),
			},
			Model: ai.model,
		})
	if err != nil {
		return nil, fmt.Errorf("AI访问失败：%w", err)
	}

	content := chatCompletion.Choices[0].Message.Content

	// 预处理返回的内容
	content = strings.TrimSpace(content)
	// 去除可能的代码块标记
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	//解析ai响应
	var topics []Topic
	if err := json.Unmarshal([]byte(content), &topics); err != nil {
		return nil, fmt.Errorf("解析ai返回的JSON失败：%w", err)
	}

	// 验证返回的数据
	if len(topics) == 0 {
		return nil, fmt.Errorf("AI 返回的数据为空")
	}

	for i := range topics {
		if ask.Type == 3 { //编程题没有answers和right
			if topics[i].Title == "" {
				return nil, fmt.Errorf("AI 返回的数据不完整: title=%v", topics[i].Title)
			}

		} else {
			if topics[i].Title == "" || len(topics[i].Answers) == 0 || len(topics[i].Right) == 0 {
				return nil, fmt.Errorf("AI 返回的数据不完整: title=%v, answers=%v, right=%v",
					topics[i].Title, topics[i].Answers, topics[i].Right)
			}
		}

		// 设置topic的属性
		topics[i].Typeid = ask.Type
		topics[i].Language = ask.Language
		topics[i].Difficulty = ask.Difficulty
		topics[i].Keyword = ask.Keyword
		topics[i].Isai = 1
		topics[i].Active = 1
	}

	return topics, nil
}

// 初始化数据库
func initDb(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			answers JSON NOT NULL,
			right JSON NOT NULL,
			type_id INTEGER NOT NULL,
			difficulty INTEGER NOT NULL,
			is_ai INTEGER NOT NULL,
			language TEXT NOT NULL,
			keyword TEXT,
			active INTEGER NOT NULL
		)
	`)

	if err != nil {
		return fmt.Errorf("初始化数据库错误: %w", err)
	}
	return nil
}

// 保存到数据库
func saveToDB(db *sql.DB, topics []Topic) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, topic := range topics {
		// 将 answers 和 right 数组转换为 JSON 字符串
		answersJSON, err := json.Marshal(topic.Answers)
		if err != nil {
			return fmt.Errorf("序列化答案选项失败：%w", err)
		}
		rightJSON, err := json.Marshal(topic.Right)
		if err != nil {
			return fmt.Errorf("序列化正确答案失败：%w", err)
		}

		// 执行插入操作
		_, err = tx.Exec(`
			INSERT INTO questions (
				title, answers, right, type_id, difficulty, 
				is_ai, language, keyword, active
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			topic.Title,
			string(answersJSON),
			string(rightJSON),
			topic.Typeid,
			topic.Difficulty,
			topic.Isai,
			topic.Language,
			topic.Keyword,
			topic.Active,
		)

		if err != nil {
			return fmt.Errorf("插入数据失败：%w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败：%w", err)
	}

	return nil
}

// 更新到数据库
func UpdateToDB(db *sql.DB, topic Topic) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 将 answers 和 right 数组转换为 JSON 字符串
	answersJSON, err := json.Marshal(topic.Answers)
	if err != nil {
		return fmt.Errorf("序列化答案选项失败：%w", err)
	}
	rightJSON, err := json.Marshal(topic.Right)
	if err != nil {
		return fmt.Errorf("序列化正确答案失败：%w", err)
	}

	// 执行更新操作
	_, err = tx.Exec(`
		UPDATE questions 
		SET title = ?, answers = ?, right = ?, type_id = ?, difficulty = ?, 
			is_ai = ?, language = ?, keyword = ?, active = ?
		WHERE id = ?
	`,
		topic.Title,
		string(answersJSON),
		string(rightJSON),
		topic.Typeid,
		topic.Difficulty,
		topic.Isai,
		topic.Language,
		topic.Keyword,
		topic.Active,
		topic.Id,
	)

	if err != nil {
		return fmt.Errorf("更新数据失败：%w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败：%w", err)
	}

	return nil
}

func main() {
	//初始化
	validate := validator.New()

	db, err := sql.Open("sqlite", sqliteDB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDb(db); err != nil {
		fmt.Println(err)
		return
	}

	r := gin.Default()

	// 获取静态文件的绝对路径
	clientDistPath := "../client/dist"
	absPath, err := filepath.Abs(clientDistPath)
	if err != nil {
		fmt.Printf("获取静态文件路径失败: %v\n", err)
	} else {
		fmt.Printf("静态文件路径: %s\n", absPath)
	}

	// 检查路径是否存在
	if _, err := os.Stat(clientDistPath); os.IsNotExist(err) {
		fmt.Printf("警告: 静态文件目录不存在: %s\n", clientDistPath)
	} else {
		fmt.Printf("静态文件目录存在: %s\n", clientDistPath)
	}

	// 提供前端静态文件
	r.Static("/assets", filepath.Join(clientDistPath, "assets"))
	r.StaticFile("/", filepath.Join(clientDistPath, "index.html"))
	r.StaticFile("/index.html", filepath.Join(clientDistPath, "index.html"))
	r.StaticFile("/vite.svg", filepath.Join(clientDistPath, "vite.svg"))

	// 添加路由处理404请求
	r.NoRoute(func(c *gin.Context) {
		// 对于任何未匹配的路由都返回index.html
		c.File(filepath.Join(clientDistPath, "index.html"))
	})

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		*v = *validate
	}

	//调用ai的API
	r.POST("/api/questions/create", func(c *gin.Context) {

		var ask Ask

		//接受并验证json
		if err := c.ShouldBindJSON(&ask); err != nil {
			handleAskError(c, err)
			return
		}
		//获取api的资源
		ai, err := envinit()
		if err != nil {
			c.JSON(-1, Response{
				Code: -101,
				Msg:  ".env 读取失败",
				Data: nil,
			})
			return
		}

		//访问ai
		topics, err := visitAi(ai, ask)
		if err != nil {
			c.JSON(-1, Response{
				Code: -102,
				Msg:  err.Error(),
				Data: nil,
			})
			return
		}
		//保存到服务端
		// err1 := saveToDB(db, topics)
		// if err1 != nil {
		// 	c.JSON(-1, Response{
		// 		Code: -103,
		// 		Msg:  "保存到数据库失败",
		// 		Data: nil,
		// 	})
		// 	return
		// }

		c.JSON(200, Response{
			Code: 0,
			Msg:  "success",
			Data: topics,
		})

	})

	//查询题库接口
	r.GET("/api/questions", func(c *gin.Context) {

		res, err := db.Query("SELECT id,title,answers,right,type_id,difficulty,is_ai,language,keyword,active FROM questions WHERE active = 1")
		if err != nil {
			c.JSON(-1, Response{
				Code: -104,
				Msg:  "数据库查询失败",
				Data: nil,
			})
			return
		}
		defer res.Close()

		var topics []Topic
		for res.Next() {
			var t Topic
			var answers, right []byte
			if err := res.Scan(&t.Id, &t.Title, &answers, &right, &t.Typeid, &t.Difficulty, &t.Isai, &t.Language, &t.Keyword, &t.Active); err != nil {
				c.JSON(-1, Response{
					Code: -105,
					Msg:  "查询读取失败",
					Data: nil,
				})
				return
			}
			err := json.Unmarshal(answers, &t.Answers) //将JSON转为string切片
			if err != nil {
				fmt.Println("题目答案反序列化失败")
			}
			err = json.Unmarshal(right, &t.Right)
			if err != nil {
				fmt.Println("正确答案反序列化失败")
			}
			topics = append(topics, t)
		}

		c.JSON(200, Response{
			Code: 0,
			Msg:  "success",
			Data: topics,
		})
	})

	//添加题目接口
	r.POST("/api/questions/add", func(c *gin.Context) {
		var topic Topic
		if err := c.ShouldBindJSON(&topic); err != nil {
			handleTopicError(c, err)
			return
		}

		//topic.Isai = 0
		topic.Active = 1

		var topics []Topic
		topics = append(topics, topic)

		//保存到服务端
		err := saveToDB(db, topics)
		if err != nil {
			c.JSON(-1, Response{
				Code: -103,
				Msg:  "保存到数据库失败",
				Data: nil,
			})
			return
		}
		c.JSON(200, Response{
			Code: 0,
			Msg:  "题目添加成功",
			Data: nil,
		})
	})

	//编辑题目接口
	r.POST("/api/questions/edit", func(c *gin.Context) {
		var updateTopic Topic //需要更新的题目
		if err := c.ShouldBindJSON(&updateTopic); err != nil {
			handleTopicError(c, err)
			return
		}

		// 验证题目ID是否存在
		if updateTopic.Id == 0 {
			c.JSON(-1, Response{
				Code: -106,
				Msg:  "题目ID不能为空",
				Data: nil,
			})
			return
		}

		// 从数据库查询现有题目
		var existingTopic Topic //数据库中的原题
		var answers, right []byte
		err := db.QueryRow("SELECT id, title, answers, right, type_id, difficulty, is_ai, language, keyword, active FROM questions WHERE id = ? AND active = 1", updateTopic.Id).
			Scan(&existingTopic.Id, &existingTopic.Title, &answers, &right, &existingTopic.Typeid, &existingTopic.Difficulty, &existingTopic.Isai, &existingTopic.Language, &existingTopic.Keyword, &existingTopic.Active)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(-1, Response{
					Code: -108,
					Msg:  "题目不存在或已删除",
					Data: nil,
				})
			} else {
				c.JSON(-1, Response{
					Code: -107,
					Msg:  "查询题目失败",
					Data: nil,
				})
			}
			return
		}

		// 解析JSON字段
		if err := json.Unmarshal(answers, &existingTopic.Answers); err != nil {
			c.JSON(-1, Response{
				Code: -109,
				Msg:  "解析答案选项失败",
				Data: nil,
			})
			return
		}
		if err := json.Unmarshal(right, &existingTopic.Right); err != nil {
			c.JSON(-1, Response{
				Code: -110,
				Msg:  "解析正确答案失败",
				Data: nil,
			})
			return
		}

		// 更新非零值字段
		if updateTopic.Title != "" {
			existingTopic.Title = updateTopic.Title
		}
		if len(updateTopic.Answers) > 0 {
			existingTopic.Answers = updateTopic.Answers
		}
		if len(updateTopic.Right) > 0 {
			existingTopic.Right = updateTopic.Right
		}
		if updateTopic.Typeid != 0 {
			existingTopic.Typeid = updateTopic.Typeid
		}
		if updateTopic.Difficulty != 0 {
			existingTopic.Difficulty = updateTopic.Difficulty
		}
		if updateTopic.Language != "" {
			existingTopic.Language = updateTopic.Language
		}
		if updateTopic.Keyword != "" {
			existingTopic.Keyword = updateTopic.Keyword
		}

		err1 := UpdateToDB(db, existingTopic)
		if err1 != nil {
			c.JSON(-1, Response{
				Code: -103,
				Msg:  "更新到数据库失败",
				Data: nil,
			})
			return
		}

		c.JSON(200, Response{
			Code: 0,
			Msg:  "题目更新成功",
			Data: nil,
		})
	})

	//删除题目接口
	r.POST("/api/questions/delete", func(c *gin.Context) {
		var t Topic
		if err := c.ShouldBindJSON(&t); err != nil {
			handleTopicError(c, err)
			if t.Id == 0 {
				c.JSON(-1, Response{
					Code: -106,
					Msg:  "传入ID不能为空",
					Data: nil,
				})
			}
			return
		}

		// 验证题目ID是否存在
		if t.Id == 0 {
			c.JSON(-1, Response{
				Code: -106,
				Msg:  "该题目ID不存在",
				Data: nil,
			})
			return
		}

		t.Active = 0
		err := UpdateToDB(db, t)
		if err != nil {
			c.JSON(-1, Response{
				Code: -103,
				Msg:  "删除失败",
				Data: nil,
			})
			return
		}

		c.JSON(200, Response{
			Code: 0,
			Msg:  "题目删除成功",
			Data: nil,
		})
	})

	r.Run(":8080")
}
