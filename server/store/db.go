package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"system_api/model"
)

// 初始化数据库
func InitDb(db *sql.DB) error {
	// 创建题目表
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
		return fmt.Errorf("创建题目表错误: %w", err)
	}

	// 创建试卷表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS papers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			config JSON NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("创建试卷表错误: %w", err)
	}

	// 创建试卷-题目关联表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS paper_questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			paper_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			score INTEGER NOT NULL,
			order_in_paper INTEGER NOT NULL,
			FOREIGN KEY (paper_id) REFERENCES papers(id) ON DELETE CASCADE,
			FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("创建试卷-题目关联表错误: %w", err)
	}

	return nil
}

// 保存到数据库
func SaveToDB(db *sql.DB, topics []model.Topic) error {
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
func UpdateToDB(db *sql.DB, topic model.Topic) error {
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

// 创建试卷到数据库
func CreatePaperToDB(db *sql.DB, request model.CreatePaperRequest) (*model.Paper, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败：%w", err)
	}
	defer tx.Rollback()

	// 将配置转换为 JSON 字符串
	configJSON, err := json.Marshal(request.Sections)
	if err != nil {
		return nil, fmt.Errorf("序列化试卷配置失败：%w", err)
	}

	// 插入试卷基本信息
	result, err := tx.Exec(`
		INSERT INTO papers (title, config) 
		VALUES (?, ?)
	`, request.Title, string(configJSON))
	if err != nil {
		return nil, fmt.Errorf("插入试卷失败：%w", err)
	}

	paperId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取试卷ID失败：%w", err)
	}

	// 按照配置的顺序插入试卷-题目关联
	order := 1
	questionIndex := 0

	for _, section := range request.Sections {
		// 为每个大题插入对应数量的题目
		for i := 0; i < section.Count; i++ {
			if questionIndex >= len(request.QuestionIds) {
				return nil, fmt.Errorf("题目数量不足，配置需要 %d 道题，但只提供了 %d 道",
					getTotalQuestionCount(request.Sections), len(request.QuestionIds))
			}

			_, err = tx.Exec(`
				INSERT INTO paper_questions (paper_id, question_id, score, order_in_paper)
				VALUES (?, ?, ?, ?)
			`, paperId, request.QuestionIds[questionIndex], section.ScoreEach, order)
			if err != nil {
				return nil, fmt.Errorf("插入试卷题目关联失败：%w", err)
			}

			order++
			questionIndex++
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败：%w", err)
	}

	// 查询并返回创建的试卷信息
	paper := &model.Paper{
		Id:     int(paperId),
		Title:  request.Title,
		Config: request.Sections,
	}

	// 查询创建时间
	err = db.QueryRow("SELECT created_at FROM papers WHERE id = ?", paperId).Scan(&paper.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("查询试卷创建时间失败：%w", err)
	}

	return paper, nil
}

// 根据ID获取试卷详情
func GetPaperDetailFromDB(db *sql.DB, paperId int) (*model.PaperDetail, error) {
	// 查询试卷基本信息
	var paper model.Paper
	var configJSON string

	err := db.QueryRow(`
		SELECT id, title, config, created_at 
		FROM papers 
		WHERE id = ?
	`, paperId).Scan(&paper.Id, &paper.Title, &configJSON, &paper.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("试卷不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询试卷失败：%w", err)
	}

	// 解析配置JSON
	err = json.Unmarshal([]byte(configJSON), &paper.Config)
	if err != nil {
		return nil, fmt.Errorf("解析试卷配置失败：%w", err)
	}

	// 查询试卷包含的题目（按顺序）
	rows, err := db.Query(`
		SELECT pq.id, pq.paper_id, pq.question_id, pq.score, pq.order_in_paper,
		       q.title, q.answers, q.right, q.type_id, q.difficulty, 
		       q.is_ai, q.language, q.keyword, q.active
		FROM paper_questions pq
		JOIN questions q ON pq.question_id = q.id
		WHERE pq.paper_id = ?
		ORDER BY pq.order_in_paper
	`, paperId)

	if err != nil {
		return nil, fmt.Errorf("查询试卷题目失败：%w", err)
	}
	defer rows.Close()

	var questions []model.PaperQuestionWithTopic
	for rows.Next() {
		var pq model.PaperQuestionWithTopic
		var answersJSON, rightJSON string

		err := rows.Scan(
			&pq.Id, &pq.PaperId, &pq.QuestionId, &pq.Score, &pq.OrderInPaper,
			&pq.Topic.Title, &answersJSON, &rightJSON, &pq.Topic.Typeid,
			&pq.Topic.Difficulty, &pq.Topic.Isai, &pq.Topic.Language,
			&pq.Topic.Keyword, &pq.Topic.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描题目数据失败：%w", err)
		}

		// 设置题目ID
		pq.Topic.Id = pq.QuestionId

		// 解析答案和正确选项的JSON
		err = json.Unmarshal([]byte(answersJSON), &pq.Topic.Answers)
		if err != nil {
			return nil, fmt.Errorf("解析题目选项失败：%w", err)
		}

		err = json.Unmarshal([]byte(rightJSON), &pq.Topic.Right)
		if err != nil {
			return nil, fmt.Errorf("解析正确答案失败：%w", err)
		}

		questions = append(questions, pq)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历题目数据失败：%w", err)
	}

	return &model.PaperDetail{
		Paper:     paper,
		Questions: questions,
	}, nil
}

// 获取试卷列表
func GetPapersFromDB(db *sql.DB) ([]model.Paper, error) {
	rows, err := db.Query(`
		SELECT id, title, config, created_at 
		FROM papers 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询试卷列表失败：%w", err)
	}
	defer rows.Close()

	var papers []model.Paper
	for rows.Next() {
		var paper model.Paper
		var configJSON string

		err := rows.Scan(&paper.Id, &paper.Title, &configJSON, &paper.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描试卷数据失败：%w", err)
		}

		// 解析配置JSON
		err = json.Unmarshal([]byte(configJSON), &paper.Config)
		if err != nil {
			return nil, fmt.Errorf("解析试卷配置失败：%w", err)
		}

		papers = append(papers, paper)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历试卷数据失败：%w", err)
	}

	return papers, nil
}

// 辅助函数：计算总题目数量
func getTotalQuestionCount(sections []model.PaperSection) int {
	total := 0
	for _, section := range sections {
		total += section.Count
	}
	return total
}

// 删除试卷
func DeletePaperFromDB(db *sql.DB, paperId int) error {
	// 检查试卷是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM papers WHERE id = ?", paperId).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询试卷失败：%w", err)
	}
	if count == 0 {
		return fmt.Errorf("试卷不存在")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败：%w", err)
	}
	defer tx.Rollback()

	// 删除试卷-题目关联
	_, err = tx.Exec("DELETE FROM paper_questions WHERE paper_id = ?", paperId)
	if err != nil {
		return fmt.Errorf("删除试卷题目关联失败：%w", err)
	}

	// 删除试卷
	_, err = tx.Exec("DELETE FROM papers WHERE id = ?", paperId)
	if err != nil {
		return fmt.Errorf("删除试卷失败：%w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败：%w", err)
	}

	return nil
}

// 更新试卷标题
func UpdatePaperTitleInDB(db *sql.DB, paperId int, title string) error {
	// 检查试卷是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM papers WHERE id = ?", paperId).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询试卷失败：%w", err)
	}
	if count == 0 {
		return fmt.Errorf("试卷不存在")
	}

	// 更新试卷标题
	_, err = db.Exec("UPDATE papers SET title = ? WHERE id = ?", title, paperId)
	if err != nil {
		return fmt.Errorf("更新试卷标题失败：%w", err)
	}

	return nil
}

// 更新试卷题目
func UpdatePaperQuestionsInDB(db *sql.DB, paperId int, title string, questionIds []int) error {
	// 检查试卷是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM papers WHERE id = ?", paperId).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询试卷失败：%w", err)
	}
	if count == 0 {
		return fmt.Errorf("试卷不存在")
	}

	// 获取原试卷信息以保留配置
	var originalTitle, configJSON string
	err = db.QueryRow("SELECT title, config FROM papers WHERE id = ?", paperId).Scan(&originalTitle, &configJSON)
	if err != nil {
		return fmt.Errorf("查询原试卷信息失败：%w", err)
	}

	// 解析原配置
	var config []model.PaperSection
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		return fmt.Errorf("解析试卷配置失败：%w", err)
	}

	// 验证题目数量是否匹配配置
	totalRequired := getTotalQuestionCount(config)
	if len(questionIds) != totalRequired {
		return fmt.Errorf("题目数量不匹配，配置需要 %d 道题，但提供了 %d 道", totalRequired, len(questionIds))
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败：%w", err)
	}
	defer tx.Rollback()

	// 更新试卷标题（如果提供了新标题）
	finalTitle := originalTitle
	if title != "" {
		finalTitle = title
		_, err = tx.Exec("UPDATE papers SET title = ? WHERE id = ?", finalTitle, paperId)
		if err != nil {
			return fmt.Errorf("更新试卷标题失败：%w", err)
		}
	}

	// 删除原有的试卷-题目关联
	_, err = tx.Exec("DELETE FROM paper_questions WHERE paper_id = ?", paperId)
	if err != nil {
		return fmt.Errorf("删除原试卷题目关联失败：%w", err)
	}

	// 重新插入试卷-题目关联
	order := 1
	questionIndex := 0

	for _, section := range config {
		// 为每个大题插入对应数量的题目
		for i := 0; i < section.Count; i++ {
			if questionIndex >= len(questionIds) {
				return fmt.Errorf("题目数量不足")
			}

			_, err = tx.Exec(`
				INSERT INTO paper_questions (paper_id, question_id, score, order_in_paper)
				VALUES (?, ?, ?, ?)
			`, paperId, questionIds[questionIndex], section.ScoreEach, order)
			if err != nil {
				return fmt.Errorf("插入试卷题目关联失败：%w", err)
			}

			order++
			questionIndex++
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败：%w", err)
	}

	return nil
}
