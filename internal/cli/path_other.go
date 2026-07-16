//go:build !windows

package cli

// AddToPath 在非 Windows 平台上是空操作。
func AddToPath() (added bool, err error) {
	return false, nil
}
