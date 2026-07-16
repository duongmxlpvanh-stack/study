package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"study/internal/config"
	"study/internal/model"
	"study/internal/storage/sqlite"
)

// DiaryService 日记服务
type DiaryService struct {
	cfg   *config.Config
	store *sqlite.DiaryStore
}

func NewDiaryService(cfg *config.Config) (*DiaryService, error) {
	store, err := sqlite.New(cfg.DiaryDBPath())
	if err != nil {
		return nil, err
	}
	return &DiaryService{cfg: cfg, store: store}, nil
}

// Open 打开指定日期的日记（用外部编辑器）
func (s *DiaryService) Open(date string) error {
	diary, err := s.store.GetDiary(date)
	if err != nil {
		return err
	}

	// 日记不存在，创建空草稿
	if diary == nil {
		if err := s.store.SaveDiary(date, ""); err != nil {
			return err
		}
		diary = &model.Diary{Date: date, Content: ""}
	}

	// 写入临时文件供编辑器打开
	tmpFile := s.cfg.DataDir + "/_diary_tmp.md"
	content := fmt.Sprintf("# 日记 %s\n\n%s", diary.Date, diary.Content)
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 打开编辑器
	editor := s.detectEditor()
	cmd := exec.Command(editor, tmpFile)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打开编辑器失败: %w", err)
	}

	// 读取修改后的内容
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("读取临时文件失败: %w", err)
	}
	newContent := string(data)
	// 去掉自动添加的标题行 "# 日记 YYYY-MM-DD"
	newContent = stripFirstLine(newContent)

	// 清理临时文件
	os.Remove(tmpFile)

	return s.store.SaveDiary(diary.Date, newContent)
}

// Search 全文搜索日记
func (s *DiaryService) Search(keyword string) ([]model.Diary, error) {
	return s.store.SearchDiaries(keyword)
}

// ListRecent 列出最近的日记条目
func (s *DiaryService) ListRecent(limit int) ([]model.Diary, error) {
	return s.store.ListRecentDiaries(limit)
}

// Get 获取指定日期日记
func (s *DiaryService) Get(date string) (*model.Diary, error) {
	return s.store.GetDiary(date)
}

// Delete 删除日记
func (s *DiaryService) Delete(date string) error {
	return s.store.DeleteDiary(date)
}

// Close 关闭数据库
func (s *DiaryService) Close() error {
	return s.store.Close()
}

// detectEditor 检测可用编辑器
func (s *DiaryService) detectEditor() string {
	// 按优先级探测
	candidates := []string{"code", "notepad", "vim", "nano"}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vim"
}

func stripFirstLine(content string) string {
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return strings.TrimSpace(content)
}
