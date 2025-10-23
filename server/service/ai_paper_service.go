package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"system_api/model"
	"system_api/store"
	"time"
)

// AI生成试卷的大题配置
type AIGeneratePaperSection struct {
	Name       string `json:"name" validate:"required"`             // 大题名称
	Type       int    `json:"type" validate:"required"`             // 题型
	Count      int    `json:"count" validate:"required,min=1"`      // 数量
	ScoreEach  int    `json:"score_each" validate:"required,min=1"` // 每题分数
	Difficulty int    `json:"difficulty" validate:"required"`       // 难度
	Language   string `json:"language" validate:"required"`         // 语言
	Keyword    string `json:"keyword" validate:"required"`          // 关键词
}

// AI生成试卷请求
type AIGeneratePaperRequest struct {
	Title    string                   `json:"title" validate:"required"`
	Sections []AIGeneratePaperSection `json:"sections" validate:"required,min=1"`
}

// 生成进度信息
type GenerateProgress struct {
	SectionIndex int    `json:"section_index"` // 当前大题索引
	SectionName  string `json:"section_name"`  // 当前大题名称
	Current      int    `json:"current"`       // 当前已生成数量
	Total        int    `json:"total"`         // 总数量
	Status       string `json:"status"`        // 状态：generating, completed, failed
	Message      string `json:"message"`       // 消息
}

// 生成结果
type GenerateResult struct {
	Success  bool               `json:"success"`
	Paper    *model.Paper       `json:"paper,omitempty"`
	Progress []GenerateProgress `json:"progress"`
	Error    string             `json:"error,omitempty"`
}

// 并发控制：最大并发数
const maxConcurrency = 3

