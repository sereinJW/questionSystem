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

// 创建试卷
func CreatePaper(c *gin.Context, db *sql.DB) {
	var request model.CreatePaperRequest

	// 接受并验证JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		HandleTopicError(c, err)
		return
	}

	// 验证题目ID是否存在且有效
	if err := validateQuestionIds(db, request.QuestionIds); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("题目验证失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	// 验证题目数量是否匹配配置
	totalRequired := 0
	for _, section := range request.Sections {
		totalRequired += section.Count
	}
	if len(request.QuestionIds) != totalRequired {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("题目数量不匹配，配置需要 %d 道题，但提供了 %d 道", totalRequired, len(request.QuestionIds)),
			Data: nil,
		})
		return
	}

	// 创建试卷
	paper, err := store.CreatePaperToDB(db, request)
	if err != nil {
		c.JSON(500, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("创建试卷失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "创建试卷成功",
		Data: paper,
	})
}

// 获取试卷详情
func GetPaperDetail(c *gin.Context, db *sql.DB) {
	// 获取路径参数中的试卷ID
	paperIdStr := c.Param("id")
	if paperIdStr == "" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID不能为空",
			Data: nil,
		})
		return
	}

	// 转换为整数
	var paperId int
	if _, err := fmt.Sscanf(paperIdStr, "%d", &paperId); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID格式错误",
			Data: nil,
		})
		return
	}

	// 查询试卷详情
	paperDetail, err := store.GetPaperDetailFromDB(db, paperId)
	if err != nil {
		if err.Error() == "试卷不存在" {
			c.JSON(404, model.Response{
				Code: -1,
				Msg:  "试卷不存在",
				Data: nil,
			})
		} else {
			c.JSON(500, model.Response{
				Code: -1,
				Msg:  fmt.Sprintf("获取试卷详情失败: %s", err.Error()),
				Data: nil,
			})
		}
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "获取试卷详情成功",
		Data: paperDetail,
	})
}

// 获取试卷列表
func GetPapers(c *gin.Context, db *sql.DB) {
	papers, err := store.GetPapersFromDB(db)
	if err != nil {
		c.JSON(500, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("获取试卷列表失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "获取试卷列表成功",
		Data: papers,
	})
}

// 辅助函数：验证题目ID是否存在且有效
func validateQuestionIds(db *sql.DB, questionIds []int) error {
	if len(questionIds) == 0 {
		return fmt.Errorf("题目ID列表不能为空")
	}

	// 构建占位符字符串
	placeholders := ""
	args := make([]interface{}, len(questionIds))
	for i, id := range questionIds {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// 查询有效的题目数量
	query := fmt.Sprintf("SELECT COUNT(*) FROM questions WHERE id IN (%s) AND active = 1", placeholders)
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询题目失败: %w", err)
	}

	if count != len(questionIds) {
		return fmt.Errorf("部分题目不存在或已被删除，预期 %d 道题，实际找到 %d 道", len(questionIds), count)
	}

	return nil
}

// 导出试卷
func ExportPaper(c *gin.Context, db *sql.DB) {
	// 获取路径参数中的试卷ID
	paperIdStr := c.Param("id")
	if paperIdStr == "" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID不能为空",
			Data: nil,
		})
		return
	}

	// 转换为整数
	var paperId int
	if _, err := fmt.Sscanf(paperIdStr, "%d", &paperId); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID格式错误",
			Data: nil,
		})
		return
	}

	// 获取导出格式
	format := c.Query("format")
	if format == "" {
		format = "docx" // 默认Word格式
	}

	// 验证格式
	if format != "docx" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "不支持的导出格式，仅支持 docx",
			Data: nil,
		})
		return
	}

	// 查询试卷详情
	paperDetail, err := store.GetPaperDetailFromDB(db, paperId)
	if err != nil {
		if err.Error() == "试卷不存在" {
			c.JSON(404, model.Response{
				Code: -1,
				Msg:  "试卷不存在",
				Data: nil,
			})
		} else {
			c.JSON(500, model.Response{
				Code: -1,
				Msg:  fmt.Sprintf("获取试卷详情失败: %s", err.Error()),
				Data: nil,
			})
		}
		return
	}

	// 生成Word文件
	fileData, err := service.ExportPaperToWord(paperDetail)
	if err != nil {
		c.JSON(500, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("生成Word文档失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	contentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	fileName := fmt.Sprintf("%s.docx", paperDetail.Paper.Title)

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// 返回文件数据
	c.Data(200, contentType, fileData)
}

// 删除试卷
func DeletePaper(c *gin.Context, db *sql.DB) {
	// 获取路径参数中的试卷ID
	paperIdStr := c.Param("id")
	if paperIdStr == "" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID不能为空",
			Data: nil,
		})
		return
	}

	// 转换为整数
	var paperId int
	if _, err := fmt.Sscanf(paperIdStr, "%d", &paperId); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID格式错误",
			Data: nil,
		})
		return
	}

	// 删除试卷
	err := store.DeletePaperFromDB(db, paperId)
	if err != nil {
		if err.Error() == "试卷不存在" {
			c.JSON(404, model.Response{
				Code: -1,
				Msg:  "试卷不存在",
				Data: nil,
			})
		} else {
			c.JSON(500, model.Response{
				Code: -1,
				Msg:  fmt.Sprintf("删除试卷失败: %s", err.Error()),
				Data: nil,
			})
		}
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "删除试卷成功",
		Data: nil,
	})
}

