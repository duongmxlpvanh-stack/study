package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"study/internal/auth"
	"study/internal/render"
	"study/internal/service"

	"github.com/spf13/cobra"
)

func newCalendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calendar",
		Aliases: []string{"cal"},
		Short:   "管理 Google Calendar 学习事件",
		Long: `在 Google Calendar 中创建、查看、管理学习计划事件。

所有学习事件都会带 AI 生成标记，可一键清除。`,
	}

	// calendar add
	addCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "创建学习日历事件",
		Long: `在 Google Calendar 中创建一个学习计划事件。

示例:
  study calendar add --subject 高等数学 --focus "复习第三章微分" --duration 60
  study calendar add --subject 线性代数 --focus "做矩阵练习题" --start "2026-07-17T14:00:00+08:00" --duration 45`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if CalendarSvc == nil {
				return fmt.Errorf("Google Calendar 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			subject, _ := cmd.Flags().GetString("subject")
			focus, _ := cmd.Flags().GetString("focus")
			if subject == "" || focus == "" {
				return fmt.Errorf("--subject 和 --focus 为必填参数")
			}

			// 解析开始时间
			var startTime time.Time
			startStr, _ := cmd.Flags().GetString("start")
			if startStr != "" {
				var err error
				startTime, err = time.Parse(time.RFC3339, startStr)
				if err != nil {
					return fmt.Errorf("开始时间格式错误，请使用 RFC3339 格式（如 2026-07-17T14:00:00+08:00）: %w", err)
				}
			} else {
				// 默认：一小时后的整点
				startTime = time.Now().Add(1 * time.Hour)
			}

			// 时长
			durationMin, _ := cmd.Flags().GetInt("duration")
			if durationMin <= 0 {
				durationMin = 60 // 默认 1 小时
			}
			endTime := startTime.Add(time.Duration(durationMin) * time.Minute)

			// 冲突检查
			noConflictCheck, _ := cmd.Flags().GetBool("no-conflict-check")

			params := service.CalendarEventParams{
				Subject:       subject,
				Focus:         focus,
				StartTime:     startTime,
				EndTime:       endTime,
				CheckConflict: !noConflictCheck,
			}

			ctx := context.Background()
			event, err := CalendarSvc.CreateStudyEvent(ctx, params)
			if err != nil {
				return fmt.Errorf("创建学习事件失败: %w", err)
			}

			fmt.Printf("%s 学习事件已添加到 Google Calendar\n", render.Green("✅"))
			fmt.Printf("  📚 %s\n", event.Summary)
			fmt.Printf("  🕐 %s → %s\n",
				startTime.Format("2006-01-02 15:04"),
				endTime.Format("15:04"),
			)
			fmt.Printf("  🔗 %s\n", render.Dim(event.HtmlLink))
			return nil
		},
	}
	addCmd.Flags().StringP("subject", "s", "", "科目名称（必填）")
	addCmd.Flags().StringP("focus", "f", "", "学习重点描述（必填）")
	addCmd.Flags().String("start", "", "开始时间 RFC3339 格式（默认一小时后）")
	addCmd.Flags().IntP("duration", "d", 60, "学习时长（分钟）")
	addCmd.Flags().Bool("no-conflict-check", false, "跳过日程冲突检查")
	cmd.AddCommand(addCmd)

	// calendar list
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出 AI 生成的学习事件",
		RunE: func(cmd *cobra.Command, args []string) error {
			if CalendarSvc == nil {
				return fmt.Errorf("Google Calendar 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			subject, _ := cmd.Flags().GetString("subject")
			days, _ := cmd.Flags().GetInt("days")

			ctx := context.Background()
			maxDate := time.Now().AddDate(0, 0, days)
			events, err := CalendarSvc.ListStudyEvents(ctx, subject, maxDate)
			if err != nil {
				return fmt.Errorf("查询学习事件失败: %w", err)
			}

			fmt.Println(render.Section("📅 学习计划事件"))

			headerText := fmt.Sprintf("  未来 %d 天", days)
			if subject != "" {
				headerText += fmt.Sprintf(" · %s", subject)
			}
			fmt.Println(render.Dim(headerText))
			fmt.Println()

			if len(events) == 0 {
				fmt.Println(render.Dim("  暂无学习计划事件。"))
				fmt.Println(render.Dim("  使用 study calendar add 添加。"))
				return nil
			}

			for i, evt := range events {
				fmt.Printf("  %d. %s\n", i+1, render.Bold(evt.Summary))
				if evt.Start != nil && evt.Start.DateTime != "" {
					startTime, _ := time.Parse(time.RFC3339, evt.Start.DateTime)
					endTime, _ := time.Parse(time.RFC3339, evt.End.DateTime)
					fmt.Printf("     🕐 %s → %s\n",
						startTime.Format("01-02 15:04"),
						endTime.Format("15:04"),
					)
				}
				if evt.ExtendedProperties != nil && evt.ExtendedProperties.Private != nil {
					subj := evt.ExtendedProperties.Private["study_subject"]
					focus := evt.ExtendedProperties.Private["study_focus"]
					if subj != "" {
						fmt.Printf("     📚 %s: %s\n", subj, render.Dim(focus))
					}
				}
				fmt.Println()
			}

			fmt.Printf(render.Dim("  共 %d 个事件\n"), len(events))
			fmt.Println()
			fmt.Println(render.Dim("  提示: 在 Google Calendar 中也可查看这些事件。"))
			return nil
		},
	}
	listCmd.Flags().StringP("subject", "s", "", "按科目筛选（可选）")
	listCmd.Flags().IntP("days", "d", 30, "查询未来 N 天")
	cmd.AddCommand(listCmd)

	// calendar conflicts
	cmd.AddCommand(&cobra.Command{
		Use:   "conflicts",
		Short: "检查指定时间段是否有日程冲突",
		Long: `检查指定时间段内是否有非学习类日程冲突。

示例:
  study calendar conflicts --start "2026-07-17T14:00:00+08:00" --end "2026-07-17T16:00:00+08:00"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if CalendarSvc == nil {
				return fmt.Errorf("Google Calendar 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			startStr, _ := cmd.Flags().GetString("start")
			endStr, _ := cmd.Flags().GetString("end")
			if startStr == "" || endStr == "" {
				return fmt.Errorf("--start 和 --end 为必填参数")
			}

			startTime, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				return fmt.Errorf("开始时间格式错误: %w", err)
			}
			endTime, err := time.Parse(time.RFC3339, endStr)
			if err != nil {
				return fmt.Errorf("结束时间格式错误: %w", err)
			}

			ctx := context.Background()
			conflicts, err := CalendarSvc.CheckConflicts(ctx, startTime, endTime)
			if err != nil {
				return fmt.Errorf("检查冲突失败: %w", err)
			}

			fmt.Println(render.Section("📅 日程冲突检查"))
			fmt.Printf("  %s → %s\n",
				startTime.Format("2006-01-02 15:04"),
				endTime.Format("15:04"),
			)
			fmt.Println()

			if len(conflicts) == 0 {
				fmt.Printf("  %s 该时段空闲，无冲突\n", render.Green("✅"))
			} else {
				fmt.Printf("  %s 发现 %d 个冲突日程:\n", render.Yellow("⚠️"), len(conflicts))
				for i, evt := range conflicts {
					fmt.Printf("    %d. %s", i+1, evt.Summary)
					if evt.Start != nil {
						fmt.Printf(" (%s)", evt.Start.DateTime)
					}
					fmt.Println()
				}
			}

			return nil
		},
	})
	// Add flags to the conflicts subcommand
	conflictsCmd := cmd.Commands()[len(cmd.Commands())-1]
	conflictsCmd.Flags().String("start", "", "开始时间 RFC3339 格式")
	conflictsCmd.Flags().String("end", "", "结束时间 RFC3339 格式")

	// calendar clear
	cmd.AddCommand(&cobra.Command{
		Use:     "clear",
		Aliases: []string{"clr"},
		Short:   "删除所有 AI 生成的学习事件",
		Long: `删除 Google Calendar 中所有标记为 AI 生成的学习事件。

此操作不可撤销，执行前会要求确认。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if CalendarSvc == nil {
				return fmt.Errorf("Google Calendar 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			// 确认
			skipConfirm, _ := cmd.Flags().GetBool("yes")
			if !skipConfirm {
				fmt.Print(render.Yellow("⚠️ 确定要删除所有 AI 生成的学习事件吗？此操作不可撤销。(y/N): "))
				var answer string
				fmt.Scanln(&answer)
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println(render.Dim("已取消。"))
					return nil
				}
			}

			ctx := context.Background()
			count, err := CalendarSvc.DeleteAllAIEvents(ctx)
			if err != nil {
				return fmt.Errorf("删除事件失败: %w", err)
			}

			fmt.Printf("%s 已删除 %d 个学习事件\n", render.Green("✅"), count)
			return nil
		},
	})
	clearCmd := cmd.Commands()[len(cmd.Commands())-1]
	clearCmd.Flags().BoolP("yes", "y", false, "跳过确认提示")

	// calendar status
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "查看学习事件统计",
		RunE: func(cmd *cobra.Command, args []string) error {
			if CalendarSvc == nil {
				return fmt.Errorf("Google Calendar 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			info := auth.GetAuthInfo()
			if !info.IsAuthorized {
				fmt.Println(render.Red("❌ Google 未授权，请运行 study google login"))
				return nil
			}

			ctx := context.Background()
			stats, err := CalendarSvc.GetStats(ctx)
			if err != nil {
				return fmt.Errorf("查询统计失败: %w", err)
			}

			fmt.Println(render.Section("📅 Google Calendar 学习统计"))
			fmt.Println()
			fmt.Printf("  📊 学习事件总数: %d\n", stats.TotalEvents)

			if stats.NextEvent != "" {
				fmt.Printf("  📅 最近事件: %s\n", stats.NextEvent)
				if !stats.NextTime.IsZero() {
					fmt.Printf("     🕐 %s\n", stats.NextTime.Format("2006-01-02 15:04"))
				}
			}

			if len(stats.Subjects) > 0 {
				fmt.Printf("  📚 涉及科目: %s\n", strings.Join(stats.Subjects, "、"))
			}

			if stats.TotalEvents > 0 {
				fmt.Println()
				fmt.Println(render.Dim("  使用 study calendar list 查看详细列表"))
				fmt.Println(render.Dim("  使用 study calendar clear 清除所有事件"))
			}

			return nil
		},
	})

	return cmd
}
