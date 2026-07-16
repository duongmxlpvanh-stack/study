package model

// HeatMapDay 热力图中的一天
type HeatMapDay struct {
	Date  string // 日期 2006-01-02
	Count int    // 当天学习记录数
	Level int    // 颜色等级 0-4（0=无，4=最多）
}
