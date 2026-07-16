package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newHeatmapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "heatmap",
		Aliases: []string{"hm"},
		Short:   "显示学习热力图（GitHub 风格）",
		Long:    "展示过去约 140 天的学习频率热力图，支持按科目筛选。",
		RunE: func(cmd *cobra.Command, args []string) error {
			subject, _ := cmd.Flags().GetString("subject")
			days, err := HeatSvc.Generate(subject)
			if err != nil {
				return err
			}

			title := "🔥 学习热力图"
			if subject != "" {
				title = fmt.Sprintf("🔥 学习热力图 - %s", subject)
			}
			fmt.Println(render.Section(title))
			fmt.Print(render.Heatmap(days))
			return nil
		},
	}

	cmd.Flags().StringP("subject", "s", "", "按科目筛选（可选）")
	return cmd
}
