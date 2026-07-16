package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

// 终端样式 — 基于 lipgloss
// 后续可扩展主题、边框、间距等高级特性

// 是否启用颜色（管道/重定向时自动禁用）
var useColor = true

func init() {
	// NO_COLOR 和 TERM=dumb 标准
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		useColor = false
	}
	// stdout 不是终端时（管道/重定向）禁用颜色
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		useColor = false
	}

	if !useColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

// 预定义的 lipgloss 样式（零分配，线程安全）
var (
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	styleCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleWhite  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

// 公共样式函数 — 保持原有 API 不变
func Bold(s string) string   { return styleBold.Render(s) }
func Dim(s string) string    { return styleDim.Render(s) }
func Red(s string) string    { return styleRed.Render(s) }
func Green(s string) string  { return styleGreen.Render(s) }
func Yellow(s string) string { return styleYellow.Render(s) }
func Blue(s string) string   { return styleBlue.Render(s) }
func Cyan(s string) string   { return styleCyan.Render(s) }
func White(s string) string  { return styleWhite.Render(s) }

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

// 热力图色块颜色（5级绿色渐变，使用 256 色调色板以获得最佳兼容性）
var heatStyles = []lipgloss.Style{
	lipgloss.NewStyle().Background(lipgloss.Color("235")), // 0: 深灰（无贡献）
	lipgloss.NewStyle().Background(lipgloss.Color("22")),  // 1: 深绿
	lipgloss.NewStyle().Background(lipgloss.Color("28")),  // 2: 中绿
	lipgloss.NewStyle().Background(lipgloss.Color("34")),  // 3: 亮绿
	lipgloss.NewStyle().Background(lipgloss.Color("40")),  // 4: 鲜绿
}

// HeatBlock 热力图色块
func HeatBlock(level int) string {
	if !useColor {
		if level <= 0 {
			return "·"
		}
		return "█"
	}
	if level < 0 {
		return "  "
	}
	if level > 4 {
		level = 4
	}
	return heatStyles[level].Render("  ")
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

// 以下为 lipgloss 增强的布局辅助函数

// Card 渲染一个带圆角边框的卡片
func Card(title, content string, width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	contentStyle := lipgloss.NewStyle().
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width)

	if title != "" {
		return cardStyle.Render(titleStyle.Render(title) + "\n" + contentStyle.Render(content))
	}
	return cardStyle.Render(contentStyle.Render(content))
}

// StatCard 统计数字小卡片
func StatCard(label, value string) string {
	labelStyle := lipgloss.NewStyle().Faint(true).Align(lipgloss.Center).Width(14)
	valueStyle := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Width(14)

	card := lipgloss.JoinVertical(lipgloss.Top,
		labelStyle.Render(label),
		valueStyle.Render(value),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(card)
}

// WarningBorder 警告边框（红色高亮）
func WarningBorder(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(0, 1).
		Render(content)
}
