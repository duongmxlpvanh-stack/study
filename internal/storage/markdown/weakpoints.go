package markdown

import (
	"os"
	"strings"
	"time"

	"study/internal/model"
)

// LoadWeakPoints 加载薄弱知识点
func LoadWeakPoints(path string) ([]model.WeakPoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var wps []model.WeakPoint
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "| 内容") || strings.HasPrefix(line, "|------") {
			continue
		}
		if strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				content := strings.TrimSpace(parts[1])
				urgencyStr := strings.TrimSpace(parts[2])
				subject := strings.TrimSpace(parts[3])

				var urgency model.Urgency
				switch urgencyStr {
				case "紧急":
					urgency = model.UrgencyUrgent
				case "不急":
					urgency = model.UrgencyRelaxed
				case "考前看":
					urgency = model.UrgencyPreExam
				default:
					continue // 无效的紧急程度，跳过
				}

				if content != "" {
					wps = append(wps, model.WeakPoint{
						Content:   content,
						Urgency:   urgency,
						Subject:   subject,
						CreatedAt: time.Now(),
					})
				}
			}
		}
	}
	return wps, nil
}

// SaveWeakPoints 保存薄弱知识点
func SaveWeakPoints(path string, wps []model.WeakPoint) error {
	var sb strings.Builder
	sb.WriteString("# 薄弱知识点\n\n")
	sb.WriteString("| 内容 | 紧急程度 | 科目 |\n")
	sb.WriteString("|------|---------|------|\n")
	for _, w := range wps {
		sb.WriteString("| ")
		sb.WriteString(w.Content)
		sb.WriteString(" | ")
		sb.WriteString(string(w.Urgency))
		sb.WriteString(" | ")
		sb.WriteString(w.Subject)
		sb.WriteString(" |\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
