package service

import (
	"time"

	"study/internal/config"
	"study/internal/model"
)

const heatmapDays = 140 // 显示约 20 周

// HeatmapService 热力图服务
type HeatmapService struct {
	recordService *RecordService
}

func NewHeatmapService(cfg *config.Config, rs *RecordService) *HeatmapService {
	return &HeatmapService{recordService: rs}
}

// Generate 生成热力图数据
// subject 为空时展示所有科目，否则只展示指定科目
func (s *HeatmapService) Generate(subject string) ([]model.HeatMapDay, error) {
	records, err := s.recordService.GetAllRecords()
	if err != nil {
		return nil, err
	}

	// 按日期聚合计数
	dateCount := make(map[string]int)
	for _, r := range records {
		if subject != "" && r.Subject != subject {
			continue
		}
		dateCount[r.Date]++
	}

	// 找到最大计数（用于确定颜色等级）
	maxCount := 0
	for _, c := range dateCount {
		if c > maxCount {
			maxCount = c
		}
	}

	// 生成过去 140 天的数据
	now := time.Now()
	var days []model.HeatMapDay
	for i := heatmapDays - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := dateCount[date]
		level := calcLevel(count, maxCount)
		days = append(days, model.HeatMapDay{
			Date:  date,
			Count: count,
			Level: level,
		})
	}

	return days, nil
}

// calcLevel 根据计数计算颜色等级 0-4
func calcLevel(count, max int) int {
	if count == 0 || max == 0 {
		return 0
	}
	// 四等分
	if max <= 1 {
		if count >= 1 {
			return 4
		}
		return 0
	}
	ratio := float64(count) / float64(max)
	switch {
	case ratio <= 0.25:
		return 1
	case ratio <= 0.5:
		return 2
	case ratio <= 0.75:
		return 3
	default:
		return 4
	}
}
