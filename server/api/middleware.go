package api

import (
	"system_api/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 请求体字段验证错误处理
func HandleAskError(c *gin.Context, err error) {
	// 处理验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Field() {
			case "Number":
				c.JSON(-1, model.Response{Code: -101, Msg: "number不能为空，至少为1", Data: nil})
			case "Language":
				c.JSON(-1, model.Response{Code: -102, Msg: "language参数错误，language必须是go/javascript/java/python/c++", Data: nil})
			case "Type":
				c.JSON(-1, model.Response{Code: -103, Msg: "type参数错误，type必须是1/2/3", Data: nil})
			case "Keyword":
				c.JSON(-1, model.Response{Code: -104, Msg: "keyword不能为空", Data: nil})
			case "Difficulty":
				c.JSON(-1, model.Response{Code: -105, Msg: "Difficulty参数错误，Difficulty必须是1/2/3", Data: nil})
			}
		}
	} else {
		c.JSON(-1, model.Response{Code: -400, Msg: "请求格式错误", Data: nil})
	}
}
func HandleTopicError(c *gin.Context, err error) {
	// 处理验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Field() {
			case "Title":
				c.JSON(-1, model.Response{Code: -101, Msg: "title不能为空", Data: nil})
			case "Language":
				c.JSON(-1, model.Response{Code: -102, Msg: "language参数错误，language必须是go/javascript/java/python/c++", Data: nil})
			case "Typeid":
				c.JSON(-1, model.Response{Code: -103, Msg: "type不能为空", Data: nil})
			case "Keyword":
				c.JSON(-1, model.Response{Code: -104, Msg: "keyword不能为空", Data: nil})
			case "Difficulty":
				c.JSON(-1, model.Response{Code: -105, Msg: "Difficulty不能为空", Data: nil})
			}
		}
	} else {
		c.JSON(-1, model.Response{Code: -400, Msg: "请求格式错误", Data: nil})
	}
}
