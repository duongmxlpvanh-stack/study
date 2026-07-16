package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// 简单的 ANSI 颜色和样式
// 后续可升级为 lipgloss

var useColor = true

func init() {
	// 检测是否支持颜色
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		useColor = false
	}
	// stdout 不是终端时（管道/重定向）禁用颜色
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		useColor = false
	}
}

// 颜色码
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

// 热力图专用颜色（绿色系，从浅到深）
var heatColors = []string{
	"\033[48;5;235m", // 0: 无（深灰背景）
	"\033[48;5;22m",  // 1: 浅绿
	"\033[48;5;28m",  // 2: 中绿
	"\033[48;5;34m",  // 3: 深绿
	"\033[48;5;40m",  // 4: 亮绿
}

func color(c string, s string) string {
	if !useColor {
		return s
	}
	return c + s + reset
}

// 公共样式函数
func Bold(s string) string     { return color(bold, s) }
func Dim(s string) string      { return color(dim, s) }
func Red(s string) string      { return color(red, s) }
func Green(s string) string    { return color(green, s) }
func Yellow(s string) string   { return color(yellow, s) }
func Blue(s string) string     { return color(blue, s) }
func Cyan(s string) string     { return color(cyan, s) }
func White(s string) string    { return color(white, s) }

// Title 大标题
func Title(s string) string {
	return Bold(s)
}

// Section 段落标题
func Section(s string) string {
	return Cyan(Bold("▎" + s))
}

// KeyValue 键值对
func KeyValue(k, v string) string {
	return Dim(k+": ") + v
}

// HeatBlock 热力图色块
func HeatBlock(level int) string {
	if !useColor {
		if level == 0 {
			return "·"
		}
		return "█"
	}
	if level < 0 {
		level = 0
	}
	if level > 4 {
		level = 4
	}
	return heatColors[level] + "  " + reset
}

// Divider 分隔线
func Divider() string {
	return Dim(strings.Repeat("─", 60))
}

// ProgressBar 简易进度条
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return Dim("[没有数据]")
	}
	filled := current * width / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s %d/%d", bar, current, total)
}
