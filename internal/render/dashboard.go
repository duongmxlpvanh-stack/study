package render

import (
	"fmt"
	"strings"

	"study/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// Dashboard 渲染全局仪表板
func Dashboard(d *model.Dashboard) string {
	var sb strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().Bold(true).PaddingBottom(1).Width(80)
	sb.WriteString(titleStyle.Render("📋 学习仪表板"))
	sb.WriteString("\n")

	// === 统计卡片行 ===
	sb.WriteString(renderStatCards(d.StudyStats))
	sb.WriteString("\n")

	// === 考试倒计时 ===
	sb.WriteString(Section("⏰ 考试倒计时"))
	sb.WriteString("\n")
	sb.WriteString(renderExamSection(d.Exams))
	sb.WriteString("\n")

	// === 薄弱点统计 ===
	sb.WriteString(Section("🎯 薄弱知识点"))
	sb.WriteString("\n")
	sb.WriteString(renderWeakPointSummary(d.WeakPointStats))
	sb.WriteString("\n")

	// === 课程概览 ===
	sb.WriteString(renderSubjectSection(d.Subjects))
	sb.WriteString("\n")

	// === 最近学习 ===
	sb.WriteString(renderRecentRecords(d.RecentRecords))
	sb.WriteString("\n")

	// === 最近日记 ===
	sb.WriteString(Section("📖 最近日记"))
	sb.WriteString("\n")
	sb.WriteString(renderRecentDiaries(d.RecentDiaries))

	return sb.String()
}

// renderStatCards 渲染 4 个统计卡片并排
func renderStatCards(stats model.StudyStats) string {
	cards := []struct {
		label string
		value string
	}{
		{"累计学习", fmt.Sprintf("%d 天", stats.TotalDays)},
		{"总记录", fmt.Sprintf("%d 条", stats.TotalRecords)},
		{"连续学习", fmt.Sprintf("%d 天", stats.StreakDays)},
		{"日均记录", fmt.Sprintf("%.1f 条", stats.AvgPerDay)},
	}

	labelStyle := lipgloss.NewStyle().Faint(true).Align(lipgloss.Center).Width(16)
	valueStyle := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Width(16)

	var rendered []string
	for _, c := range cards {
		inner := lipgloss.JoinVertical(lipgloss.Center,
			labelStyle.Render(c.label),
			valueStyle.Render(c.value),
		)
		card := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Render(inner)
		rendered = append(rendered, card)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// renderExamSection 渲染考试列表（带紧急程度颜色）
func renderExamSection(exams []model.ExamWithCountdown) string {
	if len(exams) == 0 {
		return Dim("  还没有添加考试，使用 study exam add 添加\n")
	}

	var rows []string
	for _, e := range exams {
		urgencyColor := Green
		switch {
		case e.DaysLeft < 0:
			urgencyColor = Dim
		case e.DaysLeft <= 7:
			urgencyColor = Red
		case e.DaysLeft <= 30:
			urgencyColor = Yellow
		}

		row := fmt.Sprintf("  %-12s  %s  剩余 %s 天  %s",
			e.Name,
			Dim(e.Date),
			Bold(fmt.Sprint(e.DaysLeft)),
			urgencyColor(e.UrgencyStr),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n") + "\n"
}

// renderWeakPointSummary 渲染薄弱点统计摘要
func renderWeakPointSummary(wp model.WeakPointStats) string {
	if wp.Total() == 0 {
		return Dim("  暂无记录，使用 study wp add 添加\n")
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("  总计 %s 条", Bold(fmt.Sprint(wp.Total()))))
	if wp.Urgent > 0 {
		parts = append(parts, fmt.Sprintf("%s %s 条", Red("紧急"), Bold(fmt.Sprint(wp.Urgent))))
	}
	if wp.PreExam > 0 {
		parts = append(parts, fmt.Sprintf("%s %s 条", Yellow("考前看"), Bold(fmt.Sprint(wp.PreExam))))
	}
	if wp.Relaxed > 0 {
		parts = append(parts, fmt.Sprintf("%s %s 条", Dim("不急"), Bold(fmt.Sprint(wp.Relaxed))))
	}

	return strings.Join(parts, "  │  ") + "\n"
}

// renderSubjectSection 渲染科目概览
func renderSubjectSection(subjects []model.SubjectWithCount) string {
	var sb strings.Builder
	sb.WriteString(Section("📚 课程概览"))
	sb.WriteString("\n")
	if len(subjects) == 0 {
		sb.WriteString(Dim("  还没有添加课程\n"))
	} else {
		for _, s := range subjects {
			sb.WriteString(fmt.Sprintf("  %s  %s 份资料\n",
				s.Name,
				Bold(fmt.Sprint(s.MaterialCount)),
			))
		}
	}
	return sb.String()
}

// renderRecentRecords 渲染最近学习记录
func renderRecentRecords(records []model.Record) string {
	var sb strings.Builder
	sb.WriteString(Section("📝 最近学习"))
	sb.WriteString("\n")
	if len(records) == 0 {
		sb.WriteString(Dim("  还没有学习记录，使用 study log 记录\n"))
	} else {
		for _, r := range records {
			sb.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				Dim(r.Date),
				Cyan(r.Subject),
				r.Content,
			))
		}
	}
	return sb.String()
}

// renderRecentDiaries 渲染最近日记
func renderRecentDiaries(diaries []model.Diary) string {
	if len(diaries) == 0 {
		return Dim("  还没有写日记，使用 study diary 开始\n")
	}

	var rows []string
	for _, d := range diaries {
		preview := d.Content
		runes := []rune(d.Content)
		if len(runes) > 50 {
			preview = string(runes[:50]) + "…"
		}
		rows = append(rows, fmt.Sprintf("  %s  %s字  %s",
			Dim(d.Date),
			Dim(fmt.Sprint(d.WordCount)),
			preview,
		))
	}
	return strings.Join(rows, "\n") + "\n"
}
