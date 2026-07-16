package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "overview",
		Aliases: []string{"ov", "st"},
		Short:   "显示全局仪表板 Dashboard",
		Long:    "汇总考试倒计时、科目、薄弱点、学习统计、最近记录和日记。",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := DashSvc.Overview()
			if err != nil {
				return fmt.Errorf("生成仪表板失败: %w", err)
			}
			fmt.Print(render.Dashboard(d))
			return nil
		},
	}
}
