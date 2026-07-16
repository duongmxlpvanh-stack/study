package model

import "time"

// Diary 学习日记
type Diary struct {
	ID        int       // 数据库 ID
	Date      string    // 日期 2006-01-02
	Content   string    // 日记正文
	WordCount int       // 字数统计
	CreatedAt time.Time
	UpdatedAt time.Time
}
