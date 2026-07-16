package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"study/internal/auth"
	"study/internal/render"

	"github.com/spf13/cobra"
)

func newDriveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drive",
		Aliases: []string{"dr"},
		Short:   "管理 Google Drive 文件上传",
		Long: `上传文件到 Google Drive 并管理已上传的文件。

文件存放在 study_pdfs/<科目>/ 目录下。`,
	}

	// drive upload
	uploadCmd := &cobra.Command{
		Use:     "upload",
		Aliases: []string{"up"},
		Short:   "上传文件到 Google Drive",
		Long: `将本地文件上传到 Google Drive。

文件会按科目分类存放在 study_pdfs/<科目>/ 目录下。
支持 PDF、图片、文档等常见格式。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if DriveSvc == nil {
				return fmt.Errorf("Google Drive 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			localPath := args[0]

			// 验证文件存在
			absPath, err := filepath.Abs(localPath)
			if err != nil {
				return fmt.Errorf("解析文件路径失败: %w", err)
			}
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("文件不存在: %s", localPath)
			}

			subject, _ := cmd.Flags().GetString("subject")
			if subject == "" {
				// 尝试从文件名推断科目...
				fmt.Println(render.Yellow("⚠️ 未指定科目，文件将上传到「未分类」文件夹。"))
				fmt.Println(render.Dim("  使用 --subject 或 -s 指定科目，例如: --subject 高等数学"))
			}

			ctx := context.Background()
			fmt.Printf("正在上传 %s 到 Google Drive...\n", filepath.Base(absPath))
			fileID, err := DriveSvc.UploadFile(ctx, absPath, subject)
			if err != nil {
				return fmt.Errorf("上传失败: %w", err)
			}

			fmt.Printf("%s 上传成功！(Drive ID: %s)\n", render.Green("✅"), fileID)
			return nil
		},
	}
	uploadCmd.Flags().StringP("subject", "s", "", "所属科目（如未指定则归类为「未分类」）")
	cmd.AddCommand(uploadCmd)

	// drive list
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出已上传的文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			if DriveSvc == nil {
				return fmt.Errorf("Google Drive 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			subject, _ := cmd.Flags().GetString("subject")

			ctx := context.Background()
			files, err := DriveSvc.ListFiles(ctx, subject)
			if err != nil {
				return fmt.Errorf("列出文件失败: %w", err)
			}

			fmt.Println(render.Section("📁 Google Drive 文件"))
			if subject != "" {
				fmt.Printf(render.Dim("  科目: %s\n"), subject)
			}
			fmt.Println()

			if len(files) == 0 {
				fmt.Println(render.Dim("  还没有上传过文件。"))
				fmt.Println(render.Dim("  使用 study drive upload <文件路径> --subject <科目> 上传。"))
				return nil
			}

			for i, f := range files {
				subjectLabel := ""
				if f.Description != "" {
					subjectLabel = fmt.Sprintf(" [%s]", f.Description)
				}
				fmt.Printf("  %d. %s%s\n", i+1, render.Bold(f.Name), render.Dim(subjectLabel))
				fmt.Printf("     %s | %s\n",
					render.Dim(formatFileSize(f.Size)),
					render.Dim(f.CreatedTime),
				)
			}
			fmt.Println()
			fmt.Printf(render.Dim("  共 %d 个文件\n"), len(files))
			return nil
		},
	}
	listCmd.Flags().StringP("subject", "s", "", "按科目筛选（可选，不指定则列出全部）")
	cmd.AddCommand(listCmd)

	// drive auto-upload
	cmd.AddCommand(&cobra.Command{
		Use:   "auto-upload [on|off]",
		Short: "开关自动上传功能",
		Long: `开启后，每次生成 PDF 时自动上传到 Google Drive。
状态仅在当前会话有效，重启后恢复为关闭。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if DriveSvc == nil {
				return fmt.Errorf("Google Drive 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			if len(args) == 0 {
				// 显示当前状态
				if DriveSvc.IsAutoUpload() {
					fmt.Printf("自动上传: %s\n", render.Green("已开启"))
				} else {
					fmt.Printf("自动上传: %s\n", render.Dim("已关闭"))
					fmt.Println(render.Dim("使用 study drive auto-upload on 开启"))
				}
				return nil
			}

			switch args[0] {
			case "on":
				DriveSvc.ToggleAutoUpload(true)
				fmt.Printf("%s 自动上传已开启\n", render.Green("✅"))
			case "off":
				DriveSvc.ToggleAutoUpload(false)
				fmt.Printf("%s 自动上传已关闭\n", render.Dim("🔒"))
			default:
				return fmt.Errorf("无效参数: %s，请使用 on 或 off", args[0])
			}
			return nil
		},
	})

	// drive status
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "查看 Drive 认证和存储状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			if DriveSvc == nil {
				return fmt.Errorf("Google Drive 服务未初始化，请先运行 study init 配置 Google 集成")
			}

			info := auth.GetAuthInfo()

			fmt.Println(render.Section("📁 Google Drive 状态"))
			fmt.Println()

			if info.IsAuthorized {
				fmt.Printf("  %s 认证状态: 已授权\n", render.Green("✅"))
			} else {
				fmt.Printf("  %s 认证状态: 未授权\n", render.Red("❌"))
				fmt.Println(render.Dim("    运行 study google login 进行授权"))
				return nil
			}

			// 自动上传状态
			if DriveSvc.IsAutoUpload() {
				fmt.Printf("  %s 自动上传: 已开启\n", render.Green("📤"))
			} else {
				fmt.Printf("  %s 自动上传: 已关闭\n", render.Dim("📤"))
			}

			// 存储统计
			ctx := context.Background()
			storage, err := DriveSvc.GetStorageInfo(ctx)
			if err != nil {
				fmt.Printf("  %s 存储统计: 获取失败 (%v)\n", render.Yellow("⚠️"), err)
			} else {
				fmt.Printf("  📊 文件总数: %d\n", storage.TotalFiles)
				fmt.Printf("  📂 科目文件夹: %d 个\n", len(storage.Folders))
				if len(storage.Folders) > 0 {
					for _, folder := range storage.Folders {
						fmt.Printf("     - %s\n", folder)
					}
				}
			}

			fmt.Println()
			return nil
		},
	})

	return cmd
}

// formatFileSize 将文件大小（字节）格式化为可读字符串。
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
