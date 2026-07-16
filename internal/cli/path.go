package cli

import (
	"fmt"
	"path/filepath"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "管理系统 PATH（Win+R 输入 study 启动）",
		Long: `将 study.exe 所在目录添加到用户 PATH 环境变量，
使你可以通过 Win+R 输入 study 来快速启动。`,
		// 无子命令时显示状态
		RunE: func(cmd *cobra.Command, args []string) error {
			return showPathStatus()
		},
	}

	cmd.AddCommand(newPathAddCmd())
	cmd.AddCommand(newPathCheckCmd())

	return cmd
}

func newPathAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "将 study.exe 添加到用户 PATH",
		Long: `将 study.exe 所在目录添加到用户级 PATH 环境变量。
添加后，你可以在 Win+R 中输入 study 来快速启动。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			added, err := AddToPath()
			if err != nil {
				return fmt.Errorf("添加 PATH 失败: %w", err)
			}
			if !added {
				fmt.Println(render.Green("✅ study 已在 PATH 中，无需重复添加"))
				return nil
			}
			fmt.Println(render.Green("✅ 已添加到用户 PATH"))
			fmt.Println(render.Dim("现在可以在 Win+R 中输入 study 启动了。"))
			fmt.Println(render.Dim("如果刚添加后不生效，请重新登录或重启电脑。"))
			return nil
		},
	}
}

func newPathCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "检查 study.exe 是否在 PATH 中",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showPathStatus()
		},
	}
}

func showPathStatus() error {
	binaryDir, err := getBinaryDir()
	if err != nil {
		return err
	}

	inPath, err := isInPath(binaryDir)
	if err != nil {
		return fmt.Errorf("检查 PATH 失败: %w", err)
	}

	fmt.Printf("📂 study.exe 位置: %s\n", render.Bold(filepath.Join(binaryDir, "study.exe")))
	if inPath {
		fmt.Println(render.Green("✅ 已在系统 PATH 中 — Win+R 输入 study 即可启动"))
	} else {
		fmt.Println(render.Yellow("⚠️ 不在系统 PATH 中"))
		fmt.Println()
		fmt.Println(render.Dim("运行 study path add 将 study 添加到 PATH，"))
		fmt.Println(render.Dim("之后就可以在 Win+R 中输入 study 启动。"))
	}

	return nil
}

