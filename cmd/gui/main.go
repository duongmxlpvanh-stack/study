package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"study/internal/config"
	"study/internal/gui"
	"study/internal/service"
)

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化所有业务服务
	svc, err := service.Bootstrap(cfg, func(msg string) {
		log.Printf("[WARN] %s", msg)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	// 3. 创建前端资源（开发时为空，生产构建时 embed）
	var assets gui.FrontendAssets = gui.NoAssets{}

	// 4. 创建 Wails 应用
	app := gui.NewApp(svc, assets)

	// 5. 设置日志文件
	setupLogging(cfg.DataDir)

	// 6. 启动 GUI（阻塞直到退出）
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "GUI 启动失败: %v\n", err)
		os.Exit(1)
	}
}

// setupLogging 将日志输出到数据目录
func setupLogging(dataDir string) {
	logPath := filepath.Join(dataDir, "gui.log")
	_ = logPath
	// 日志默认输出到 stderr，后续可改为文件
	log.SetPrefix("[study-gui] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// 确保未使用的导入不出错
var _ = strings.TrimSpace
