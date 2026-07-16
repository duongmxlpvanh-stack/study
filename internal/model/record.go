package model

import "time"

// Record 一条学习记录
type Record struct {
	Date    string    // 日期，格式 2006-01-02
	Subject string    // 科目名称
	Content string    // 学习内容描述
	Time    time.Time // 精确时间戳
}
