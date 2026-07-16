//go:build windows

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	modUser32                       = syscall.NewLazyDLL("user32.dll")
	procSendMessageTimeoutW         = modUser32.NewProc("SendMessageTimeoutW")
)

const (
	hwndBroadcast   = uintptr(0xFFFF)
	wmSettingChange = uintptr(0x001A)
	smtoAbortIfHung = uintptr(0x0002)
)

// getBinaryDir 返回当前 exe 所在的目录绝对路径。
func getBinaryDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序路径失败: %w", err)
	}
	return filepath.Dir(exe), nil
}

// isInPath 检查指定目录是否已在用户级 PATH 环境变量中。
// 大小写不敏感，并会去除末尾反斜杠后比较。
func isInPath(binaryDir string) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()

	pathVal, _, err := k.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil // 值不存在，视为不在 PATH 中
		}
		return false, fmt.Errorf("读取 PATH 失败: %w", err)
	}

	normalizedDir := strings.TrimRight(binaryDir, `\`)
	for _, segment := range strings.Split(pathVal, ";") {
		normalizedSegment := strings.TrimRight(strings.TrimSpace(segment), `\`)
		if strings.EqualFold(normalizedDir, normalizedSegment) {
			return true, nil
		}
	}
	return false, nil
}

// addToUserPath 将指定目录追加到用户级 PATH 环境变量。
func addToUserPath(binaryDir string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()

	existingPath, _, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("读取 PATH 失败: %w", err)
	}

	var newPath string
	if existingPath == "" {
		newPath = binaryDir
	} else {
		newPath = strings.TrimRight(existingPath, ";") + ";" + binaryDir
	}

	if err := k.SetExpandStringValue("Path", newPath); err != nil {
		return fmt.Errorf("写入 PATH 失败: %w", err)
	}
	return nil
}

// broadcastEnvChange 通知所有运行中的应用程序环境变量已变更。
// 这是一个 best-effort 操作，失败不会返回错误。
func broadcastEnvChange() {
	envStr, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	procSendMessageTimeoutW.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(envStr)),
		smtoAbortIfHung,
		5000,
		0,
	)
}

// AddToPath 将 study.exe 所在目录添加到用户级 PATH 环境变量。
// 返回 added=true 表示已添加，added=false 表示已在 PATH 中无需操作。
func AddToPath() (added bool, err error) {
	binaryDir, err := getBinaryDir()
	if err != nil {
		return false, err
	}

	inPath, err := isInPath(binaryDir)
	if err != nil {
		return false, err
	}
	if inPath {
		return false, nil // 已在 PATH 中
	}

	if err := addToUserPath(binaryDir); err != nil {
		return false, err
	}

	broadcastEnvChange()
	return true, nil
}
