package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"study/internal/config"
	"study/internal/model"
)

// CourseworkService 试题/讲义生成服务
type CourseworkService struct {
	cfg       *config.Config
	pythonBin string // "python" 或 "python3"
}

// NewCourseworkService 创建生成服务
func NewCourseworkService(cfg *config.Config) *CourseworkService {
	svc := &CourseworkService{cfg: cfg}
	svc.pythonBin = svc.findPython()
	return svc
}

// findPython 查找可用的 Python 解释器
func (s *CourseworkService) findPython() string {
	for _, bin := range []string{"python", "python3"} {
		if _, err := exec.LookPath(bin); err == nil {
			return bin
		}
	}
	return "" // 未找到，后续 CheckEnv 会报告
}

// IsAvailable 检查 Python 是否可用
func (s *CourseworkService) IsAvailable() bool {
	return s.pythonBin != ""
}

// CheckEnv 检查运行环境（Python、XeLaTeX/Tectonic）
// 返回问题列表，空列表表示一切就绪
func (s *CourseworkService) CheckEnv() []string {
	var issues []string

	if s.pythonBin == "" {
		issues = append(issues, "未找到 Python 3.10+，请安装: https://www.python.org/downloads/")
	} else {
		// 检查 Python 版本 ≥ 3.10
		out, err := exec.Command(s.pythonBin, "-c", "import sys; print(sys.version_info[:2])").Output()
		if err == nil {
			ver := strings.TrimSpace(string(out))
			if !strings.HasPrefix(ver, "(3, 1") && ver != "(3, 10)" && ver != "(3, 11)" &&
				ver != "(3, 12)" && ver != "(3, 13)" && ver != "(3, 14)" {
				issues = append(issues, fmt.Sprintf("需要 Python 3.10+，当前: %s", ver))
			}
		}
	}

	// 检查 PDF 编译引擎
	hasXeLaTeX := false
	hasTectonic := false
	if _, err := exec.LookPath("xelatex"); err == nil {
		hasXeLaTeX = true
	}
	if _, err := exec.LookPath("tectonic"); err == nil {
		hasTectonic = true
	}
	if !hasXeLaTeX && !hasTectonic {
		issues = append(issues, "未找到 XeLaTeX 或 Tectonic — 无法编译 PDF（可先生成 .tex 文件）")
	}

	return issues
}

// LoadSubjectMeta 扫描约定文件目录，返回所有科目标识
func (s *CourseworkService) LoadSubjectMeta() ([]model.SubjectMeta, error) {
	dir := filepath.Join(s.cfg.PythonProjectDir(), "conventions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取约定目录失败: %w", err)
	}

	var subjects []model.SubjectMeta
	reTitle := regexp.MustCompile(`^#\s*(.+?)(?:\s*约定\s*)?$`)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		key := strings.TrimSuffix(e.Name(), ".md")

		// 读取第一行获取课程中文名
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		name := key // 默认用文件名
		if scanner.Scan() {
			first := strings.TrimSpace(scanner.Text())
			if m := reTitle.FindStringSubmatch(first); m != nil {
				name = strings.TrimSpace(m[1])
			}
		}
		f.Close()

		subjects = append(subjects, model.SubjectMeta{Key: key, Name: name})
	}
	return subjects, nil
}

// ListSections 获取某科目的章节列表（从约定文件解析）
func (s *CourseworkService) ListSections(subject string) ([]string, error) {
	path := filepath.Join(s.cfg.PythonProjectDir(), "conventions", subject+".md")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("未找到科目 %q 的约定文件: %w", subject, err)
	}
	defer f.Close()

	var sections []string
	inBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "<!-- SYLLABUS_START -->" {
			inBlock = true
			continue
		}
		if line == "<!-- SYLLABUS_END -->" {
			break
		}
		if inBlock && line != "" && !strings.HasPrefix(line, ">") {
			sections = append(sections, line)
		}
	}
	return sections, nil
}

// ParseIntent 调用 LLM 解析自然语言 → GenIntent
func (s *CourseworkService) ParseIntent(nl string) (*model.GenIntent, error) {
	apiKey := s.resolveAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("未找到 API key，请设置 STUDY_API_KEY 或 LLM_API_KEY 环境变量")
	}

	// 构建可用科目/章节列表供 LLM 参考
	subjects, err := s.LoadSubjectMeta()
	if err != nil {
		return nil, err
	}
	var subjectList strings.Builder
	for _, m := range subjects {
		subjectList.WriteString(fmt.Sprintf("- %s: %s\n", m.Key, m.Name))
	}

	systemPrompt := fmt.Sprintf(`你是一个意图解析器。从用户输入中提取生成试题/讲义的参数。
输出纯 JSON（不要 markdown 代码块包裹），只输出 JSON。

JSON 字段说明：
- subject: 科目标识（必须从下面列表中选择一个最匹配的）
- mode: "exam"（试题）或 "study"（讲义），默认 "exam"
- sections: 章节名数组，空数组表示"全部章节"
- count: 题目数量（exam 模式），默认 5
- types: 题型数组，如 ["选择题","填空题","计算题"]，空数组表示不限
- difficulty: "基础"/"期末"/"竞赛"，空字符串表示不限
- build: 是否编译 PDF，默认 true
- combined: 试卷和答案是否合订，默认 false
- figures: 是否生成配图，默认 false
- title: 自定义标题（可选）

可用科目列表：
%s`, subjectList.String())

	userPrompt := fmt.Sprintf("用户输入: %s", nl)

	// 调用 LLM
	resp, err := s.callLLM(apiKey, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 清洗响应（去除可能的 markdown 包裹）
	resp = cleanJSONResponse(resp)

	var intent model.GenIntent
	if err := json.Unmarshal([]byte(resp), &intent); err != nil {
		return nil, fmt.Errorf("解析 LLM 返回的 JSON 失败: %w\n原始响应: %s", err, resp)
	}

	intent.RawInput = nl
	return &intent, nil
}

