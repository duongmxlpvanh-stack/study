//go:build windows

package cli

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modKernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32SnapshotW = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW           = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW            = modKernel32.NewProc("Process32NextW")
)

const th32csSnapprocess = 0x00000002

type processEntry32W struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

// getParentProcessName 返回当前进程的父进程名称（小写），失败时返回空字符串。
func getParentProcessName() string {
	snapshot, _, _ := procCreateToolhelp32SnapshotW.Call(th32csSnapprocess, 0)
	if snapshot == 0 || snapshot == uintptr(syscall.InvalidHandle) {
		return ""
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var pe processEntry32W
	pe.dwSize = uint32(unsafe.Sizeof(pe))

	currentPID := uint32(os.Getpid())
	var parentPID uint32

	// 第一遍：找到当前进程，获取父进程 PID
	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	for ret != 0 {
		if pe.th32ProcessID == currentPID {
			parentPID = pe.th32ParentProcessID
			break
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	}

	if parentPID == 0 {
		return ""
	}

	// 第二遍：找到父进程，获取进程名
	ret, _, _ = procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	for ret != 0 {
		if pe.th32ProcessID == parentPID {
			return strings.ToLower(syscall.UTF16ToString(pe.szExeFile[:]))
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	}

	return ""
}

// ensureTerminal 检测当前运行环境，如果是从资源管理器双击启动的，
// 则在 Windows Terminal 中重新启动以获得更好的 ANSI 支持和现代终端体验。
func ensureTerminal() {
	// 已在 Windows Terminal 中运行
	if os.Getenv("WT_SESSION") != "" {
		return
	}

	// 有命令行参数（如 study overview），说明是从终端手动启动的
	if len(os.Args) > 1 {
		return
	}

	// 仅在父进程是 explorer.exe（双击启动）时才重新启动
	parent := getParentProcessName()
	if parent != "explorer.exe" {
		return
	}

	// 检查 Windows Terminal 是否可用
	wt, err := exec.LookPath("wt.exe")
	if err != nil {
		return // Windows Terminal 未安装，保持当前 CMD 窗口
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}

	// 在 Windows Terminal 中重新启动，退出当前 CMD 窗口
	cmd := exec.Command(wt, exe)
	cmd.Start()
	os.Exit(0)
}
