package markdown

import (
	"os"
	"strings"
	"time"

	"study/internal/model"
)

// LoadExams 加载考试列表
func LoadExams(path string) ([]model.Exam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var exams []model.Exam
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "| 考试名称") || strings.HasPrefix(line, "|------") {
			continue
		}
		if strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				name := strings.TrimSpace(parts[1])
				dateStr := strings.TrimSpace(parts[2])
				if name != "" && dateStr != "" {
					t, err := time.Parse("2006-01-02", dateStr)
					if err != nil {
						continue
					}
					exams = append(exams, model.Exam{
						Name: name,
						Date: t,
					})
				}
			}
		}
	}
	return exams, nil
}

// SaveExams 保存考试列表
func SaveExams(path string, exams []model.Exam) error {
	var sb strings.Builder
	sb.WriteString("# 考试列表\n\n")
	sb.WriteString("| 考试名称 | 日期 |\n")
	sb.WriteString("|----------|------|\n")
	for _, e := range exams {
		sb.WriteString("| ")
		sb.WriteString(e.Name)
		sb.WriteString(" | ")
		sb.WriteString(e.Date.Format("2006-01-02"))
		sb.WriteString(" |\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
