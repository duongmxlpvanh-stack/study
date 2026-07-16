package service

import (
	"fmt"
	"math"
	"time"

	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/markdown"
)

// ExamService 考试管理服务
type ExamService struct {
	cfg *config.Config
}

func NewExamService(cfg *config.Config) *ExamService {
	return &ExamService{cfg: cfg}
}

// List 列出所有考试（带倒计时）
func (s *ExamService) List() ([]model.ExamWithCountdown, error) {
	exams, err := markdown.LoadExams(s.cfg.ExamsPath())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var result []model.ExamWithCountdown
	for _, e := range exams {
		examDay := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, e.Date.Location())
		daysLeft := int(math.Ceil(examDay.Sub(today).Hours() / 24))

		ec := model.ExamWithCountdown{
			Name:     e.Name,
			Date:     e.Date.Format("2006-01-02"),
			DaysLeft: daysLeft,
		}

		switch {
		case daysLeft < 0:
			ec.UrgencyStr = "✅ 已结束"
		case daysLeft <= 7:
			ec.UrgencyStr = "🔴 临近"
		case daysLeft <= 30:
			ec.UrgencyStr = "🟡 适中"
		default:
			ec.UrgencyStr = "🟢 充裕"
		}

		result = append(result, ec)
	}
	return result, nil
}

// Add 添加考试
func (s *ExamService) Add(name string, dateStr string) error {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD 格式，例如 2026-07-15")
	}

	exams, err := markdown.LoadExams(s.cfg.ExamsPath())
	if err != nil {
		return err
	}

	exams = append(exams, model.Exam{Name: name, Date: t})
	return markdown.SaveExams(s.cfg.ExamsPath(), exams)
}

// Delete 删除考试（按序号，从 1 开始）
func (s *ExamService) Delete(index int) error {
	exams, err := markdown.LoadExams(s.cfg.ExamsPath())
	if err != nil {
		return err
	}
	if index < 1 || index > len(exams) {
		return fmt.Errorf("序号无效，共 %d 场考试", len(exams))
	}
	exams = append(exams[:index-1], exams[index:]...)
	return markdown.SaveExams(s.cfg.ExamsPath(), exams)
}