// 编辑试卷标题
func EditPaper(c *gin.Context, db *sql.DB) {
	// 获取路径参数中的试卷ID
	paperIdStr := c.Param("id")
	if paperIdStr == "" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID不能为空",
			Data: nil,
		})
		return
	}

	// 转换为整数
	var paperId int
	if _, err := fmt.Sscanf(paperIdStr, "%d", &paperId); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID格式错误",
			Data: nil,
		})
		return
	}

	// 解析请求体
	var request struct {
		Title string `json:"title" validate:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "请求参数错误",
			Data: nil,
		})
		return
	}

	// 更新试卷标题
	err := store.UpdatePaperTitleInDB(db, paperId, request.Title)
	if err != nil {
		if err.Error() == "试卷不存在" {
			c.JSON(404, model.Response{
				Code: -1,
				Msg:  "试卷不存在",
				Data: nil,
			})
		} else {
			c.JSON(500, model.Response{
				Code: -1,
				Msg:  fmt.Sprintf("更新试卷失败: %s", err.Error()),
				Data: nil,
			})
		}
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "更新试卷成功",
		Data: nil,
	})
}

// 更新试卷题目
func UpdatePaperQuestions(c *gin.Context, db *sql.DB) {
	// 获取路径参数中的试卷ID
	paperIdStr := c.Param("id")
	if paperIdStr == "" {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID不能为空",
			Data: nil,
		})
		return
	}

	// 转换为整数
	var paperId int
	if _, err := fmt.Sscanf(paperIdStr, "%d", &paperId); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "试卷ID格式错误",
			Data: nil,
		})
		return
	}

	// 解析请求体
	var request struct {
		Title       string `json:"title"`
		QuestionIds []int  `json:"question_ids" validate:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  "请求参数错误",
			Data: nil,
		})
		return
	}

	// 验证题目ID是否存在且有效
	if err := validateQuestionIds(db, request.QuestionIds); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("题目验证失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	// 更新试卷
	err := store.UpdatePaperQuestionsInDB(db, paperId, request.Title, request.QuestionIds)
	if err != nil {
		if err.Error() == "试卷不存在" {
			c.JSON(404, model.Response{
				Code: -1,
				Msg:  "试卷不存在",
				Data: nil,
			})
		} else {
			c.JSON(500, model.Response{
				Code: -1,
				Msg:  fmt.Sprintf("更新试卷失败: %s", err.Error()),
				Data: nil,
			})
		}
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "更新试卷成功",
		Data: nil,
	})
}

// AI生成试卷
func AIGeneratePaper(c *gin.Context, db *sql.DB) {
	var request service.AIGeneratePaperRequest

	// 接受并验证JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("请求参数错误: %s", err.Error()),
			Data: nil,
		})
		return
	}

	// 获取AI配置
	ai, err := config.Envinit()
	if err != nil {
		c.JSON(500, model.Response{
			Code: -101,
			Msg:  ".env 读取失败",
			Data: nil,
		})
		return
	}

	// 调用AI生成试卷服务
	paper, err := service.QuickAIGeneratePaper(db, ai, request)
	if err != nil {
		c.JSON(500, model.Response{
			Code: -1,
			Msg:  fmt.Sprintf("AI生成试卷失败: %s", err.Error()),
			Data: nil,
		})
		return
	}

	c.JSON(200, model.Response{
		Code: 0,
		Msg:  "AI生成试卷成功",
		Data: paper,
	})
}
