package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"study/internal/config"

	"github.com/zalando/go-keyring"
)

const (
	gitKeyringService = "study-github-sync"
	gitKeyringUser    = "github-token"

	// gitignoreContent 云端同步时自动生成的 .gitignore
	gitignoreContent = `# SQLite 日记数据库（二进制，不适合 Git）
diary.db

# 资料文件（大文件，用户自行管理）
materials/

# 生成文件（本地重新生成即可）
gen/

# 同步配置文件（每台机器独立）
.sync_config

# 系统文件
.DS_Store
Thumbs.db
`
)

// SyncStatus 云端同步状态
type SyncStatus struct {
	Configured     bool   // 是否已配置远程仓库
	RemoteURL      string // 远程仓库地址（脱敏，隐藏 token）
	LastSync       string // 上次同步时间
	PendingChanges int    // 待提交的变更数
	HasGit         bool   // 系统是否安装了 Git
}

// GitSyncService 基于 Git 的云端同步服务
type GitSyncService struct {
	cfg     *config.Config
	enabled bool // Git 可用且仓库已配置
	mu      sync.Mutex
}

// NewGitSyncService 创建同步服务
func NewGitSyncService(cfg *config.Config) *GitSyncService {
	svc := &GitSyncService{cfg: cfg}

	// 检查系统是否安装了 Git
	if _, err := exec.LookPath("git"); err != nil {
		svc.enabled = false
		return svc
	}

	// 检查数据目录是否已经是一个 Git 仓库
	gitDir := filepath.Join(cfg.DataDir, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		svc.enabled = false
		return svc
	}

	// 检查是否配置了远程仓库
	remoteURL, err := svc.gitOutput("config", "--get", "remote.origin.url")
	if err != nil || remoteURL == "" {
		svc.enabled = false
		return svc
	}

	svc.enabled = true
	return svc
}

// IsEnabled 返回同步功能是否可用
func (s *GitSyncService) IsEnabled() bool {
	return s.enabled
}

// Setup 初始化 Git 仓库并配置远程
// remoteURL: 如 "https://github.com/user/study-data"
// token: GitHub Personal Access Token
func (s *GitSyncService) Setup(remoteURL, token string) error {
	// 1. 检查 Git 是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("未检测到 Git，请先安装 Git for Windows: https://git-scm.com/download/win")
	}

	// 2. 验证 URL 格式
	if !strings.HasPrefix(remoteURL, "https://github.com/") {
		return fmt.Errorf("仓库地址格式错误，应为 https://github.com/<用户名>/<仓库名>")
	}

	// 3. 如果已经是一个 Git 仓库，先检查
	gitDir := filepath.Join(s.cfg.DataDir, ".git")
	if info, _ := os.Stat(gitDir); info != nil && info.IsDir() {
		return fmt.Errorf("数据目录已经是 Git 仓库。如需重新配置，请手动删除 %s 后重试", gitDir)
	}

	// 4. git init
	if err := s.runGit("init"); err != nil {
		return fmt.Errorf("git init 失败: %w", err)
	}

	// 5. 写入 .gitignore
	giPath := filepath.Join(s.cfg.DataDir, ".gitignore")
	if err := os.WriteFile(giPath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("写入 .gitignore 失败: %w", err)
	}

	// 6. 保存 token 到 Windows 凭据管理器
	if err := keyring.Set(gitKeyringService, gitKeyringUser, token); err != nil {
		return fmt.Errorf("保存凭据失败: %w", err)
	}

	// 7. 配置 Git 用户信息（如果没有全局配置的话）
	name, _ := s.gitOutput("config", "--global", "user.name")
	email, _ := s.gitOutput("config", "--global", "user.email")
	if name == "" {
		s.runGit("config", "user.name", "study-user")
	}
	if email == "" {
		s.runGit("config", "user.email", "study@local")
	}

	// 8. 添加远程仓库（将 token 嵌入 URL 以自动认证）
	tokenURL := buildTokenURL(remoteURL, token)
	if err := s.runGit("remote", "add", "origin", tokenURL); err != nil {
		return fmt.Errorf("添加远程仓库失败: %w", err)
	}

	// 9. 暂存并提交所有数据
	if err := s.runGit("add", "-A"); err != nil {
		return fmt.Errorf("git add 失败: %w", err)
	}

	// 检查是否有内容可提交
	statusOut, _ := s.gitOutput("status", "--porcelain")
	if statusOut != "" {
		if err := s.runGit("commit", "-m", "初始化: study 学习数据"); err != nil {
			// 提交失败不阻断（可能没有变更）
		}

		// 10. 推送到远程（先尝试 main，再尝试 master）
		branch := s.detectBranch()
		if err := s.runGit("push", "-u", "origin", branch); err != nil {
			return fmt.Errorf("推送失败: %w\n请检查 Token 权限（需要 'repo' 权限）和仓库地址是否正确", err)
		}
	}

	// 11. 保存同步配置
	s.cfg.GitRemote = remoteURL
	s.cfg.SyncEnabled = true
	if err := s.cfg.SaveSyncConfig(); err != nil {
		return fmt.Errorf("保存同步配置失败: %w", err)
	}

	// 12. 启用服务
	s.enabled = true
	return nil
}

