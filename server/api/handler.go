package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"system_api/config"
	"system_api/model"
	"system_api/service"
	"system_api/store"

	"github.com/gin-gonic/gin"
)

func CreateQuestion(c *gin.Context, db *sql.DB) {
	var ask model.Ask

	//接受并验证json
	if err := c.ShouldBindJSON(&ask); err != nil {
		HandleAskError(c, err)
		return
	}
	//获取api的资源
	ai, err := config.Envinit()
	if err != nil {
		c.JSON(-1, model.Response{
			Code: -101,
			Msg:  ".env 读取失败",
			Data: nil,
		})
		return
	}

	//访问ai
	topics, err := service.VisitAi(ai, ask)
	if err != nil {
		c.JSON(-1, model.Response{
			Code: -102,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "success",
		Data: topics,
	})
}

func GetQuestions(c *gin.Context, db *sql.DB) {
	res, err := db.Query("SELECT id,title,answers,right,type_id,difficulty,is_ai,language,keyword,active FROM questions WHERE active = 1")
	if err != nil {
		c.JSON(-1, model.Response{
			Code: -104,
			Msg:  "数据库查询失败",
			Data: nil,
		})
		return
	}
	defer res.Close()

	var topics []model.Topic
	for res.Next() {
		var t model.Topic
		var answers, right []byte
		if err := res.Scan(&t.Id, &t.Title, &answers, &right, &t.Typeid, &t.Difficulty, &t.Isai, &t.Language, &t.Keyword, &t.Active); err != nil {
			c.JSON(-1, model.Response{
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

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "success",
		Data: topics,
	})
}

func AddQuestion(c *gin.Context, db *sql.DB) {
	var topic model.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		HandleTopicError(c, err)
		return
	}

	topic.Active = 1

	var topics []model.Topic
	topics = append(topics, topic)

	//保存到服务端
	err := store.SaveToDB(db, topics)
	if err != nil {
		c.JSON(-1, model.Response{
			Code: -103,
			Msg:  "保存到数据库失败",
			Data: nil,
		})
		return
	}
	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "题目添加成功",
		Data: nil,
	})
}

func EditQuestion(c *gin.Context, db *sql.DB) {
	var updateTopic model.Topic //需要更新的题目
	if err := c.ShouldBindJSON(&updateTopic); err != nil {
		HandleTopicError(c, err)
		return
	}

	// 验证题目ID是否存在
	if updateTopic.Id == 0 {
		c.JSON(-1, model.Response{
			Code: -106,
			Msg:  "题目ID不能为空",
			Data: nil,
		})
		return
	}

	// 从数据库查询现有题目
	var existingTopic model.Topic //数据库中的原题
	var answers, right []byte
	err := db.QueryRow("SELECT id, title, answers, right, type_id, difficulty, is_ai, language, keyword, active FROM questions WHERE id = ? AND active = 1", updateTopic.Id).
		Scan(&existingTopic.Id, &existingTopic.Title, &answers, &right, &existingTopic.Typeid, &existingTopic.Difficulty, &existingTopic.Isai, &existingTopic.Language, &existingTopic.Keyword, &existingTopic.Active)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(-1, model.Response{
				Code: -108,
				Msg:  "题目不存在或已删除",
				Data: nil,
			})
		} else {
			c.JSON(-1, model.Response{
				Code: -107,
				Msg:  "查询题目失败",
				Data: nil,
			})
		}
		return
	}

	// 解析JSON字段
	if err := json.Unmarshal(answers, &existingTopic.Answers); err != nil {
		c.JSON(-1, model.Response{
			Code: -109,
			Msg:  "解析答案选项失败",
			Data: nil,
		})
		return
	}
	if err := json.Unmarshal(right, &existingTopic.Right); err != nil {
		c.JSON(-1, model.Response{
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

	err1 := store.UpdateToDB(db, existingTopic)
	if err1 != nil {
		c.JSON(-1, model.Response{
			Code: -103,
			Msg:  "更新到数据库失败",
			Data: nil,
		})
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "题目更新成功",
		Data: nil,
	})
}

func DeleteQuestion(c *gin.Context, db *sql.DB) {
	var t model.Topic
	if err := c.ShouldBindJSON(&t); err != nil {
		HandleTopicError(c, err)
		if t.Id == 0 {
			c.JSON(-1, model.Response{
				Code: -106,
				Msg:  "传入ID不能为空",
				Data: nil,
			})
		}
		return
	}

	// 验证题目ID是否存在
	if t.Id == 0 {
		c.JSON(-1, model.Response{
			Code: -106,
			Msg:  "该题目ID不存在",
			Data: nil,
		})
		return
	}

	t.Active = 0
	err := store.UpdateToDB(db, t)
	if err != nil {
		c.JSON(-1, model.Response{
			Code: -103,
			Msg:  "删除失败",
			Data: nil,
		})
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "题目删除成功",
		Data: nil,
	})
}
