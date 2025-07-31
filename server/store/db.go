package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"system_api/model"
)

// 初始化数据库
func InitDb(db *sql.DB) error {
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
