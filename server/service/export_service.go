package service

import (
	"archive/zip"
	"bytes"
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

	// 构建文档内容XML
	var content strings.Builder
	content.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	content.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	content.WriteString(`<w:body>`)

	// 试卷标题
	addParagraph(&content, paperDetail.Paper.Title, true, true, 32)
	addParagraph(&content, "", false, false, 0)

	// 基本信息
	addParagraph(&content, fmt.Sprintf("试卷ID: %d", paperDetail.Paper.Id), false, false, 0)
	addParagraph(&content, fmt.Sprintf("创建时间: %s", createdAt.Format("2006-01-02 15:04:05")), false, false, 0)
	addParagraph(&content, fmt.Sprintf("总题数: %d 题", len(paperDetail.Questions)), false, false, 0)
	addParagraph(&content, fmt.Sprintf("总分: %d 分", totalScore), false, false, 0)
	addParagraph(&content, "", false, false, 0)

	// 题型分布
	addParagraph(&content, "题型分布:", true, false, 0)
	for _, section := range paperDetail.Paper.Config {
		addParagraph(&content, fmt.Sprintf("  %s: %d题 × %d分 = %d分",
			section.Name, section.Count, section.ScoreEach, section.Count*section.ScoreEach), false, false, 0)
	}
	addParagraph(&content, "", false, false, 0)
	addParagraph(&content, "==========================================", false, false, 0)
	addParagraph(&content, "", false, false, 0)

	// 按大题组织题目
	questionIndex := 0

	for _, section := range paperDetail.Paper.Config {
		// 大题标题
		addParagraph(&content, fmt.Sprintf("%s (每题 %d 分，共 %d 题，计 %d 分)",
			section.Name, section.ScoreEach, section.Count, section.Count*section.ScoreEach), true, false, 24)
		addParagraph(&content, "", false, false, 0)

		// 该大题的题目
		for i := 0; i < section.Count && questionIndex < len(paperDetail.Questions); i++ {
			question := paperDetail.Questions[questionIndex]

			// 题目标题
			addParagraph(&content, fmt.Sprintf("%d. %s", question.OrderInPaper, question.Topic.Title), true, false, 0)

			// 选项（非编程题）
			if question.Topic.Typeid != 3 && len(question.Topic.Answers) > 0 {
				for _, answer := range question.Topic.Answers {
					addParagraph(&content, fmt.Sprintf("   %s", answer), false, false, 0)
				}

				// 正确答案
				if len(question.Topic.Right) > 0 {
					addParagraph(&content, fmt.Sprintf("   正确答案: %s", strings.Join(question.Topic.Right, ", ")), false, false, 0)
				}
			}

			// 编程题提示
			if question.Topic.Typeid == 3 {
				addParagraph(&content, "   这是一道编程题，请提供完整的代码实现。", false, false, 0)
			}

			addParagraph(&content, "", false, false, 0)
			questionIndex++
		}

		addParagraph(&content, "", false, false, 0)
	}

	// 页脚
	addParagraph(&content, "==========================================", false, false, 0)
	addParagraph(&content, "本试卷由考试出题系统自动生成 | ontheway", false, true, 0)

	content.WriteString(`</w:body>`)
	content.WriteString(`</w:document>`)

	// 创建docx压缩包
	return createDocxZip(content.String())
}

func addParagraph(content *strings.Builder, text string, bold bool, center bool, size int) {
	content.WriteString(`<w:p>`)

	if center {
		content.WriteString(`<w:pPr><w:jc w:val="center"/></w:pPr>`)
	}

	content.WriteString(`<w:r><w:rPr>`)
	if bold {
		content.WriteString(`<w:b/>`)
	}
	if size > 0 {
		content.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, size))
	}
	content.WriteString(`</w:rPr>`)
	content.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, escapeXML(text)))
	content.WriteString(`</w:r>`)
	content.WriteString(`</w:p>`)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func createDocxZip(documentXML string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 1. [Content_Types].xml
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	if err := writeZipFile(zipWriter, "[Content_Types].xml", contentTypes); err != nil {
		return nil, err
	}

	// 2. _rels/.rels
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if err := writeZipFile(zipWriter, "_rels/.rels", rels); err != nil {
		return nil, err
	}

	// 3. word/document.xml
	if err := writeZipFile(zipWriter, "word/document.xml", documentXML); err != nil {
		return nil, err
	}

	// 关闭zip writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("关闭zip失败: %w", err)
	}

	return buf.Bytes(), nil
}

func writeZipFile(zipWriter *zip.Writer, filename string, content string) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("创建zip文件 %s 失败: %w", filename, err)
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		return fmt.Errorf("写入zip文件 %s 失败: %w", filename, err)
	}
	return nil
}
