package model

import "time"

// Urgency 紧急程度
type Urgency string

const (
	UrgencyUrgent    Urgency = "紧急" // 马上要考
	UrgencyRelaxed   Urgency = "不急" // 可以以后看
	UrgencyPreExam   Urgency = "考前看" // 考前临时抱佛脚
)

// WeakPoint 薄弱知识点
type WeakPoint struct {
	Content   string    // 知识点内容
	Urgency   Urgency   // 紧急程度
	Subject   string    // 关联科目（可选）
	CreatedAt time.Time // 添加时间
}

// WeakPointStats 薄弱点统计（Dashboard 用）
type WeakPointStats struct {
	Urgent  int // 紧急数量
	Relaxed int // 不急数量
	PreExam int // 考前看数量
}

func (s WeakPointStats) Total() int {
	return s.Urgent + s.Relaxed + s.PreExam
}
