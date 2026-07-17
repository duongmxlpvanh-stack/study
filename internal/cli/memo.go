package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newMemoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo",
		Aliases: []string{"mm"},
		Short:   "管理行政事务备忘",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "添加备忘",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := MemoSvc.Add(args[0]); err != nil {
				return err
			}
			afterWrite("添加备忘")
			fmt.Println(render.Green("✅ 已添加备忘"))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出所有备忘",
		RunE: func(cmd *cobra.Command, args []string) error {
			memos, err := MemoSvc.List()
			if err != nil {
				return err
			}
			if len(memos) == 0 {
				fmt.Println(render.Dim("暂无备忘"))
				return nil
			}
			fmt.Println(render.Section("📋 行政备忘"))
			for i, m := range memos {
				fmt.Printf("  %d. %s  %s\n",
					i+1,
					render.Dim(m.CreatedAt.Format("2006-01-02")),
					m.Content,
				)
			}
			return nil
		},
	})

	return cmd
}
