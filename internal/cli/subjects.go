package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newSubjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subject",
		Aliases: []string{"subj", "sub"},
		Short:   "管理科目与资料",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "添加科目",
		Example: "study subj add 高等数学",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := SubjSvc.Add(name); err != nil {
				return err
			}
			fmt.Printf(render.Green("✅ 已添加科目: %s（资料文件夹已创建）\n"), name)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出所有科目",
		RunE: func(cmd *cobra.Command, args []string) error {
			subjects, err := SubjSvc.ListWithMaterialCount()
			if err != nil {
				return err
			}
			if len(subjects) == 0 {
				fmt.Println(render.Dim("暂无科目，使用 study subj add 添加"))
				return nil
			}
			fmt.Println(render.Section("📚 课程列表"))
			for i, s := range subjects {
				fmt.Printf("  %d. %s  %s 份资料\n",
					i+1, s.Name,
					render.Bold(fmt.Sprint(s.MaterialCount)),
				)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "open",
		Aliases: []string{"o"},
		Short:   "打开科目资料文件夹",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dir := cfg.SubjectMaterialsDir(name)

			var c *exec.Cmd
			switch runtime.GOOS {
			case "windows":
				c = exec.Command("explorer", dir)
			case "darwin":
				c = exec.Command("open", dir)
			default:
				c = exec.Command("xdg-open", dir)
			}
			if err := c.Run(); err != nil {
				return fmt.Errorf("打开文件夹失败: %w\n路径: %s", err, dir)
			}
			fmt.Printf(render.Green("📂 已打开: %s\n"), dir)
			return nil
		},
	})

	return cmd
}
