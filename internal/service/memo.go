package service

import (
	"fmt"
	"strings"
	"time"

	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/markdown"
)

// MemoService 行政备忘服务
type MemoService struct {
	cfg *config.Config
}

func NewMemoService(cfg *config.Config) *MemoService {
	return &MemoService{cfg: cfg}
}

// List 列出所有备忘
func (s *MemoService) List() ([]model.Memo, error) {
	return markdown.LoadMemos(s.cfg.MemosPath())
}

// Add 添加备忘
func (s *MemoService) Add(content string) error {
	memos, err := s.List()
	if err != nil {
		return err
	}
	memos = append(memos, model.Memo{
		Content:   content,
		CreatedAt: time.Now(),
	})
	return markdown.SaveMemos(s.cfg.MemosPath(), memos)
}

// Search 搜索备忘
func (s *MemoService) Search(keyword string) ([]model.Memo, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	var result []model.Memo
	for _, m := range all {
		if strings.Contains(m.Content, keyword) {
			result = append(result, m)
		}
	}
	return result, nil
}

// Delete 删除备忘（按序号，从 1 开始）
func (s *MemoService) Delete(index int) error {
	memos, err := s.List()
	if err != nil {
		return err
	}
	if index < 1 || index > len(memos) {
		return fmt.Errorf("序号无效，共 %d 条备忘", len(memos))
	}
	memos = append(memos[:index-1], memos[index:]...)
	return markdown.SaveMemos(s.cfg.MemosPath(), memos)
}
