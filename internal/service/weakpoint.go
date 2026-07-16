package service

import (
	"fmt"

	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/markdown"
)

// WeakPointService 薄弱知识点服务
type WeakPointService struct {
	cfg *config.Config
}

func NewWeakPointService(cfg *config.Config) *WeakPointService {
	return &WeakPointService{cfg: cfg}
}

// List 列出所有薄弱点
func (s *WeakPointService) List() ([]model.WeakPoint, error) {
	return markdown.LoadWeakPoints(s.cfg.WeakPointsPath())
}

// Add 添加薄弱知识点
func (s *WeakPointService) Add(content string, urgency model.Urgency, subject string) error {
	wps, err := s.List()
	if err != nil {
		return err
	}
	wps = append(wps, model.WeakPoint{
		Content: content,
		Urgency: urgency,
		Subject: subject,
	})
	return markdown.SaveWeakPoints(s.cfg.WeakPointsPath(), wps)
}

// Delete 删除薄弱点（按序号，从 1 开始）
func (s *WeakPointService) Delete(index int) error {
	wps, err := s.List()
	if err != nil {
		return err
	}
	if index < 1 || index > len(wps) {
		return fmt.Errorf("序号无效，共 %d 条薄弱知识点", len(wps))
	}
	wps = append(wps[:index-1], wps[index:]...)
	return markdown.SaveWeakPoints(s.cfg.WeakPointsPath(), wps)
}

// Stats 统计各紧急程度的数量
func (s *WeakPointService) Stats() (model.WeakPointStats, error) {
	wps, err := s.List()
	if err != nil {
		return model.WeakPointStats{}, err
	}
	var stats model.WeakPointStats
	for _, w := range wps {
		switch w.Urgency {
		case model.UrgencyUrgent:
			stats.Urgent++
		case model.UrgencyRelaxed:
			stats.Relaxed++
		case model.UrgencyPreExam:
			stats.PreExam++
		}
	}
	return stats, nil
}
