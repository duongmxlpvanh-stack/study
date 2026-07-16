package config

import (
	"os"
	"path/filepath"
)

// Google OAuth 相关常量
const (
	// GoogleOAuthRedirectFmt OAuth2 回调地址模板（%d 为端口号）
	GoogleOAuthRedirectFmt = "http://127.0.0.1:%d/callback"
)

// GoogleScopes 返回所有 Google 服务所需的 OAuth2 权限范围
func GoogleScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/drive.file",       // Drive: 仅访问本应用创建的文件
		"https://www.googleapis.com/auth/calendar.events",  // Calendar: 读写事件
	}
}

// Config 应用配置
type Config struct {
	DataDir string // 数据根目录
}

// DefaultDataDir 返回默认数据目录
// Windows: %USERPROFILE%\.study\
func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 回退：使用当前目录下的 data/
		return "data"
	}
	return filepath.Join(home, ".study")
}

// Load 加载配置
// 优先使用环境变量 STUDY_DATA_DIR，否则使用默认路径
func Load() *Config {
	cfg := &Config{
		DataDir: DefaultDataDir(),
	}
	if dir := os.Getenv("STUDY_DATA_DIR"); dir != "" {
		cfg.DataDir = dir
	}
	return cfg
}

// EnsureDirs 确保所有数据子目录存在
func (c *Config) EnsureDirs() error {
	dirs := []string{
		c.DataDir,
		filepath.Join(c.DataDir, "records"),
		filepath.Join(c.DataDir, "materials"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// RecordsDir 学习记录目录
func (c *Config) RecordsDir() string {
	return filepath.Join(c.DataDir, "records")
}

// MaterialsDir 资料根目录
func (c *Config) MaterialsDir() string {
	return filepath.Join(c.DataDir, "materials")
}

// SubjectMaterialsDir 某科目的资料目录
func (c *Config) SubjectMaterialsDir(subject string) string {
	return filepath.Join(c.DataDir, "materials", subject)
}

// DiaryDBPath 日记数据库路径
func (c *Config) DiaryDBPath() string {
	return filepath.Join(c.DataDir, "diary.db")
}

// SubjectsPath 科目文件路径
func (c *Config) SubjectsPath() string {
	return filepath.Join(c.DataDir, "subjects.md")
}

// ExamsPath 考试文件路径
func (c *Config) ExamsPath() string {
	return filepath.Join(c.DataDir, "exams.md")
}

// WeakPointsPath 薄弱知识点文件路径
func (c *Config) WeakPointsPath() string {
	return filepath.Join(c.DataDir, "weakpoints.md")
}

// MemosPath 备忘文件路径
func (c *Config) MemosPath() string {
	return filepath.Join(c.DataDir, "memos.md")
}
