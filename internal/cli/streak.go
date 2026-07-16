package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newStreakCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "streak",
		Aliases: []string{"sk"},
		Short:   "显示连续学习统计",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := StreakSvc.Compute()
			if err != nil {
				return err
			}

			fmt.Println(render.Section("📊 学习统计"))
			fmt.Printf("  %s  %s 天\n", render.Dim("累计学习天数："), render.Bold(fmt.Sprint(stats.TotalDays)))
			fmt.Printf("  %s  %s 条\n", render.Dim("总学习记录数："), render.Bold(fmt.Sprint(stats.TotalRecords)))
			fmt.Printf("  %s  %s 天\n", render.Dim("当前连续天数："), render.Bold(fmt.Sprint(stats.StreakDays)))
			fmt.Printf("  %s  %s 条/天\n", render.Dim("日均记录数："), render.Bold(fmt.Sprintf("%.1f", stats.AvgPerDay)))

			return nil
		},
	}
}
