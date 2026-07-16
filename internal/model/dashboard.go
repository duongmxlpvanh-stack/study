package model

// Dashboard 全局仪表板数据
type Dashboard struct {
	Exams           []ExamWithCountdown // 考试倒计时
	Subjects        []SubjectWithCount  // 科目及资料数量
	WeakPointStats  WeakPointStats      // 薄弱点统计
	StudyStats      StudyStats          // 学习统计
	RecentRecords   []Record            // 最近学习记录（最多 5 条）
	RecentDiaries   []Diary             // 最近日记（最多 5 条）
}

// StudyStats 学习统计数据
type StudyStats struct {
	TotalDays    int     // 累计学习天数
	TotalRecords int     // 总记录数
	StreakDays   int     // 当前连续学习天数
	AvgPerDay    float64 // 日均记录数
}