// AI生成试卷（带并发控制和进度反馈）
func AIGeneratePaper(ctx context.Context, db *sql.DB, ai model.Ai, request AIGeneratePaperRequest) (*GenerateResult, error) {
	result := &GenerateResult{
		Success:  false,
		Progress: make([]GenerateProgress, 0),
	}

	// 计算总题目数
	totalQuestions := 0
	for _, section := range request.Sections {
		totalQuestions += section.Count
	}

	// 成本控制：限制单次生成的题目总数
	const maxQuestionsPerPaper = 100
	if totalQuestions > maxQuestionsPerPaper {
		return nil, fmt.Errorf("单次生成题目数量不能超过 %d 道，当前配置需要 %d 道", maxQuestionsPerPaper, totalQuestions)
	}

	// 存储所有生成的题目ID
	allQuestionIds := make([]int, 0, totalQuestions)

	// 用于并发控制的信号量
	semaphore := make(chan struct{}, maxConcurrency)

	// 用于收集错误
	var errMutex sync.Mutex
	var firstError error

	// 用于同步等待所有goroutine完成
	var wg sync.WaitGroup

	// 按大题顺序生成题目
	for sectionIndex, section := range request.Sections {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("生成已取消")
		default:
		}

		progress := GenerateProgress{
			SectionIndex: sectionIndex,
			SectionName:  section.Name,
			Current:      0,
			Total:        section.Count,
			Status:       "generating",
			Message:      fmt.Sprintf("正在生成 %s...", section.Name),
		}

		// 批量生成当前大题的题目（使用并发控制）
		sectionQuestions := make([]model.Topic, 0, section.Count)
		var sectionMutex sync.Mutex

		// 计算批次大小（每批生成的题目数）
		batchSize := 5
		if section.Count < batchSize {
			batchSize = section.Count
		}

		// 分批生成
		for i := 0; i < section.Count; i += batchSize {
			// 检查是否有错误发生
			errMutex.Lock()
			if firstError != nil {
				errMutex.Unlock()
				break
			}
			errMutex.Unlock()

			// 计算当前批次的数量
			currentBatchSize := batchSize
			if i+batchSize > section.Count {
				currentBatchSize = section.Count - i
			}

			wg.Add(1)
			go func(batchIndex, batchCount int) {
				defer wg.Done()

				// 获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// 检查上下文
				select {
				case <-ctx.Done():
					return
				default:
				}

				// 构造AI请求
				ask := model.Ask{
					Number:     batchCount,
					Language:   section.Language,
					Type:       section.Type,
					Difficulty: section.Difficulty,
					Keyword:    section.Keyword,
				}

				// 调用AI生成题目
				topics, err := VisitAi(ai, ask)
				if err != nil {
					errMutex.Lock()
					if firstError == nil {
						firstError = fmt.Errorf("生成 %s 失败: %w", section.Name, err)
					}
					errMutex.Unlock()
					return
				}

				// 添加到当前大题的题目列表
				sectionMutex.Lock()
				sectionQuestions = append(sectionQuestions, topics...)
				progress.Current = len(sectionQuestions)
				sectionMutex.Unlock()

			}(i, currentBatchSize)
		}

		// 等待当前大题的所有批次完成
		wg.Wait()

		// 检查是否有错误
		if firstError != nil {
			progress.Status = "failed"
			progress.Message = firstError.Error()
			result.Progress = append(result.Progress, progress)
			result.Error = firstError.Error()
			return result, firstError
		}

		// 检查生成的题目数量是否正确
		if len(sectionQuestions) != section.Count {
			err := fmt.Errorf("生成题目数量不匹配，期望 %d 道，实际生成 %d 道", section.Count, len(sectionQuestions))
			progress.Status = "failed"
			progress.Message = err.Error()
			result.Progress = append(result.Progress, progress)
			result.Error = err.Error()
			return result, err
		}

		// 保存当前大题的题目到数据库
		err := store.SaveToDB(db, sectionQuestions)
		if err != nil {
			progress.Status = "failed"
			progress.Message = fmt.Sprintf("保存题目失败: %v", err)
			result.Progress = append(result.Progress, progress)
			result.Error = err.Error()
			return result, err
		}

		// 获取保存后的题目ID
		// 查询最近插入的题目ID
		rows, err := db.Query(`
			SELECT id FROM questions 
			WHERE is_ai = 1 
			ORDER BY id DESC 
			LIMIT ?
		`, section.Count)
		if err != nil {
			progress.Status = "failed"
			progress.Message = fmt.Sprintf("查询题目ID失败: %v", err)
			result.Progress = append(result.Progress, progress)
			result.Error = err.Error()
			return result, err
		}

		sectionIds := make([]int, 0, section.Count)
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				progress.Status = "failed"
				progress.Message = fmt.Sprintf("读取题目ID失败: %v", err)
				result.Progress = append(result.Progress, progress)
				result.Error = err.Error()
				return result, err
			}
			sectionIds = append(sectionIds, id)
		}
		rows.Close()

		// 反转ID列表（因为查询是倒序的）
		for i := len(sectionIds) - 1; i >= 0; i-- {
			allQuestionIds = append(allQuestionIds, sectionIds[i])
		}

		// 更新进度
		progress.Status = "completed"
		progress.Message = fmt.Sprintf("%s 生成完成", section.Name)
		result.Progress = append(result.Progress, progress)
	}

	// 构造试卷配置
	paperSections := make([]model.PaperSection, len(request.Sections))
	for i, section := range request.Sections {
		paperSections[i] = model.PaperSection{
			Name:      section.Name,
			Type:      section.Type,
			Count:     section.Count,
			ScoreEach: section.ScoreEach,
		}
	}

	// 创建试卷
	createPaperRequest := model.CreatePaperRequest{
		Title:       request.Title,
		Sections:    paperSections,
		QuestionIds: allQuestionIds,
	}

	paper, err := store.CreatePaperToDB(db, createPaperRequest)
	if err != nil {
		result.Error = fmt.Sprintf("创建试卷失败: %v", err)
		return result, err
	}

	result.Success = true
	result.Paper = paper
	return result, nil
}

// 简化版：快速AI生成试卷（不返回详细进度）
func QuickAIGeneratePaper(db *sql.DB, ai model.Ai, request AIGeneratePaperRequest) (*model.Paper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := AIGeneratePaper(ctx, db, ai, request)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("生成失败: %s", result.Error)
	}

	return result.Paper, nil
}
