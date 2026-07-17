//go:build !windows

package gui

import (
	"fmt"
	"os"
	"syscall"
)

// acquireSingletonLock Unix 单实例检测（flock 排他锁）
func acquireSingletonLock(lockPath string) error {
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("无法打开锁文件: %w", err)
	}

	// 尝试获取排他锁（非阻塞）
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return fmt.Errorf("另一个 study-gui 实例已在运行")
	}

	// 写入 PID
	f.Truncate(0)
	f.Seek(0, 0)
	fmt.Fprintf(f, "%d", os.Getpid())

	return nil
}

// releaseSingletonLock 释放单实例锁
func releaseSingletonLock(lockPath string) {
	// flock 在文件关闭时自动释放
	os.Remove(lockPath)
}
