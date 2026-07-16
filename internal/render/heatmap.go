package render

import (
	"fmt"
	"strings"

	"study/internal/model"
)

// Heatmap 渲染 GitHub 风格的热力图
func Heatmap(days []model.HeatMapDay) string {
	if len(days) == 0 {
		return Dim("没有数据")
	}

	var sb strings.Builder

	// 按周分组（每周 7 天，周日起）
	weeks := groupByWeek(days)

	// 星期标签
	dayLabels := []string{"日", "一", "二", "三", "四", "五", "六"}

	for row := 0; row < 7; row++ {
		// 只在第 0, 2, 4, 6 行显示标签
		label := "  "
		if row%2 == 0 {
			label = dayLabels[row] + " "
		}
		sb.WriteString(Dim(label))

		for _, week := range weeks {
			if row < len(week) {
				sb.WriteString(HeatBlock(week[row].Level))
			} else {
				sb.WriteString("  ") // 补齐
			}
		}
		sb.WriteString("\n")
	}

	// 图例
	sb.WriteString("\n")
	sb.WriteString(Dim("  Less "))
	sb.WriteString(HeatBlock(0))
	sb.WriteString(HeatBlock(1))
	sb.WriteString(HeatBlock(2))
	sb.WriteString(HeatBlock(3))
	sb.WriteString(HeatBlock(4))
	sb.WriteString(Dim(" More"))
	sb.WriteString("\n")

	return sb.String()
}

// groupByWeek 将日期按周分组
func groupByWeek(days []model.HeatMapDay) [][]model.HeatMapDay {
	if len(days) == 0 {
		return nil
	}

	var weeks [][]model.HeatMapDay
	var currentWeek []model.HeatMapDay

	// 计算第一天是星期几（补齐第一周前面的空位）
	firstDate := days[0].Date
	firstDayOfWeek := weekdayOf(firstDate) // 0=Sun

	// 补齐前面的空白
	for i := 0; i < firstDayOfWeek; i++ {
		currentWeek = append(currentWeek, model.HeatMapDay{Level: -1})
	}

	for _, d := range days {
		dow := weekdayOf(d.Date)
		if dow == 0 && len(currentWeek) > 0 {
			// 新的一周
			weeks = append(weeks, currentWeek)
			currentWeek = nil
		}
		currentWeek = append(currentWeek, d)
	}

	// 最后一周
	if len(currentWeek) > 0 {
		// 补齐末尾
		for len(currentWeek) < 7 {
			currentWeek = append(currentWeek, model.HeatMapDay{Level: -1})
		}
		weeks = append(weeks, currentWeek)
	}

	return weeks
}

// weekdayOf 返回日期是星期几 (0=周日, 1=周一, ...)
func weekdayOf(dateStr string) int {
	// dateStr 格式 2006-01-02
	// 简化：解析 YYYY-MM-DD 计算星期
	var y, m, d int
	fmt.Sscanf(dateStr, "%d-%d-%d", &y, &m, &d)

	// Zeller's formula 简化版
	if m < 3 {
		m += 12
		y--
	}
	k := y % 100
	j := y / 100
	h := (d + (13*(m+1))/5 + k + k/4 + j/4 - 2*j) % 7
	// h: 0=Sat, 1=Sun, ...
	switch h {
	case 0:
		return 6 // Sat -> 6
	case 1:
		return 0 // Sun -> 0
	default:
		return h - 1 // Mon=1...Fri=5
	}
}
