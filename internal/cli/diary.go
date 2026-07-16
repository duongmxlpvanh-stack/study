package cli

import (
	"fmt"
	"time"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newDiaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diary",
		Aliases: []string{"dj"},
		Short:   "学习日记",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "open",
		Aliases: []string{"o"},
		Short:   "打开/编辑日记（默认今天）",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			date := time.Now().Format("2006-01-02")
			if len(args) > 0 {
				date = args[0]
			}
			if DiarySvc == nil {
				return fmt.Errorf("日记服务未初始化")
			}
			if err := DiarySvc.Open(date); err != nil {
				return err
			}
			fmt.Println(render.Green("📖 日记已保存"))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出最近的日记",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if DiarySvc == nil {
				return fmt.Errorf("日记服务未初始化")
			}
			diaries, err := DiarySvc.ListRecent(10)
			if err != nil {
				return err
			}
			if len(diaries) == 0 {
				fmt.Println(render.Dim("暂无日记"))
				return nil
			}
			fmt.Println(render.Section("📖 最近日记"))
			t := render.NewTable("日期", "字数", "预览")
			for _, d := range diaries {
				preview := d.Content
				runes := []rune(d.Content)
				if len(runes) > 60 {
					preview = string(runes[:60]) + "…"
				}
				t.AddRow(
					render.Dim(d.Date),
					render.Dim(fmt.Sprint(d.WordCount)),
					preview,
				)
			}
			fmt.Print(t.Render())
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "search",
		Aliases: []string{"s", "find"},
		Short:   "全文搜索日记",
		Example: "study diary search 微分方程",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyword := args[0]
			if DiarySvc == nil {
				return fmt.Errorf("日记服务未初始化")
			}
			results, err := DiarySvc.Search(keyword)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Printf(render.Dim("未找到与 \"%s\" 相关的日记\n"), keyword)
				return nil
			}
			fmt.Printf(render.Section("🔍 搜索 \"%s\" 的结果 (%d条)\n"), keyword, len(results))
			for _, d := range results {
				fmt.Printf("  %s  %s\n", render.Dim(d.Date), d.Content)
				fmt.Println(render.Dim("  ──"))
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "del",
		Aliases: []string{"d", "rm"},
		Short:   "删除日记",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			date := args[0]
			if DiarySvc == nil {
				return fmt.Errorf("日记服务未初始化")
			}
			if err := DiarySvc.Delete(date); err != nil {
				return err
			}
			fmt.Printf(render.Green("✅ 已删除 %s 的日记\n"), date)
			return nil
		},
	})

	return cmd
}
