package service

import (
	"context"
	"fmt"

	"study/internal/auth"
	"study/internal/config"
)

// AllServices 聚合所有服务实例
// GUI 通过此结构体访问所有业务服务，CLI 保持自己的全局变量不变
type AllServices struct {
	Config   *config.Config
	Record   *RecordService
	Exam     *ExamService
	WP       *WeakPointService
	Subj     *SubjectService
	Memo     *MemoService
	Diary    *DiaryService // 可能为 nil（SQLite 初始化失败时）
	Dash     *DashboardService
	Heat     *HeatmapService
	Streak   *StreakService
	Sync     *GitSyncService
	Gen      *CourseworkService
	Drive    *GoogleDriveService    // nil = 未配置
	Calendar *GoogleCalendarService // nil = 未配置
}

// Bootstrap 初始化所有服务（不含 CLI 专属逻辑）
// cfg 必须已调用 EnsureDirs()
// warn 接收非致命警告信息（CLI 输出到 stderr，GUI 输出到日志）
func Bootstrap(cfg *config.Config, warn func(string)) (*AllServices, error) {
	if err := cfg.EnsureDirs(); err != nil {
		warn(fmt.Sprintf("创建数据目录失败: %v", err))
	}

	svc := &AllServices{Config: cfg}

	// 基础服务（不依赖其他服务）
	svc.Record = NewRecordService(cfg)
	svc.Exam = NewExamService(cfg)
	svc.WP = NewWeakPointService(cfg)
	svc.Subj = NewSubjectService(cfg)
	svc.Memo = NewMemoService(cfg)

	// Diary 可能失败 → 降级而非崩溃
	var err error
	svc.Diary, err = NewDiaryService(cfg)
	if err != nil {
		warn(fmt.Sprintf("初始化日记数据库失败: %v\n（日记功能不可用）", err))
		svc.Diary = nil
	}

	// 聚合服务
	svc.Heat = NewHeatmapService(cfg, svc.Record)
	svc.Streak = NewStreakService(cfg, svc.Record)
	svc.Dash = NewDashboardService(cfg, svc.Exam, svc.Subj, svc.WP, svc.Record, svc.Diary)

	// 可选服务（失败不阻断）
	svc.Gen = NewCourseworkService(cfg)
	svc.Sync = NewGitSyncService(cfg)

	// 加载同步配置
	cfg.LoadSyncConfig()

	// Google 服务（可选）
	googleClient, err := auth.NewHTTPClient(context.Background(), config.GoogleScopes())
	if err != nil {
		warn(fmt.Sprintf("Google 服务初始化失败: %v", err))
	} else if googleClient != nil {
		svc.Drive, err = NewGoogleDriveService(cfg, googleClient)
		if err != nil {
			warn(fmt.Sprintf("Google Drive 服务初始化失败: %v", err))
		}
		svc.Calendar, err = NewGoogleCalendarService(cfg, googleClient)
		if err != nil {
			warn(fmt.Sprintf("Google Calendar 服务初始化失败: %v", err))
		}
	}

	return svc, nil
}

// Close 关闭所有需要清理的资源
func (svc *AllServices) Close() {
	if svc.Diary != nil {
		svc.Diary.Close()
	}
	if svc.Drive != nil {
		svc.Drive.Close()
	}
	if svc.Calendar != nil {
		svc.Calendar.Close()
	}
}
