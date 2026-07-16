package service

import (
	"os"
	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/markdown"
)

// SubjectService 科目管理服务
type SubjectService struct {
	cfg *config.Config
}

func NewSubjectService(cfg *config.Config) *SubjectService {
	return &SubjectService{cfg: cfg}
}

// List 列出所有科目
func (s *SubjectService) List() ([]model.Subject, error) {
	return markdown.LoadSubjects(s.cfg.SubjectsPath())
}

// Add 添加科目
func (s *SubjectService) Add(name string) error {
	subjects, err := s.List()
	if err != nil {
		return err
	}

	// 检查重复
	for _, sub := range subjects {
		if sub.Name == name {
			return nil // 已存在，不报错
		}
	}

	subjects = append(subjects, model.Subject{Name: name})
	if err := markdown.SaveSubjects(s.cfg.SubjectsPath(), subjects); err != nil {
		return err
	}

	// 同时创建资料文件夹
	materialDir := s.cfg.SubjectMaterialsDir(name)
	return os.MkdirAll(materialDir, 0755)
}

// ListWithMaterialCount 列出科目及资料数量（Dashboard 用）
func (s *SubjectService) ListWithMaterialCount() ([]model.SubjectWithCount, error) {
	subjects, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []model.SubjectWithCount
	for _, sub := range subjects {
		dir := s.cfg.SubjectMaterialsDir(sub.Name)
		count := countFiles(dir)
		result = append(result, model.SubjectWithCount{
			Name:         sub.Name,
			MaterialCount: count,
		})
	}
	return result, nil
}

// countFiles 统计目录中文件数量
func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}
