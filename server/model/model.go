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

var Choices [4]string = [4]string{"", "单选题", "多选题", "编程题"}
var Difficulties [4]string = [4]string{"", "简单", "中等", "困难"}
