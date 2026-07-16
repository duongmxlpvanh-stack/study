package cli

import (
	"context"
	"fmt"
	"os"

	"study/internal/auth"
	"study/internal/config"
	"study/internal/render"
	"study/internal/service"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config

	// 服务实例（全局，各命令共享）
	RecordSvc *service.RecordService
	ExamSvc   *service.ExamService
	WpSvc     *service.WeakPointService
	SubjSvc   *service.SubjectService
	MemoSvc   *service.MemoService
	DashSvc   *service.DashboardService
	HeatSvc   *service.HeatmapService
	StreakSvc *service.StreakService
	DiarySvc  *service.DiaryService

	// Google 服务（nil = 未配置或初始化失败，功能降级可用）
	DriveSvc    *service.GoogleDriveService
	CalendarSvc *service.GoogleCalendarService

	// 全局 rootCmd，REPL 复用
	rootCmd *cobra.Command
)

// Init 初始化所有服务（在程序启动时调用一次）
func Init() {
	// 初始化配置
	cfg = config.Load()

	// 确保数据目录存在（必须在初始化 Diary 之前）
	if err := cfg.EnsureDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 创建数据目录失败: %v\n", err)
	}

	// 初始化服务
	RecordSvc = service.NewRecordService(cfg)
	ExamSvc = service.NewExamService(cfg)
	WpSvc = service.NewWeakPointService(cfg)
	SubjSvc = service.NewSubjectService(cfg)
	MemoSvc = service.NewMemoService(cfg)
	var err error
	DiarySvc, err = service.NewDiaryService(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 初始化日记数据库失败: %v\n（日记功能不可用）\n", err)
	}

	HeatSvc = service.NewHeatmapService(cfg, RecordSvc)
	StreakSvc = service.NewStreakService(cfg, RecordSvc)
	DashSvc = service.NewDashboardService(cfg, ExamSvc, SubjSvc, WpSvc, RecordSvc, DiarySvc)

	// 初始化 Google 服务（可选 — 未配置时静默跳过）
	googleClient, err := auth.NewHTTPClient(context.Background(), config.GoogleScopes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: Google 服务初始化失败: %v\n", err)
	} else if googleClient != nil {
		DriveSvc, err = service.NewGoogleDriveService(cfg, googleClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: Google Drive 服务初始化失败: %v\n", err)
		}
		CalendarSvc, err = service.NewGoogleCalendarService(cfg, googleClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: Google Calendar 服务初始化失败: %v\n", err)
		}
	}

	// 构建命令树
	rootCmd = buildRootCmd()
}

// buildRootCmd 构建命令树（只做命令注册，不做服务初始化）
func buildRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "study",
		Short: "📋 study — 个人学习管理工具",
		Long: render.Title("📋 study 管理中心") + `

一个面向大学生的个人学习管理工具。
打开 Dashboard 就能看清全局，敲一条命令就能完成记录。

设计哲学：把精力留给学习本身，管理交给工具。`,
		// 无子命令时默认进入 REPL 或显示帮助
		Run: func(cmd *cobra.Command, args []string) {
			// 非交互终端（管道/重定向）：显示帮助后退出，避免卡在 REPL
			if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
				cmd.Help()
				return
			}
			// 进入 REPL 交互模式
			RunREPL()
		},
	}

	// 注册所有子命令
	cmd.AddCommand(newLogCmd())
	cmd.AddCommand(newExamCmd())
	cmd.AddCommand(newWeakPointCmd())
	cmd.AddCommand(newSubjectCmd())
	cmd.AddCommand(newDiaryCmd())
	cmd.AddCommand(newMemoCmd())
	cmd.AddCommand(newOverviewCmd())
	cmd.AddCommand(newHeatmapCmd())
	cmd.AddCommand(newStreakCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newGoogleAuthCmd())
	cmd.AddCommand(newDriveCmd())
	cmd.AddCommand(newCalendarCmd())
	cmd.AddCommand(newPathCmd())

	return cmd
}

// GetRootCmd 获取全局 root command（REPL 用）
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// GetConfig 获取全局配置（init 向导用）
func GetConfig() *config.Config {
	return cfg
}

// Execute 运行根命令
func Execute() {
	Init()

	// Windows: 如果从资源管理器双击启动，自动在 Windows Terminal 中重新打开
	ensureTerminal()

	// 设置终端标签页标题
	render.SetTerminalTitle("🕮 study")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// 程序退出前关闭资源
	if DiarySvc != nil {
		DiarySvc.Close()
	}
	if DriveSvc != nil {
		DriveSvc.Close()
	}
	if CalendarSvc != nil {
		CalendarSvc.Close()
	}
}
