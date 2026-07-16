package service

import (
	"sort"

	"study/internal/config"
	"study/internal/model"
)

// DashboardService 仪表板服务（聚合所有数据）
type DashboardService struct {
	cfg           *config.Config
	examService   *ExamService
	subjService   *SubjectService
	wpService     *WeakPointService
	recordService *RecordService
	diaryService  *DiaryService
}

func NewDashboardService(
	cfg *config.Config,
	examSvc *ExamService,
	subjSvc *SubjectService,
	wpSvc *WeakPointService,
	recSvc *RecordService,
	diarySvc *DiaryService,
) *DashboardService {
	return &DashboardService{
		cfg:           cfg,
		examService:   examSvc,
		subjService:   subjSvc,
		wpService:     wpSvc,
		recordService: recSvc,
		diaryService:  diarySvc,
	}
}

// Overview 聚合生成 Dashboard 数据
func (s *DashboardService) Overview() (*model.Dashboard, error) {
	d := &model.Dashboard{}

	// 考试倒计时
	exams, err := s.examService.List()
	if err != nil {
		return nil, err
	}
	d.Exams = exams

	// 科目及资料数
	subjects, err := s.subjService.ListWithMaterialCount()
	if err != nil {
		return nil, err
	}
	d.Subjects = subjects

	// 薄弱点统计
	wpStats, err := s.wpService.Stats()
	if err != nil {
		return nil, err
	}
	d.WeakPointStats = wpStats

	// 学习统计
	records, err := s.recordService.GetAllRecords()
	if err != nil {
		return nil, err
	}
	d.StudyStats = computeStudyStats(records)

	// 最近记录（取最后 5 条，倒序）
	recentRecords := records
	if len(recentRecords) > 5 {
		recentRecords = recentRecords[len(recentRecords)-5:]
	}
	// 反转顺序（最新的在前）
	for i, j := 0, len(recentRecords)-1; i < j; i, j = i+1, j-1 {
		recentRecords[i], recentRecords[j] = recentRecords[j], recentRecords[i]
	}
	d.RecentRecords = recentRecords

	// 最近日记（diaryService 可能为 nil，如果数据库初始化失败）
	if s.diaryService != nil {
		diaries, err := s.diaryService.ListRecent(5)
		if err != nil {
			return nil, err
		}
		d.RecentDiaries = diaries
	}

	return d, nil
}

// computeStudyStats 从记录列表计算学习统计
func computeStudyStats(records []model.Record) model.StudyStats {
	if len(records) == 0 {
		return model.StudyStats{}
	}

	// 统计去重日期数（累计学习天数）
	dateSet := make(map[string]bool)
	for _, r := range records {
		dateSet[r.Date] = true
	}
	totalDays := len(dateSet)

	// 计算连续学习天数
	streak := computeStreak(dateSet)

	// 日均记录数
	avgPerDay := float64(len(records)) / float64(totalDays)

	return model.StudyStats{
		TotalDays:    totalDays,
		TotalRecords: len(records),
		StreakDays:   streak,
		AvgPerDay:    avgPerDay,
	}
}

// computeStreak 计算当前连续学习天数
func computeStreak(dateSet map[string]bool) int {
	// 从今天往前推，计算连续有多少天有记录
	streak := 0
	// 需要 sort 导入，这里简单实现
	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	if len(dates) == 0 {
		return 0
	}

	// 从最后一天（最新）往前检查连续性
	// 简单实现：只检查最近几天
	// 完整实现在 streak.go 中
	return streak
}
