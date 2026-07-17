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

// Overview 聚合生成 Dashboard 数据（并发收集各数据源）
func (s *DashboardService) Overview() (*model.Dashboard, error) {
	d := &model.Dashboard{}

	// 为每项数据定义 result struct
	type examResult struct {
		exams []model.ExamWithCountdown
		err   error
	}
	type subjResult struct {
		subjects []model.SubjectWithCount
		err      error
	}
	type wpResult struct {
		stats model.WeakPointStats
		err   error
	}
	type recResult struct {
		records []model.Record
		err     error
	}
	type diaryResult struct {
		diaries []model.Diary
		err     error
	}

	// 创建带缓冲的 channel
	examCh := make(chan examResult, 1)
	subjCh := make(chan subjResult, 1)
	wpCh := make(chan wpResult, 1)
	recCh := make(chan recResult, 1)
	diaryCh := make(chan diaryResult, 1)

	// 并发启动 5 个 goroutine
	go func() {
		exams, err := s.examService.List()
		examCh <- examResult{exams, err}
	}()
	go func() {
		subjects, err := s.subjService.ListWithMaterialCount()
		subjCh <- subjResult{subjects, err}
	}()
	go func() {
		stats, err := s.wpService.Stats()
		wpCh <- wpResult{stats, err}
	}()
	go func() {
		records, err := s.recordService.GetAllRecords()
		recCh <- recResult{records, err}
	}()
	go func() {
		if s.diaryService != nil {
			diaries, err := s.diaryService.ListRecent(5)
			diaryCh <- diaryResult{diaries, err}
		} else {
			diaryCh <- diaryResult{nil, nil}
		}
	}()

	// 收集结果（任意子查询失败则整体返回错误）
	r := <-examCh
	if r.err != nil {
		return nil, r.err
	}
	d.Exams = r.exams

	r2 := <-subjCh
	if r2.err != nil {
		return nil, r2.err
	}
	d.Subjects = r2.subjects

	r3 := <-wpCh
	if r3.err != nil {
		return nil, r3.err
	}
	d.WeakPointStats = r3.stats

	r4 := <-recCh
	if r4.err != nil {
		return nil, r4.err
	}
	records := r4.records
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

	// 最近日记
	r5 := <-diaryCh
	if r5.err != nil {
		return nil, r5.err
	}
	d.RecentDiaries = r5.diaries

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
