package model

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
	Api_key string
	Url     string
	Model   string
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 试卷大题配置
type PaperSection struct {
	Name      string `json:"name"`       // 大题名称，如"一、单选题"
	Type      int    `json:"type"`       // 题目类型 1:单选 2:多选 3:编程
	Count     int    `json:"count"`      // 题目数量
	ScoreEach int    `json:"score_each"` // 每题分数
}

// 创建试卷的请求体
type CreatePaperRequest struct {
	Title       string         `json:"title" validate:"required"`              // 试卷标题
	Sections    []PaperSection `json:"sections" validate:"required,min=1"`     // 大题配置
	QuestionIds []int          `json:"question_ids" validate:"required,min=1"` // 选中的题目ID列表
}

// 试卷
type Paper struct {
	Id        int            `json:"id"`         // 试卷ID
	Title     string         `json:"title"`      // 试卷标题
	Config    []PaperSection `json:"config"`     // 大题配置（JSON格式存储）
	CreatedAt string         `json:"created_at"` // 创建时间
}

// 试卷-题目关联
type PaperQuestion struct {
	Id           int `json:"id"`             // 主键
	PaperId      int `json:"paper_id"`       // 试卷ID
	QuestionId   int `json:"question_id"`    // 题目ID
	Score        int `json:"score"`          // 这道题的分数
	OrderInPaper int `json:"order_in_paper"` // 题目在试卷中的顺序
}

// 试卷详情（包含完整题目信息）
type PaperDetail struct {
	Paper     Paper                    `json:"paper"`     // 试卷基本信息
	Questions []PaperQuestionWithTopic `json:"questions"` // 题目列表（带完整题目信息）
}

// 试卷题目（包含完整题目信息）
type PaperQuestionWithTopic struct {
	PaperQuestion       // 继承试卷-题目关联信息
	Topic         Topic `json:"topic"` // 完整的题目信息
}

var Choices [4]string = [4]string{"", "单选题", "多选题", "编程题"}
var Difficulties [4]string = [4]string{"", "简单", "中等", "困难"}
