package markdown

import (
	"bufio"
	"os"
	"strings"
	"time"

	"study/internal/model"
)

// LoadMemos 加载行政备忘
func LoadMemos(path string) ([]model.Memo, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var memos []model.Memo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过标题和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 格式: "- 2026-07-16 内容"
		content := strings.TrimPrefix(line, "- ")
		// 提取日期前缀
		parts := strings.SplitN(content, " ", 2)
		if len(parts) >= 2 {
			dateStr := parts[0]
			text := parts[1]
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil && text != "" {
				memos = append(memos, model.Memo{
					Content:   text,
					CreatedAt: t,
				})
			}
		}
	}
	return memos, scanner.Err()
}

// SaveMemos 保存行政备忘
func SaveMemos(path string, memos []model.Memo) error {
	var sb strings.Builder
	sb.WriteString("# 行政备忘\n\n")
	for _, m := range memos {
		sb.WriteString("- ")
		sb.WriteString(m.CreatedAt.Format("2006-01-02"))
		sb.WriteString(" ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
