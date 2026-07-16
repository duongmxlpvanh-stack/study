package model

import "time"

// Memo 行政事务备忘
type Memo struct {
	Content   string    // 备忘内容
	CreatedAt time.Time // 添加时间
}