// Run 执行 Python 管线生成 PDF
func (s *CourseworkService) Run(intent *model.GenIntent) (*model.GenResult, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("未找到 Python，请安装 Python 3.10+")
	}

	args := s.buildPythonArgs(intent)
	scriptPath := filepath.Join(s.cfg.PythonProjectDir(), "scripts", "gen_sections.py")

	cmd := exec.Command(s.pythonBin, append([]string{scriptPath}, args...)...)
	// 传递环境变量（API key）
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("⏳ 正在生成 %s...\n", intent.Subject)
	fmt.Printf("   命令: %s %s %s\n", s.pythonBin, scriptPath, strings.Join(args, " "))

	err := cmd.Run()

	// 无论成功与否，都打印输出
	if stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprint(os.Stderr, stderr.String())
	}

	if err != nil {
		return &model.GenResult{
			Success: false,
			Errors:  []string{err.Error(), stderr.String()},
		}, fmt.Errorf("Python 执行失败: %w", err)
	}

	// 从输出中提取产物路径
	outPaths := extractOutputPaths(stdout.String())
	return &model.GenResult{
		Success:  true,
		OutPaths: outPaths,
		Summary:  fmt.Sprintf("%s 生成完成，共 %d 个文件", intent.Subject, len(outPaths)),
	}, nil
}

// ========== 私有方法 ==========

// resolveAPIKey 按优先级读取 API key
// 优先级: STUDY_API_KEY > LLM_API_KEY > OPENAI_API_KEY
func (s *CourseworkService) resolveAPIKey() string {
	for _, key := range []string{"STUDY_API_KEY", "LLM_API_KEY", "OPENAI_API_KEY"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

// resolveBaseURL 按优先级读取 API base URL
func (s *CourseworkService) resolveBaseURL() string {
	for _, key := range []string{"STUDY_BASE_URL", "LLM_BASE_URL", "OPENAI_BASE_URL"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return "https://openrouter.ai/api/v1"
}

// resolveModel 按优先级读取模型名
func (s *CourseworkService) resolveModel() string {
	for _, key := range []string{"STUDY_MODEL", "LLM_MODEL", "OPENAI_MODEL"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return "deepseek/deepseek-v4-pro"
}

// callLLM 调用 OpenAI-compatible API
func (s *CourseworkService) callLLM(apiKey, system, user string) (string, error) {
	baseURL := strings.TrimRight(s.resolveBaseURL(), "/")
	model := s.resolveModel()

	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.3, // 解析用低温
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析 API 响应失败: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API 返回空结果")
	}
	return result.Choices[0].Message.Content, nil
}

// buildPythonArgs 根据 GenIntent 构建 Python 脚本的命令行参数
func (s *CourseworkService) buildPythonArgs(intent *model.GenIntent) []string {
	var args []string

	args = append(args, "--subject", intent.Subject)

	if intent.Mode != "" && intent.Mode != "exam" {
		args = append(args, "--mode", intent.Mode)
	}

	if len(intent.Sections) > 0 {
		args = append(args, "--sections", strings.Join(intent.Sections, ","))
	} else {
		args = append(args, "--sections", "all")
	}

	if intent.Count > 0 {
		args = append(args, "--count", fmt.Sprintf("%d", intent.Count))
	}

	if len(intent.Types) > 0 {
		// 题型: ["选择题","填空题"] → "选择:1,填空:1"
		parts := make([]string, len(intent.Types))
		for i, t := range intent.Types {
			parts[i] = fmt.Sprintf("%s:1", t)
		}
		args = append(args, "--types", strings.Join(parts, ","))
	}

	if intent.Difficulty != "" {
		args = append(args, "--difficulty", intent.Difficulty)
	}

	if intent.Combined {
		args = append(args, "--combined")
	}

	if intent.Figures {
		args = append(args, "--figures")
	}

	if intent.Title != "" {
		args = append(args, "--title", intent.Title)
	}

	// 输出路径
	outDir := s.cfg.GenOutputDir()
	outName := fmt.Sprintf("%s_%s", intent.Subject, time.Now().Format("2006-01-02_150405"))
	args = append(args, "--out", filepath.Join(outDir, outName))

	if intent.Build {
		args = append(args, "--build")
	}

	return args
}

// cleanJSONResponse 去除 LLM 响应中可能的 markdown 代码块包裹
func cleanJSONResponse(s string) string {
	s = strings.TrimSpace(s)
	// 去掉 ```json ... ``` 包裹
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

// extractOutputPaths 从 Python stdout 中提取产物路径
func extractOutputPaths(stdout string) []string {
	var paths []string
	re := regexp.MustCompile(`(?:PDF|TEX|输出)[:：]\s*(.+\.(?:pdf|tex))`)
	for _, m := range re.FindAllStringSubmatch(stdout, -1) {
		if len(m) > 1 {
			paths = append(paths, strings.TrimSpace(m[1]))
		}
	}
	return paths
}
