package cli

import (
	"context"
	"fmt"

	"study/internal/auth"
	"study/internal/config"
	"study/internal/render"

	"github.com/spf13/cobra"
)

func newGoogleAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "google",
		Aliases: []string{"g"},
		Short:   "管理 Google 服务认证",
		Long: `管理 Google 服务的 OAuth2 认证。

支持 login（授权）、logout（退出）、status（查看状态）操作。`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "login",
		Aliases: []string{"li"},
		Short:   "登录 Google 账号",
		Long: `启动浏览器完成 Google OAuth2 授权。

首次使用前，请先在 Google Cloud Console 创建 OAuth2 桌面应用凭据，
然后通过 study init 引导向导输入 Client ID 和 Client Secret。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(render.Section("🔐 Google 账号授权"))
			fmt.Println()

			// 检查凭据是否已配置
			clientID, _, err := auth.GetClientIDSecret()
			if err != nil {
				return fmt.Errorf("读取凭据失败: %w", err)
			}
			if clientID == "" {
				fmt.Println(render.Red("❌ 尚未配置 Google 客户端凭据。"))
				fmt.Println(render.Dim("  请先运行 study init 进行配置。"))
				fmt.Println()
				fmt.Println(render.Dim("  需要准备："))
				fmt.Println(render.Dim("  1. 前往 https://console.cloud.google.com/ 创建项目"))
				fmt.Println(render.Dim("  2. 启用 Google Drive API 和 Google Calendar API"))
				fmt.Println(render.Dim("  3. 创建 OAuth 2.0 客户端 ID（桌面应用类型）"))
				fmt.Println(render.Dim("  4. 将 Client ID 和 Client Secret 输入引导向导"))
				return nil
			}

			fmt.Println(render.Dim("即将打开浏览器，请登录 Google 账号并授权..."))
			fmt.Println()

			ctx := context.Background()
			if err := auth.Reauthorize(ctx, config.GoogleScopes()); err != nil {
				return fmt.Errorf("授权失败: %w", err)
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "logout",
		Aliases: []string{"lo"},
		Short:   "退出 Google 账号",
		Long:    "清除本地保存的 Google OAuth2 Token（不会删除已上传的 Drive 文件或 Calendar 事件）。",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := auth.GetAuthInfo()
			if !info.IsAuthorized {
				fmt.Println(render.Dim("当前未登录 Google 账号。"))
				return nil
			}

			if err := auth.ClearToken(); err != nil {
				return fmt.Errorf("清除 Token 失败: %w", err)
			}

			fmt.Println(render.Green("✅ 已退出 Google 账号。"))
			fmt.Println(render.Dim("重新授权请运行 study google login。"))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "查看 Google 认证状态",
		Long:    "显示当前 Google 服务配置状态和授权信息。",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := auth.GetAuthInfo()

			fmt.Println(render.Section("🔗 Google 服务状态"))
			fmt.Println()

			// 配置状态
			if info.IsConfigured {
				fmt.Printf("  %s 客户端凭据: 已配置\n", render.Green("✅"))
			} else {
				fmt.Printf("  %s 客户端凭据: 未配置\n", render.Red("❌"))
				fmt.Println(render.Dim("    运行 study init 配置 Google 集成"))
			}

			// 授权状态
			if info.IsAuthorized {
				fmt.Printf("  %s 授权状态: 已授权\n", render.Green("✅"))
			} else if info.IsConfigured {
				fmt.Printf("  %s 授权状态: 未授权\n", render.Yellow("⚠️"))
				fmt.Println(render.Dim("    运行 study google login 进行授权"))
			}

			fmt.Println()

			// 提示
			if info.IsConfigured && info.IsAuthorized {
				fmt.Println(render.Dim("  Drive 命令: study drive upload/list/status"))
				fmt.Println(render.Dim("  Calendar 命令: study calendar add/list/clear"))
			}

			return nil
		},
	})

	return cmd
}
