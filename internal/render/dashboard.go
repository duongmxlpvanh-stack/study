package render

import (
	"fmt"
	"strings"

	"study/internal/model"
)

// Dashboard 渲染全局仪表板
func Dashboard(d *model.Dashboard) string {
	var sb strings.Builder

	// 标题
	sb.WriteString(Title("📋 学习仪表板"))
	sb.WriteString("\n\n")

	// === 统计卡片行 ===
	sb.WriteString(Section("📊 学习统计"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  累计学习 %s 天  │  总记录 %s 条  │  连续 %s 天  │  日均 %s 条",
		Bold(fmt.Sprint(d.StudyStats.TotalDays)),
		Bold(fmt.Sprint(d.StudyStats.TotalRecords)),
		Bold(fmt.Sprint(d.StudyStats.StreakDays)),
		Bold(fmt.Sprintf("%.1f", d.StudyStats.AvgPerDay)),
	))
	sb.WriteString("\n\n")

	// === 考试倒计时 ===
	sb.WriteString(Section("⏰ 考试倒计时"))
	sb.WriteString("\n")
	if len(d.Exams) == 0 {
		sb.WriteString(Dim("  还没有添加考试，使用 study exam add 添加\n"))
	} else {
		for _, e := range d.Exams {
			urgencyColor := Green
			switch {
			case e.DaysLeft < 0:
				urgencyColor = Dim
			case e.DaysLeft <= 7:
				urgencyColor = Red
			case e.DaysLeft <= 30:
				urgencyColor = Yellow
			}
			sb.WriteString(fmt.Sprintf("  %s  %s  剩余 %s 天  %s\n",
				e.Name,
				e.Date,
				Bold(fmt.Sprint(e.DaysLeft)),
				urgencyColor(e.UrgencyStr),
			))
		}
	}
	sb.WriteString("\n")

	// === 薄弱点统计 ===
	sb.WriteString(Section("🎯 薄弱知识点"))
	sb.WriteString("\n")
	wp := d.WeakPointStats
	if wp.Total() == 0 {
		sb.WriteString(Dim("  暂无记录，使用 study wp add 添加\n"))
	} else {
		sb.WriteString(fmt.Sprintf("  总计 %s 条", Bold(fmt.Sprint(wp.Total()))))
		if wp.Urgent > 0 {
			sb.WriteString(fmt.Sprintf("  │  %s %s 条", Red("紧急"), Bold(fmt.Sprint(wp.Urgent))))
		}
		if wp.PreExam > 0 {
			sb.WriteString(fmt.Sprintf("  │  %s %s 条", Yellow("考前看"), Bold(fmt.Sprint(wp.PreExam))))
		}
		if wp.Relaxed > 0 {
			sb.WriteString(fmt.Sprintf("  │  %s %s 条", Dim("不急"), Bold(fmt.Sprint(wp.Relaxed))))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// === 科目概览 ===
	sb.WriteString(Section("📚 课程概览"))
	sb.WriteString("\n")
	if len(d.Subjects) == 0 {
		sb.WriteString(Dim("  还没有添加课程，使用 study subj add 添加\n"))
	} else {
		for _, s := range d.Subjects {
			sb.WriteString(fmt.Sprintf("  %s  %s 份资料\n",
				s.Name,
				Bold(fmt.Sprint(s.MaterialCount)),
			))
		}
	}
	sb.WriteString("\n")

	// === 最近记录 ===
	sb.WriteString(Section("📝 最近学习"))
	sb.WriteString("\n")
	if len(d.RecentRecords) == 0 {
		sb.WriteString(Dim("  还没有学习记录，使用 study log 记录\n"))
	} else {
		for _, r := range d.RecentRecords {
			sb.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				Dim(r.Date),
				Cyan(r.Subject),
				r.Content,
			))
		}
	}
	sb.WriteString("\n")

	// === 最近日记 ===
	sb.WriteString(Section("📖 最近日记"))
	sb.WriteString("\n")
	if len(d.RecentDiaries) == 0 {
		sb.WriteString(Dim("  还没有写日记，使用 study diary 开始\n"))
	} else {
		for _, d := range d.RecentDiaries {
			preview := d.Content
			// 截取前 50 个字符作为预览
			runes := []rune(d.Content)
			if len(runes) > 50 {
				preview = string(runes[:50]) + "…"
			}
			sb.WriteString(fmt.Sprintf("  %s  %s字  %s\n",
				Dim(d.Date),
				Dim(fmt.Sprint(d.WordCount)),
				preview,
			))
		}
	}

	return sb.String()
}
