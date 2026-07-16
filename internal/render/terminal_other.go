//go:build !windows

package render

// init 在非 Windows 平台上无需操作（ANSI 转义序列天然支持）。
func init() {}
