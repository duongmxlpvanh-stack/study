//go:build windows

package gui

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

var lockFileHandle *os.File

// acquireSingletonLock Windows 单实例检测
// 通过创建锁文件 + PID 记录实现
func acquireSingletonLock(lockPath string) error {
	// 尝试读取已有锁文件
	if data, err := os.ReadFile(lockPath); err == nil {
		// 锁文件存在，检查进程是否仍在运行
		if pid, parseErr := strconv.Atoi(string(data)); parseErr == nil && pid > 0 {
			// Windows 上 FindProcess 总是成功，通过 OpenProcess 实际检查
			handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
			if err == nil {
				syscall.CloseHandle(handle)
				return fmt.Errorf("另一个 study-gui 实例已在运行 (PID: %d)", pid)
			}
		}
		// 进程已不存在，删除旧锁文件
		os.Remove(lockPath)
	}

	// 创建锁文件，写入当前 PID
	pid := os.Getpid()
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("无法创建锁文件: %w", err)
	}

	return nil
}

// releaseSingletonLock 释放单实例锁
func releaseSingletonLock(lockPath string) {
	os.Remove(lockPath)
}
