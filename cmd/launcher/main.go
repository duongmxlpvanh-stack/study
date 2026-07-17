//go:build windows

// 启动器：无控制台窗口，双击直接打开 Windows Terminal 运行 study.exe
//
// 编译: go build -ldflags="-H windowsgui -s -w" -o 启动study.exe ./cmd/launcher/
//
// -H windowsgui: 不分配控制台，零闪窗
// 用户双击 启动study.exe → Windows Terminal 直接打开，体验流畅

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

func main() {
	// 1. 定位 study.exe（与启动器同目录）
	exePath, err := os.Executable()
	if err != nil {
		showError("无法获取启动器路径", err)
		return
	}

	dir := filepath.Dir(exePath)
	studyExe := filepath.Join(dir, "study.exe")

	// 2. 检查 study.exe 是否存在
	if _, err := os.Stat(studyExe); os.IsNotExist(err) {
		showError("未找到 study.exe\n\n请确保 启动study.exe 与 study.exe 在同一目录下", nil)
		return
	}

	// 3. 优先使用 Windows Terminal（wt.exe）
	wt, err := exec.LookPath("wt.exe")
	if err == nil {
		cmd := exec.Command(wt, studyExe)
		cmd.Start()
		return
	}

	// 4. 降级：使用 cmd.exe 的 start 命令
	cmd := exec.Command("cmd", "/c", "start", "", studyExe)
	cmd.Start()
}

// showError 弹出 Windows 错误对话框（无控制台时的唯一交互方式）
func showError(msg string, err error) {
	text := msg
	if err != nil {
		text += "\n" + err.Error()
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString("study 启动器")
	body, _ := syscall.UTF16PtrFromString(text)
	msgBox.Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x00000010) // MB_ICONERROR
}
