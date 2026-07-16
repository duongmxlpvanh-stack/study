//go:build !windows

package cli

// ensureTerminal 在非 Windows 平台上是空操作。
func ensureTerminal() {}
