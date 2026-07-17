package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	// 3. 创建 Wails 应用（前端开发时由 wails3 dev 处理）
	app := gui.NewApp(svc)

	// 4. 设置日志
	setupLogging(cfg.DataDir)

	// 5. 启动 GUI（阻塞直到退出）
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "GUI 启动失败: %v\n", err)
		os.Exit(1)
	}
}

// setupLogging 将日志输出到数据目录
func setupLogging(dataDir string) {
	logPath := filepath.Join(dataDir, "gui.log")
	_ = logPath
	log.SetPrefix("[study-gui] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
