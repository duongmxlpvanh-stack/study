package cli

import (
	"fmt"

	"study/internal/render"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"sy"},
		Short:   "管理 GitHub 云同步",
		Long: `将学习数据（Markdown 文件）同步到 GitHub 私有仓库。

日记数据库 (diary.db) 和资料文件夹 (materials/) 不会上传。
每次成功的写操作后会自动触发后台同步。`,
	}

	// study sync status
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "查看同步状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			if SyncSvc == nil {
				return fmt.Errorf("同步服务未初始化")
			}
			st := SyncSvc.Status()
			fmt.Println(render.Section("☁️  GitHub 同步状态"))
			fmt.Println()

			if !st.HasGit {
				fmt.Println(render.Dim("  Git 未安装，云端同步不可用。"))
				fmt.Println(render.Dim("  请安装 Git for Windows: https://git-scm.com/download/win"))
				return nil
			}

			if !st.Configured {
				fmt.Println(render.Dim("  未配置云端同步。"))
				fmt.Println(render.Dim("  使用 study sync setup 设置，或在 study init 中配置。"))
				return nil
			}

			fmt.Printf("  %s 同步状态: 已配置\n", render.Green("✅"))
			if st.RemoteURL != "" {
				fmt.Printf("  📡 远程仓库: %s\n", st.RemoteURL)
			}
			if st.LastSync != "" {
				fmt.Printf("  🕐 上次提交: %s\n", st.LastSync)
			}
			if st.PendingChanges > 0 {
				fmt.Printf("  📝 待同步变更: %s 个文件\n", render.Yellow(fmt.Sprint(st.PendingChanges)))
			} else {
				fmt.Println(render.Dim("  📝 无待同步变更"))
			}
			return nil
		},
	})

	// study sync push
	cmd.AddCommand(&cobra.Command{
		Use:   "push",
		Short: "手动推送到 GitHub",
		Long:  "暂存所有本地变更，提交并推送到远程 GitHub 仓库。",
		RunE: func(cmd *cobra.Command, args []string) error {
			if SyncSvc == nil || !SyncSvc.IsEnabled() {
				return fmt.Errorf("云端同步未配置。使用 study sync setup 设置")
			}
			fmt.Print(render.Dim("正在同步到 GitHub..."))
			output, err := SyncSvc.ManualSync()
			if err != nil {
				fmt.Println()
				return err
			}
			fmt.Println(render.Green(" ✅"))
			fmt.Print(output)
			return nil
		},
	})

	// study sync setup
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "设置或重新配置 GitHub 云同步",
		Long: `配置 GitHub 私有仓库用于云端同步。

需要:
  1. 一个 GitHub 私有仓库 URL（如 https://github.com/user/study-data）
  2. 一个 Personal Access Token（需要 'repo' 权限）

Token 获取方式:
  前往 https://github.com/settings/tokens → Generate new token (classic)
  → 勾选 'repo' → 生成后复制 Token`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if SyncSvc == nil {
				return fmt.Errorf("同步服务未初始化")
			}
			repoURL, _ := cmd.Flags().GetString("repo")
			token, _ := cmd.Flags().GetString("token")

			if repoURL == "" || token == "" {
				fmt.Println(render.Section("🔗 配置 GitHub 云同步"))
				fmt.Println()
				fmt.Println(render.Dim("  需要提供:"))
				fmt.Println(render.Dim("  1. GitHub 私有仓库 URL"))
				fmt.Println(render.Dim("  2. Personal Access Token（repo 权限）"))
				fmt.Println()
				fmt.Println(render.Dim("  示例:"))
				fmt.Println(render.Dim("  study sync setup --repo https://github.com/user/study-data --token ghp_xxxx"))
				return nil
			}

			if err := SyncSvc.Setup(repoURL, token); err != nil {
				return fmt.Errorf("设置失败: %w", err)
			}
			fmt.Println(render.Green("✅ GitHub 云同步已配置！"))
			fmt.Println(render.Dim("  每次写操作后将自动推送至远程仓库。"))
			return nil
		},
	}
	setupCmd.Flags().StringP("repo", "r", "", "GitHub 仓库 URL")
	setupCmd.Flags().StringP("token", "t", "", "GitHub Personal Access Token")
	cmd.AddCommand(setupCmd)

	// study sync disable
	cmd.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "禁用云端同步",
		Long:  "禁用自动同步（不删除 Git 仓库和远程配置）。",
		RunE: func(cmd *cobra.Command, args []string) error {
			if SyncSvc == nil {
				return fmt.Errorf("同步服务未初始化")
			}
			if err := SyncSvc.Disable(); err != nil {
				return err
			}
			fmt.Println(render.Green("✅ 云端同步已禁用"))
			fmt.Println(render.Dim("  使用 study sync setup 重新启用。"))
			return nil
		},
	})

	return cmd
}

// afterWrite 在写操作成功后触发后台 Git 同步
// 使用示例: afterWrite("log: %s", input)
func afterWrite(format string, args ...any) {
	if SyncSvc == nil || !SyncSvc.IsEnabled() {
		return
	}
	SyncSvc.AutoSync(fmt.Sprintf(format, args...))
}
