//go:build !windows

package cli

import "fmt"

// getBinaryDir 在非 Windows 平台返回空（path check 不用）。
func getBinaryDir() (string, error) {
	return "", fmt.Errorf("当前平台不支持 PATH 管理")
}

// isInPath 在非 Windows 平台返回 false。
func isInPath(binaryDir string) (bool, error) {
	return false, nil
}

// AddToPath 在非 Windows 平台上是空操作。
func AddToPath() (added bool, err error) {
	return false, nil
}
