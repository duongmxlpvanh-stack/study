package main

import (
	"fmt"
	"log"
	"net/http"
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

	// 3. 设置日志
	setupLogging(cfg.DataDir)

	// 4. 定位前端资源目录
	frontendDir := findFrontendDir()
	log.Printf("[GUI] 前端目录: %s", frontendDir)

	// 5. 创建前端资源 HTTP 处理器
	assetsHandler := http.FileServer(http.Dir(frontendDir))

	// 6. 创建 Wails 应用
	app := gui.NewApp(svc, assetsHandler)

	// 7. 启动 GUI（阻塞直到退出）
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "GUI 启动失败: %v\n", err)
		os.Exit(1)
	}
}

// findFrontendDir 查找前端资源目录
// 优先级: 1) 可执行文件同级的 frontend/  2) 当前工作目录的 frontend/
func findFrontendDir() string {
	// 1. 可执行文件所在目录
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exePath), "frontend")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// 2. 当前工作目录
	if cwd, err := os.Getwd(); err == nil {
		dir := filepath.Join(cwd, "frontend")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// 3. 降级：返回当前目录下的 frontend
	return "frontend"
}

// setupLogging 将日志输出到数据目录
func setupLogging(dataDir string) {
	logPath := filepath.Join(dataDir, "gui.log")
	_ = logPath
	log.SetPrefix("[study-gui] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
