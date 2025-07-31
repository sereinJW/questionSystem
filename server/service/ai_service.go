package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"system_api/model"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// 生成对应message
func ToString(ask model.Ask) string {
	if ask.Type == 3 {
		return fmt.Sprintf(`请生成 %d 道难度为 %s 的 %s 语言的关于 %s 知识点的 %s ，并以JSON数组格式输出，每个题目包含以下字段：title(题干)。示例格式：[{"title":"请用GO实现冒泡排序，代码编程实现"}]。只输出JSON数组，不需要其他内容。`, ask.Number, model.Difficulties[ask.Difficulty], ask.Language, ask.Keyword, model.Choices[ask.Type])
	}
	return fmt.Sprintf(`请生成 %d 道难度为 %s 的 %s 语言的关于 %s 知识点的 %s ，并以JSON数组格式输出，每个题目包含以下字段：title(题干),answers(选项数组),right(正确答案数组)。示例格式：[{"title":"Gin框架的作用是什么","answers":["A.Gin是一个用于前端开发的JavaScript框架，类似React或Vue","B.Gin是Go语言的高性能HTTP Web框架，支持路由分组、中间件等功能","C.Gin主要用于数据库操作，是ORM工具的一种","D.Gin是Python的异步Web框架，类似Django或Flask"],"right":["B"]}]。只输出JSON数组，不需要其他内容。`, ask.Number, model.Difficulties[ask.Difficulty], ask.Language, ask.Keyword, model.Choices[ask.Type])
}

// 访问AI并处理
func VisitAi(ai model.Ai, ask model.Ask) ([]model.Topic, error) {
	//根据申请生成对应message
	message := ToString(ask)

	//启动ai访问
	client := openai.NewClient(
		option.WithAPIKey(ai.Api_key),
		option.WithBaseURL(ai.Url),
	)

	//正式访问
	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(message),
			},
			Model: ai.Model,
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
	var topics []model.Topic
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
