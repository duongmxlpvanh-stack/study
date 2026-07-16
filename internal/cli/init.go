package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "首次运行引导向导",
		Long: `交互式设置向导，帮助你快速完成初始配置。

引导你完成：
  1. 添加本学期课程
  2. 添加考试日期
  3. 设置数据目录（可选）`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitWizard()
		},
	}
}

func runInitWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(render.Title("🚀 欢迎使用 study 管理中心！"))
	fmt.Println()
	fmt.Println(render.Dim("这将是一个简短的设置向导，帮你快速上手。"))
	fmt.Println(render.Dim("你可以随时按 Ctrl+C 退出，之后用 study init 重新开始。"))
	fmt.Println()

	// 1. 确保数据目录存在
	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}
	fmt.Printf(render.Green("✅ 数据目录: %s\n"), cfg.DataDir)
	fmt.Println()

	// 2. 添加课程
	fmt.Println(render.Section("📚 步骤 1/3: 添加本学期课程"))
	fmt.Println(render.Dim("  输入课程名称，一行一个。输入空行结束。"))
	fmt.Println(render.Dim("  例如: 高等数学、大学物理、线性代数"))

	for {
		fmt.Print("  课程名称: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			break
		}
		if err := SubjSvc.Add(name); err != nil {
			fmt.Printf("  %s %v\n", render.Red("添加失败:"), err)
		} else {
			fmt.Printf("  %s 已添加: %s\n", render.Green("✅"), name)
		}
	}
	fmt.Println()

	// 3. 添加考试
	fmt.Println(render.Section("⏰ 步骤 2/3: 添加考试日期"))
	fmt.Println(render.Dim("  输入考试名称和日期（YYYY-MM-DD），输入空行结束。"))
	fmt.Println(render.Dim("  例如: 期末考试 2026-07-15"))

	for {
		fmt.Print("  考试名称和日期: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			break
		}
		// 解析：最后一个空格分隔名称和日期
		lastSpace := strings.LastIndex(input, " ")
		if lastSpace == -1 {
			fmt.Println(render.Red("  格式错误，请输入: 考试名称 YYYY-MM-DD"))
			continue
		}
		name := strings.TrimSpace(input[:lastSpace])
		dateStr := strings.TrimSpace(input[lastSpace+1:])

		// 验证日期格式
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			fmt.Println(render.Red("  日期格式错误，请使用 YYYY-MM-DD"))
			continue
		}

		if err := ExamSvc.Add(name, dateStr); err != nil {
			fmt.Printf("  %s %v\n", render.Red("添加失败:"), err)
		} else {
			fmt.Printf("  %s 已添加: %s (%s)\n", render.Green("✅"), name, dateStr)
		}
	}
	fmt.Println()

	// 4. 数据目录确认
	fmt.Println(render.Section("📁 步骤 3/3: 数据存储位置"))
	fmt.Printf("  当前数据目录: %s\n", render.Bold(cfg.DataDir))
	fmt.Println(render.Dim("  如需修改，可设置环境变量 STUDY_DATA_DIR"))
	fmt.Println(render.Dim("  整个目录复制到新电脑即可迁移所有数据。"))
	fmt.Println()

	// 完成
	fmt.Println(render.Green("🎉 设置完成！"))
	fmt.Println()
	fmt.Println(render.Dim("  常用命令:"))
	fmt.Println(render.Dim("    study overview  查看仪表板"))
	fmt.Println(render.Dim("    study log       记录学习进度"))
	fmt.Println(render.Dim("    study diary     写学习日记"))
	fmt.Println(render.Dim("    study --help    查看所有命令"))
	fmt.Println()

	return nil
}