// AutoSync 自动同步（后台 goroutine，fire-and-forget）
// msg 作为 commit 信息的一部分
func (s *GitSyncService) AutoSync(msg string) {
	if !s.enabled {
		return
	}
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// 1. 暂存所有变更
		if err := s.runGitIgnoreError("add", "-A"); err != nil {
			return
		}

		// 2. 检查是否有变更（--quiet: exit 0=无变更, exit 1=有变更）
		if err := s.runGit("diff", "--cached", "--quiet"); err == nil {
			return // 无变更，跳过
		}

		// 3. 提交
		commitMsg := fmt.Sprintf("auto: %s", msg)
		if err := s.runGitIgnoreError("commit", "-m", commitMsg); err != nil {
			return
		}

		// 4. 拉取远程变更
		s.runGitIgnoreError("pull", "--rebase")

		// 5. 推送
		if err := s.runGit("push"); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ 云端同步失败，下次操作会自动重试\n")
			// 非致命：不阻断用户操作
		}
	}()
}

// ManualSync 手动同步：先 pull 再 push
func (s *GitSyncService) ManualSync() (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("云端同步未配置。使用 study sync setup 设置")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var result strings.Builder

	// 1. 暂存本地变更
	if err := s.runGitIgnoreError("add", "-A"); err != nil {
		return "", fmt.Errorf("git add 失败: %w", err)
	}

	// 2. 提交（如果有变更）
	if err := s.runGit("diff", "--cached", "--quiet"); err != nil {
		// 有变更，提交
		msg := fmt.Sprintf("手动同步: %s", time.Now().Format("2006-01-02 15:04"))
		if err := s.runGit("commit", "-m", msg); err != nil {
			// 提交失败可能因为无变更，继续
		}
		result.WriteString("已提交本地变更\n")
	}

	// 3. 拉取
	pullOut, err := s.gitOutput("pull", "--rebase")
	if err != nil {
		return "", fmt.Errorf("git pull 失败: %s %w", pullOut, err)
	}
	if pullOut != "" {
		result.WriteString(strings.TrimSpace(pullOut))
		result.WriteString("\n")
	}

	// 4. 推送
	pushOut, err := s.gitOutput("push")
	if err != nil {
		return "", fmt.Errorf("git push 失败: %s %w", pushOut, err)
	}
	result.WriteString("推送成功 ✅")
	result.WriteString("\n")

	return result.String(), nil
}

// Status 返回同步状态
func (s *GitSyncService) Status() *SyncStatus {
	st := &SyncStatus{
		HasGit: exec.Command("git", "--version").Run() == nil,
	}

	if !s.enabled {
		// 即使未启用，也填充已有的仓库地址
		if s.cfg.GitRemote != "" {
			st.Configured = true
			st.RemoteURL = s.cfg.GitRemote
		}
		return st
	}

	st.Configured = true

	// 远程仓库地址（脱敏）
	remoteURL, _ := s.gitOutput("config", "--get", "remote.origin.url")
	st.RemoteURL = maskToken(remoteURL)

	// 上次提交时间
	lastCommit, _ := s.gitOutput("log", "-1", "--format=%ci")
	if lastCommit != "" {
		st.LastSync = strings.TrimSpace(lastCommit)
	}

	// 待提交的变更数
	statusOut, _ := s.gitOutput("status", "--porcelain")
	if statusOut != "" {
		st.PendingChanges = len(strings.Split(strings.TrimSpace(statusOut), "\n"))
	}

	return st
}

// Disable 禁用云端同步（不移除 Git 仓库，仅标记禁用）
func (s *GitSyncService) Disable() error {
	s.enabled = false
	s.cfg.SyncEnabled = false
	return s.cfg.SaveSyncConfig()
}

// ========== 内部辅助方法 ==========

// runGit 执行 git 命令，失败时返回错误
func (s *GitSyncService) runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.cfg.DataDir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// runGitIgnoreError 执行 git 命令，忽略失败
func (s *GitSyncService) runGitIgnoreError(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.cfg.DataDir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.CombinedOutput() // 忽略输出和错误
	return nil
}

// gitOutput 执行 git 命令，返回输出内容
func (s *GitSyncService) gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.cfg.DataDir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// detectBranch 检测当前分支名
func (s *GitSyncService) detectBranch() string {
	branch, err := s.gitOutput("branch", "--show-current")
	if err != nil || branch == "" {
		// 检查是否有 master
		if _, err := s.gitOutput("rev-parse", "master"); err == nil {
			return "master"
		}
		return "main"
	}
	return branch
}

// buildTokenURL 将 token 嵌入仓库 URL
// https://github.com/user/repo → https://<token>@github.com/user/repo
func buildTokenURL(repoURL, token string) string {
	// 去掉可能已有的 https:// 前缀
	clean := strings.TrimPrefix(repoURL, "https://")
	return fmt.Sprintf("https://%s@%s", token, clean)
}

// maskToken 隐藏 URL 中的 token 部分
// https://ghp_xxx@github.com/user/repo → https://github.com/user/repo
func maskToken(url string) string {
	if idx := strings.Index(url, "@"); idx != -1 {
		return "https://" + url[idx+1:]
	}
	return url
}
