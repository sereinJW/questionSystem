package service

import (
	"fmt"
	"strings"
	"system_api/model"
	"time"
)

// 导出试卷为Word文档
func ExportPaperToWord(paperDetail *model.PaperDetail) ([]byte, error) {
	// 计算总分
	totalScore := 0
	for _, section := range paperDetail.Paper.Config {
		totalScore += section.Count * section.ScoreEach
	}

	// 格式化创建时间
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", paperDetail.Paper.CreatedAt)

	var content strings.Builder

	// 试卷标题和基本信息
	content.WriteString(fmt.Sprintf("试卷标题: %s\n", paperDetail.Paper.Title))
	content.WriteString(fmt.Sprintf("试卷ID: %d\n", paperDetail.Paper.Id))
	content.WriteString(fmt.Sprintf("创建时间: %s\n", createdAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("总题数: %d 题\n", len(paperDetail.Questions)))
	content.WriteString(fmt.Sprintf("总分: %d 分\n", totalScore))
	content.WriteString("\n")

	// 题型分布
	content.WriteString("题型分布:\n")
	for _, section := range paperDetail.Paper.Config {
		content.WriteString(fmt.Sprintf("  %s: %d题 × %d分 = %d分\n",
			section.Name, section.Count, section.ScoreEach, section.Count*section.ScoreEach))
	}
	content.WriteString("\n")
	content.WriteString("==========================================\n\n")

	// 按大题组织题目
	questionIndex := 0

	for _, section := range paperDetail.Paper.Config {
		// 大题标题
		content.WriteString(fmt.Sprintf("%s (每题 %d 分，共 %d 题，计 %d 分)\n",
			section.Name, section.ScoreEach, section.Count, section.Count*section.ScoreEach))
		content.WriteString("\n")

		// 该大题的题目
		for i := 0; i < section.Count && questionIndex < len(paperDetail.Questions); i++ {
			question := paperDetail.Questions[questionIndex]

			// 题目标题
			content.WriteString(fmt.Sprintf("%d. %s\n", question.OrderInPaper, question.Topic.Title))

			// 选项（非编程题）
			if question.Topic.Typeid != 3 && len(question.Topic.Answers) > 0 {
				for _, answer := range question.Topic.Answers {
					content.WriteString(fmt.Sprintf("   %s\n", answer))
				}

				// 正确答案
				if len(question.Topic.Right) > 0 {
					content.WriteString(fmt.Sprintf("   正确答案: %s\n", strings.Join(question.Topic.Right, ", ")))
				}
			}

			// 编程题提示
			if question.Topic.Typeid == 3 {
				content.WriteString("   这是一道编程题，请提供完整的代码实现。\n")
			}

			content.WriteString("\n")
			questionIndex++
		}

		content.WriteString("\n")
	}

	// 页脚
	content.WriteString("==========================================\n")
	content.WriteString("本试卷由考试出题系统自动生成 | wust ontheway\n")

	return []byte(content.String()), nil
}
