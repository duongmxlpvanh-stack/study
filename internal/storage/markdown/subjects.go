package markdown

import (
	"bufio"
	"os"
	"strings"
	"time"

	"study/internal/model"
)

// LoadSubjects 从文件加载科目列表
// 格式: # 课程列表\n- 科目名\n- 科目名\n...
func LoadSubjects(path string) ([]model.Subject, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var subjects []model.Subject
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过标题行和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 "- 科目名" 或 "- 科目名 (添加于 2006-01-02)"
		name := strings.TrimPrefix(line, "- ")
		// 提取纯科目名（去掉可能的时间戳后缀）
		if idx := strings.LastIndex(name, " (添加于 "); idx != -1 {
			name = name[:idx]
		}
		name = strings.TrimSpace(name)
		if name != "" {
			subjects = append(subjects, model.Subject{
				Name:      name,
				CreatedAt: time.Now(), // Markdown 不存时间，简化处理
			})
		}
	}
	return subjects, scanner.Err()
}

// SaveSubjects 保存科目列表到文件
func SaveSubjects(path string, subjects []model.Subject) error {
	var sb strings.Builder
	sb.WriteString("# 课程列表\n\n")
	for _, s := range subjects {
		sb.WriteString("- ")
		sb.WriteString(s.Name)
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
