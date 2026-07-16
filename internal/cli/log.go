package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Aliases: []string{"lg"},
		Short:   "记录学习进度",
		Long: `记录今天学了什么。

用法:
  study log "科目: 内容"
  study log "高等数学: 完成多元函数微分"

也可在 REPL 模式下直接输入: log 内容`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			if err := RecordSvc.Log(input); err != nil {
				return err
			}
			fmt.Println(render.Green("✅ 已记录学习进度"))
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "浏览历史学习记录",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := RecordSvc.ListRecent(20)
			if err != nil {
				return err
			}
			if len(records) == 0 {
				fmt.Println(render.Dim("暂无学习记录"))
				return nil
			}
			fmt.Println(render.Section("📝 最近学习记录"))
			for i := len(records) - 1; i >= 0; i-- {
				r := records[i]
				fmt.Printf("  %s  %s  %s\n",
					render.Dim(r.Date),
					render.Cyan(r.Subject),
					r.Content,
				)
			}
			return nil
		},
	})

	return cmd
}
