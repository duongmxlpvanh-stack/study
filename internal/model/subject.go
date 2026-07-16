package model

import "time"

// Subject 科目（课程）
type Subject struct {
	Name      string    // 科目名称，如"高等数学"
	CreatedAt time.Time // 添加时间
}

// SubjectWithCount 带资料数量的科目（Dashboard 用）
type SubjectWithCount struct {
	Name         string
	MaterialCount int // 资料文件夹中的文件数
}
