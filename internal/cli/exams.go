package cli

import (
	"fmt"
	"strconv"

	"study/internal/model"
	"study/internal/render"

	"github.com/spf13/cobra"
)

func newExamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exam",
		Aliases: []string{"ex"},
		Short:   "管理考试倒计时",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "添加考试",
		Example: "study exam add \"期末考试\" 2026-07-15",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			date := args[1]
			if err := ExamSvc.Add(name, date); err != nil {
				return err
			}
			InvalidateCache("exams")
			afterWrite("添加考试: %s (%s)", name, date)
			fmt.Printf(render.Green("✅ 已添加考试: %s (%s)\n"), name, date)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出所有考试倒计时",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 尝试从缓存读取
			exams, _ := cacheGet("exams").([]model.ExamWithCountdown)

			var err error
			if exams == nil {
				exams, err = ExamSvc.List()
				if err != nil {
					return err
				}
				cacheSet("exams", exams)
			}

			if len(exams) == 0 {
				fmt.Println(render.Dim("暂无考试，使用 study exam add 添加"))
				return nil
			}
			fmt.Println(render.Section("⏰ 考试倒计时"))
			t := render.NewTable("序号", "考试名称", "考试日期", "剩余天数", "状态")
			for i, e := range exams {
				urgencyColor := render.Green
				switch {
				case e.DaysLeft < 0:
					urgencyColor = render.Dim
				case e.DaysLeft <= 7:
					urgencyColor = render.Red
				case e.DaysLeft <= 30:
					urgencyColor = render.Yellow
				}
				t.AddRow(
					strconv.Itoa(i+1),
					e.Name,
					render.Dim(e.Date),
					render.Bold(strconv.Itoa(e.DaysLeft))+" 天",
					urgencyColor(e.UrgencyStr),
				)
			}
			fmt.Print(t.Render())
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "del",
		Aliases: []string{"d", "rm"},
		Short:   "删除考试（按序号）",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("请输入序号数字，使用 study exam list 查看")
			}
			if err := ExamSvc.Delete(idx); err != nil {
				return err
			}
			InvalidateCache("exams")
			afterWrite("删除考试")
			fmt.Println(render.Green("✅ 已删除考试"))
			return nil
		},
	})

	return cmd
}
