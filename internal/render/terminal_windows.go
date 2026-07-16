//go:build windows

package render

import (
	"os"

	"golang.org/x/sys/windows"
)

func init() {
	enableWindowsVirtualTerminal()
}

// enableWindowsVirtualTerminal 在 Windows 控制台上启用 ANSI 转义序列处理。
// 非控制台句柄时静默跳过（重定向 / ConPTY 本身已支持 ANSI）。
func enableWindowsVirtualTerminal() {
	stdout := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(stdout, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	windows.SetConsoleMode(stdout, mode)
}
