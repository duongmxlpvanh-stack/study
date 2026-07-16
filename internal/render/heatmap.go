package render

import (
	"fmt"
	"strings"

	"study/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// 当前日期高亮边框
var todayHighlight = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#58a6ff")).
	BorderTop(false).BorderBottom(false).BorderLeft(false)

// Heatmap 渲染 GitHub 风格的热力图
func Heatmap(days []model.HeatMapDay) string {
	if len(days) == 0 {
		return Dim("没有数据")
	}

	// 按周分组
	weeks := groupByWeek(days)

	var sb strings.Builder

	// 月份标签行
	monthLabels := buildMonthLabels(weeks)
	if len(monthLabels) > 0 {
		sb.WriteString(renderMonthRow(monthLabels, len(weeks)))
		sb.WriteString("\n")
	}

	// 星期标签
	dayLabels := []string{"日", "一", "二", "三", "四", "五", "六"}

	// 获取今天的日期用于高亮
	todayStr := ""
	if len(days) > 0 {
		// 取 last entry date as approximation of today (latest data day)
		todayStr = days[len(days)-1].Date
	}

	for row := 0; row < 7; row++ {
		label := "  "
		if row%2 == 0 {
			label = dayLabels[row] + " "
		}
		sb.WriteString(Dim(label))

		for col, week := range weeks {
			if row < len(week) {
				day := week[row]
				block := renderHeatCell(day.Level)
				// 在最后一天（今天）加高亮
				if day.Date == todayStr && day.Level > 0 {
					block = todayHighlight.Render(strings.TrimRight(block, " "))
				}
				sb.WriteString(block)
			} else {
				sb.WriteString("  ")
			}
			// 列间微间距
			if col < len(weeks)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// 图例
	sb.WriteString("\n")
	sb.WriteString(Dim("  Less "))
	for i := 0; i <= 4; i++ {
		sb.WriteString(renderHeatCell(i))
		sb.WriteString(" ")
	}
	sb.WriteString(Dim("More"))
	sb.WriteString("\n")

	return sb.String()
}

// renderHeatCell 渲染单个热力图单元格
func renderHeatCell(level int) string {
	if level < 0 {
		return "  "
	}
	if level > 4 {
		level = 4
	}
	return heatStyles[level].Render("  ")
}

// monthLabel 月份标签（月份，列索引）
type monthLabel struct {
	month string
	col   int
}

// buildMonthLabels 构建月份标签列表
func buildMonthLabels(weeks [][]model.HeatMapDay) []monthLabel {
	if len(weeks) == 0 {
		return nil
	}

	var labels []monthLabel
	lastMonth := ""
	monthNames := []string{"", "1月", "2月", "3月", "4月", "5月", "6月",
		"7月", "8月", "9月", "10月", "11月", "12月"}

	for col, week := range weeks {
		// 找该周第一个有效的日期
		for _, day := range week {
			if day.Level < 0 {
				continue
			}
			// 解析月份
			var y, m, d int
			fmt.Sscanf(day.Date, "%d-%d-%d", &y, &m, &d)
			monthName := monthNames[m]
			if monthName != lastMonth {
				labels = append(labels, monthLabel{month: monthName, col: col})
				lastMonth = monthName
			}
			break
		}
	}

	return labels
}

// renderMonthRow 渲染月份标签行
func renderMonthRow(labels []monthLabel, totalCols int) string {
	// 每个单元格宽度 = 2（色块） + 1（间距）
	cellWidth := 3

	// 构建一个字符数组表示月份行
	row := make([]rune, totalCols*cellWidth+2) // +2 for left padding
	for i := range row {
		row[i] = ' '
	}

	for _, label := range labels {
		pos := 2 + label.col*cellWidth  // 2 for left padding
		runes := []rune(label.month)
		for i, r := range runes {
			idx := pos + i
			if idx < len(row) {
				row[idx] = r
			}
		}
	}

	return Dim(string(row))
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
