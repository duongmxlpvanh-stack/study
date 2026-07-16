package model

import "time"

// Exam 考试信息
type Exam struct {
	Name string    // 考试名称，如"期末考试"
	Date time.Time // 考试日期
}

// ExamWithCountdown 带倒计时的考试（展示用）
type ExamWithCountdown struct {
	Name       string // 考试名称
	Date       string // 考试日期 2006-01-02
	DaysLeft   int    // 剩余天数
	UrgencyStr string // 紧急程度文字：🔴临近 / 🟡适中 / 🟢充裕
}
