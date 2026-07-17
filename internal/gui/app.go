package gui

import (
	"log"
	"runtime"

	"study/internal/service"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// App 是 Wails GUI 应用的完整包装
type App struct {
	app    *application.App
	svc    *service.AllServices
	tray   *application.SystemTray
	window *application.WebviewWindow
}

// NewApp 创建 GUI 应用实例
// frontendFS: 前端静态资源（开发时传 nil，wails3 dev 会处理）
func NewApp(svc *service.AllServices, frontendFS FrontendAssets) *App {
	handler := NewHandler(svc)

	opts := application.Options{
		Name:        "🕮 study管理中心",
		Description: "个人学习管理工具 — Dashboard 看清全局，敲命令完成记录",
		Services: []application.Service{
			application.NewService(handler),
		},
		Windows: application.WindowsOptions{
			// 关键：关闭窗口 = 隐藏到托盘，不退出
			DisableQuitOnLastWindowClosed: true,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	}

	// TODO: 阶段4 — 通过 embed.FS 嵌入生产构建的前端资源
	// 开发时 wails3 dev 自动处理前端热重载，此处留空
	_ = frontendFS

	guiApp := &App{
		svc: svc,
		app: application.New(opts),
	}

	return guiApp
}

// Run 启动 GUI 应用（阻塞直到退出）
func (a *App) Run() error {
	// 创建主窗口
	a.window = a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "🕮 study管理中心",
		Width:     960,
		Height:    680,
		MinWidth:  720,
		MinHeight: 480,
		URL:       "/",
	})

	// 拦截窗口关闭事件 → 隐藏到托盘
	a.window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		a.window.Hide()
		e.Cancel() // 阻止默认的销毁行为
	})

	// 设置系统托盘
	a.setupSystray()

	// 日志输出
	log.Println("[GUI] study管理中心 启动成功")
	log.Printf("[GUI] 数据目录: %s", a.svc.Config.DataDir)

	return a.app.Run()
}

// Quit 完全退出应用（从托盘菜单调用）
func (a *App) Quit() {
	if a.svc != nil {
		a.svc.Close()
	}
	if a.app != nil {
		a.app.Quit()
	}
}

// setupSystray 创建系统托盘图标和菜单
func (a *App) setupSystray() {
	a.tray = a.app.SystemTray.New()

	// 托盘图标（使用内嵌 PNG）
	// if runtime.GOOS == "darwin" {
	// 	a.tray.SetTemplateIcon(icons.SystrayMacTemplate)
	// }
	_ = runtime.GOOS // 后续添加图标

	a.tray.SetTooltip("study管理中心")

	// 单击托盘图标 → 切换窗口显示/隐藏
	a.tray.OnClick(func() {
		a.ToggleWindow()
	})

	// 托盘右键菜单
	menu := a.app.NewMenu()
	menu.Add("📋 显示/隐藏").OnClick(func(ctx *application.Context) {
		a.ToggleWindow()
	})
	menu.AddSeparator()
	menu.Add("🚪 退出").OnClick(func(ctx *application.Context) {
		a.Quit()
	})

	a.tray.SetMenu(menu)

	// 绑定浮窗到托盘图标
	a.tray.AttachWindow(a.window).WindowOffset(5)
}

// ToggleWindow 切换窗口可见性
func (a *App) ToggleWindow() {
	if a.window == nil {
		return
	}
	if a.window.IsVisible() {
		a.window.Hide()
	} else {
		a.window.Show()
	}
}

// ShowWindow 显示窗口
func (a *App) ShowWindow() {
	if a.window != nil {
		a.window.Show()
	}
}

// ==================== 前端资源抽象 ====================

// FrontendAssets 前端资源接口
// 阶段4实现：生产构建时通过 embed.FS 传入
type FrontendAssets interface {
	HasAssets() bool
}

// NoAssets 空前端资源（开发模式用，wails3 dev 自动处理热重载）
type NoAssets struct{}

func (n NoAssets) HasAssets() bool { return false }
