package gui

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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
func NewApp(svc *service.AllServices) *App {
	handler := NewHandler(svc)

	opts := application.Options{
		Name:        "🕮 study管理中心",
		Description: "个人学习管理工具 — Dashboard 看清全局，敲命令完成记录",
		Services: []application.Service{
			application.NewService(handler),
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	}

	return &App{
		svc: svc,
		app: application.New(opts),
	}
}

// Run 启动 GUI 应用（阻塞直到退出）
func (a *App) Run() error {
	// 单实例检测
	if err := a.acquireLock(); err != nil {
		log.Printf("[GUI] 已有实例在运行: %v", err)
		return nil
	}
	defer a.releaseLock()

	// 创建主窗口
	a.window = a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      "main",
		Title:     "🕮 study管理中心",
		Width:     960,
		Height:    680,
		MinWidth:  720,
		MinHeight: 480,
		URL:       "/",
	})

	// 设置应用图标
	iconPNG := GenerateTrayIcon()
	if len(iconPNG) > 0 {
		a.app.SetIcon(iconPNG)
	}

	// 拦截窗口关闭事件 → 隐藏到托盘
	a.window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		a.window.Hide()
		e.Cancel()
	})

	// 设置系统托盘
	a.setupSystray()

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("[GUI] 收到退出信号")
		a.Quit()
	}()

	log.Println("[GUI] study管理中心 启动成功")
	log.Printf("[GUI] 数据目录: %s", a.svc.Config.DataDir)

	return a.app.Run()
}

// Quit 完全退出应用
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

	// 设置托盘图标
	iconPNG := GenerateTrayIcon()
	if len(iconPNG) > 0 {
		a.tray.SetIcon(iconPNG)
	}

	if runtime.GOOS == "darwin" {
		a.tray.SetTemplateIcon(iconPNG)
	}

	a.tray.SetTooltip("study管理中心")

	// 单击托盘图标 → 切换窗口
	a.tray.OnClick(func() {
		a.ToggleWindow()
	})

	// 右键菜单
	menu := a.app.NewMenu()
	menu.Add("📋 显示/隐藏").OnClick(func(ctx *application.Context) {
		a.ToggleWindow()
	})
	menu.AddSeparator()

	// 开机自启（使用 Wails 内置 Autostart 管理器）
	if a.app.Autostart != nil {
		autoStartLabel := "☐ 开机自启"
		if enabled, err := a.app.Autostart.IsEnabled(); err == nil && enabled {
			autoStartLabel = "✅ 开机自启"
		}
		autoStartItem := menu.Add(autoStartLabel)
		autoStartItem.OnClick(func(ctx *application.Context) {
			if a.app.Autostart == nil {
				return
			}
			if enabled, err := a.app.Autostart.IsEnabled(); err == nil && enabled {
				if err := a.app.Autostart.Disable(); err != nil {
					log.Printf("[GUI] 禁用开机自启失败: %v", err)
				} else {
					log.Println("[GUI] 已禁用开机自启")
				}
			} else {
				if err := a.app.Autostart.Enable(); err != nil {
					log.Printf("[GUI] 启用开机自启失败: %v", err)
				} else {
					log.Println("[GUI] 已启用开机自启")
				}
			}
		})
	}

	menu.AddSeparator()
	menu.Add("🚪 退出").OnClick(func(ctx *application.Context) {
		a.Quit()
	})

	a.tray.SetMenu(menu)
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
		a.window.Focus()
	}
}

// ShowWindow 显示窗口
func (a *App) ShowWindow() {
	if a.window != nil {
		a.window.Show()
		a.window.Focus()
	}
}

// ==================== 单实例检测（Windows 命名 Mutex） ====================

func (a *App) acquireLock() error {
	lockPath := a.svc.Config.DataDir + "\\.gui.lock"
	return acquireSingletonLock(lockPath)
}

func (a *App) releaseLock() {
	lockPath := a.svc.Config.DataDir + "\\.gui.lock"
	releaseSingletonLock(lockPath)
}

// FrontendAssets 别名
type FrontendAssets = struct{}

// NoAssets 空前端资源
type NoAssets = struct{}
