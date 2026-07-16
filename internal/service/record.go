package service

import (
	"fmt"
	"strings"
	"time"

	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/markdown"
)

// RecordService 学习记录服务
type RecordService struct {
	cfg *config.Config
}

func NewRecordService(cfg *config.Config) *RecordService {
	return &RecordService{cfg: cfg}
}

// Log 记录一条学习记录
// 输入格式: "科目: 内容" 或 "科目 内容"
func (s *RecordService) Log(input string) error {
	subject, content := parseLogInput(input)
	if content == "" {
		return fmt.Errorf("请使用格式: study log \"科目: 内容\"")
	}

	r := model.Record{
		Date:    time.Now().Format("2006-01-02"),
		Subject: subject,
		Content: content,
		Time:    time.Now(),
	}
	return markdown.AppendRecord(s.cfg.RecordsDir(), r)
}

// ListRecent 列出最近的记录
func (s *RecordService) ListRecent(limit int) ([]model.Record, error) {
	all, err := markdown.LoadAllRecords(s.cfg.RecordsDir())
	if err != nil {
		return nil, err
	}
	// 记录已按时间倒序（先写入最新），直接截取
	if limit > 0 && len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
}

// ListBySubject 按科目筛选记录
func (s *RecordService) ListBySubject(subject string) ([]model.Record, error) {
	all, err := markdown.LoadAllRecords(s.cfg.RecordsDir())
	if err != nil {
		return nil, err
	}
	var filtered []model.Record
	for _, r := range all {
		if strings.Contains(r.Subject, subject) {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// GetAllRecords 获取所有历史记录（用于统计）
func (s *RecordService) GetAllRecords() ([]model.Record, error) {
	return markdown.LoadAllRecords(s.cfg.RecordsDir())
}

// parseLogInput 解析输入
// "高数: 完成多元函数微分" → ("高数", "完成多元函数微分")
// "完成多元函数微分" → ("", "完成多元函数微分")
func parseLogInput(input string) (subject, content string) {
	// 尝试用 ":" 分割
	if idx := strings.Index(input, ":"); idx != -1 {
		subject = strings.TrimSpace(input[:idx])
		content = strings.TrimSpace(input[idx+1:])
		return
	}
	// 尝试用空格分割第一个词作为科目
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 2 {
		subject = parts[0]
		content = parts[1]
		return
	}
	// 整个作为内容
	content = input
	return
}
