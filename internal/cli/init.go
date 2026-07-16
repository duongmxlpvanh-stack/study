package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"study/internal/auth"
	"study/internal/config"
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

	render.Typewriter(render.Title("🚀 欢迎使用 study 管理中心！"), 25*time.Millisecond)
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
	fmt.Println(render.Section("📚 步骤 1/4: 添加本学期课程"))
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
	fmt.Println(render.Section("⏰ 步骤 2/4: 添加考试日期"))
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

	// 3. 配置 Google 集成（可选）
	fmt.Println(render.Section("🔗 步骤 3/4: 连接 Google 服务（可选）"))
	fmt.Println(render.Dim("  连接后可上传文件到 Google Drive 并同步学习计划到 Google Calendar。"))
	fmt.Println(render.Dim("  需要 Google Cloud Console 中的 OAuth 2.0 客户端凭据（桌面应用类型）。"))
	fmt.Println(render.Dim("  输入空行跳过此步骤，之后可随时用 study google login 配置。"))
	fmt.Println()

	fmt.Printf("  是否配置 Google 集成？(y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		fmt.Println()
		fmt.Println(render.Dim("  请按以下步骤获取凭据："))
		fmt.Println(render.Dim("  1. 前往 https://console.cloud.google.com/"))
		fmt.Println(render.Dim("  2. 创建项目 → 启用 Drive API + Calendar API"))
		fmt.Println(render.Dim("  3. API 和服务 → 凭据 → 创建 OAuth 2.0 客户端 ID"))
		fmt.Println(render.Dim("  4. 应用类型选择「桌面应用」"))
		fmt.Println()

		fmt.Printf("  Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		if clientID != "" {
			fmt.Printf("  Client Secret: ")
			clientSecret, _ := reader.ReadString('\n')
			clientSecret = strings.TrimSpace(clientSecret)

			if clientSecret != "" {
				if err := auth.SaveClientIDSecret(clientID, clientSecret); err != nil {
					fmt.Printf("  %s 保存凭据失败: %v\n", render.Red("❌"), err)
				} else {
					fmt.Printf("  %s 凭据已安全保存到 Windows 凭据管理器\n", render.Green("✅"))

					// 立即发起 OAuth 授权
					fmt.Println()
					fmt.Println(render.Dim("  正在启动浏览器进行授权..."))
					ctx := context.Background()
					_, err := auth.NewHTTPClient(ctx, config.GoogleScopes())
					if err != nil {
						fmt.Printf("  %s 授权失败: %v\n", render.Red("❌"), err)
						fmt.Println(render.Dim("  可稍后使用 study google login 重新授权"))
					} else {
						fmt.Printf("  %s Google 授权成功！\n", render.Green("✅"))
						fmt.Println(render.Dim("  提示: Google 服务将在下次启动 study 时可用"))
					}
				}
			}
		}
	} else {
		fmt.Println(render.Dim("  已跳过 Google 集成。可稍后使用 study google login 配置。"))
	}
	fmt.Println()

	// 4. 数据目录确认
	fmt.Println(render.Section("📁 步骤 4/4: 数据存储位置"))
	fmt.Printf("  当前数据目录: %s\n", render.Bold(cfg.DataDir))
	fmt.Println(render.Dim("  如需修改，可设置环境变量 STUDY_DATA_DIR"))
	fmt.Println(render.Dim("  整个目录复制到新电脑即可迁移所有数据。"))
	fmt.Println()

	// 附加步骤：添加到系统 PATH
	fmt.Println(render.Section("⚡ 附加步骤: 添加到系统 PATH"))
	fmt.Println(render.Dim("  将 study.exe 所在目录添加到用户 PATH 后，"))
	fmt.Println(render.Dim("  你可以在任意终端中直接输入 study 来启动。"))
	fmt.Println()

	fmt.Printf("  是否添加到 PATH？(y/N): ")
	addPathAnswer, _ := reader.ReadString('\n')
	addPathAnswer = strings.TrimSpace(strings.ToLower(addPathAnswer))

	if addPathAnswer == "y" || addPathAnswer == "yes" {
		added, err := AddToPath()
		if err != nil {
			fmt.Printf("  %s 添加 PATH 失败: %v\n", render.Red("❌"), err)
		} else if !added {
			fmt.Printf("  %s study 已在 PATH 中，无需重复添加\n", render.Green("✅"))
		} else {
			fmt.Printf("  %s 已添加到用户 PATH\n", render.Green("✅"))
			fmt.Println(render.Dim("  请重新打开终端以使 PATH 生效。"))
		}
	} else {
		fmt.Println(render.Dim("  已跳过。可稍后手动将 study.exe 目录添加到 PATH。"))
	}
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
