package markdown

import (
	"fmt"
	"os"
	"strings"
	"time"

	"study/internal/model"
)

// GetRecordsFile 获取当月记录文件路径
func GetRecordsFile(recordsDir string) string {
	month := time.Now().Format("2006-01")
	return recordsDir + "/" + month + ".md"
}

// AppendRecord 追加一条学习记录
func AppendRecord(recordsDir string, r model.Record) error {
	filePath := GetRecordsFile(recordsDir)
	// 确保目录存在
	if err := os.MkdirAll(recordsDir, 0755); err != nil {
		return err
	}

	// 检查文件是否存在，不存在则写入表头
	_, err := os.Stat(filePath)
	isNew := os.IsNotExist(err)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if isNew {
		// 写入文件头
		now := time.Now()
		header := fmt.Sprintf("# %s 学习记录\n\n| 日期 | 科目 | 内容 |\n|------|------|------|\n",
			now.Format("2006年01月"))
		if _, err := f.WriteString(header); err != nil {
			return err
		}
	}

	line := fmt.Sprintf("| %s | %s | %s |\n", r.Date, r.Subject, r.Content)
	_, err = f.WriteString(line)
	return err
}

// LoadRecords 加载当月所有记录
func LoadRecords(recordsDir string) ([]model.Record, error) {
	filePath := GetRecordsFile(recordsDir)
	return loadRecordsFromFile(filePath)
}

// LoadRecordsByMonth 加载指定月份的记录
func LoadRecordsByMonth(recordsDir, yearMonth string) ([]model.Record, error) {
	filePath := recordsDir + "/" + yearMonth + ".md"
	return loadRecordsFromFile(filePath)
}

func loadRecordsFromFile(filePath string) ([]model.Record, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var records []model.Record
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过标题行、表头行、分隔行、空行
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "| 日期") || strings.HasPrefix(line, "|------") {
			continue
		}
		if strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				date := strings.TrimSpace(parts[1])
				subject := strings.TrimSpace(parts[2])
				content := strings.TrimSpace(parts[3])
				if date != "" && content != "" {
					records = append(records, model.Record{
						Date:    date,
						Subject: subject,
						Content: content,
					})
				}
			}
		}
	}
	return records, nil
}

// LoadAllRecords 加载所有历史记录（遍历所有月份文件）
func LoadAllRecords(recordsDir string) ([]model.Record, error) {
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var all []model.Record
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		recs, err := loadRecordsFromFile(recordsDir + "/" + e.Name())
		if err != nil {
			continue // 跳过损坏的文件
		}
		all = append(all, recs...)
	}
	return all, nil
}
