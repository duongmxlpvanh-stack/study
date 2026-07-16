package service

import (
	"sort"
	"time"

	"study/internal/config"
	"study/internal/model"
)

// StreakService 连续学习统计服务
type StreakService struct {
	recordService *RecordService
}

func NewStreakService(cfg *config.Config, rs *RecordService) *StreakService {
	return &StreakService{recordService: rs}
}

// Compute 计算学习统计
func (s *StreakService) Compute() (model.StudyStats, error) {
	records, err := s.recordService.GetAllRecords()
	if err != nil {
		return model.StudyStats{}, err
	}

	if len(records) == 0 {
		return model.StudyStats{}, nil
	}

	// 构建日期集合
	dateSet := make(map[string]bool)
	for _, r := range records {
		dateSet[r.Date] = true
	}

	// 排序日期
	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	totalDays := len(dates)
	totalRecords := len(records)
	streak := calcCurrentStreak(dates)
	avgPerDay := float64(totalRecords) / float64(totalDays)

	return model.StudyStats{
		TotalDays:    totalDays,
		TotalRecords: totalRecords,
		StreakDays:   streak,
		AvgPerDay:    avgPerDay,
	}, nil
}

// calcCurrentStreak 计算当前连续学习天数
// 从今天开始往前推，寻找连续的有记录日期
func calcCurrentStreak(sortedDates []string) int {
	if len(sortedDates) == 0 {
		return 0
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 检查今天或昨天是否有记录（起始点）
	lastDate := sortedDates[len(sortedDates)-1]
	if lastDate != today {
		// 如果最后记录不是今天，检查是否是昨天
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
		if lastDate != yesterday {
			return 0 // 今天和昨天都没记录，连续中断
		}
	}

	// 从后往前数连续天数
	streak := 1
	for i := len(sortedDates) - 1; i > 0; i-- {
		current, _ := time.Parse("2006-01-02", sortedDates[i])
		prev, _ := time.Parse("2006-01-02", sortedDates[i-1])

		diff := current.Sub(prev).Hours() / 24
		if diff == 1.0 {
			streak++
		} else {
			break
		}
	}
	return streak
}
