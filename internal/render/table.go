package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table 简单的终端表格渲染器。
// 支持表头加粗、自动列宽、斑马条纹、列对齐。
type Table struct {
	headers    []string
	rows       [][]string
	alignRight []int // 右对齐列索引集合
}

// NewTable 创建一个表格
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
	}
}

// AddRow 添加一行数据
func (t *Table) AddRow(cols ...string) *Table {
	t.rows = append(t.rows, cols)
	return t
}

// SetAlignRight 设置右对齐的列索引
func (t *Table) SetAlignRight(cols ...int) *Table {
	t.alignRight = cols
	return t
}

func (t *Table) isAlignRight(col int) bool {
	for _, c := range t.alignRight {
		if c == col {
			return true
		}
	}
	return false
}

// Render 渲染表格
func (t *Table) Render() string {
	colCount := len(t.headers)
	if colCount == 0 {
		return ""
	}

	// 计算每列最大宽度
	widths := make([]int, colCount)
	for i, h := range t.headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i >= colCount {
				break
			}
			w := displayWidth(cell)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	// 总宽度（列宽 + 3 空格间距）
	var sb strings.Builder

	// 表头
	headerStyle := lipgloss.NewStyle().Bold(true)
	sb.WriteString("  ")
	for i, h := range t.headers {
		cell := padCell(h, widths[i], t.isAlignRight(i))
		sb.WriteString(headerStyle.Render(cell))
		if i < colCount-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// 表头分隔线
	sb.WriteString(Dim("  " + strings.Repeat("─", totalWidth(widths, colCount))))
	sb.WriteString("\n")

	// 数据行
	rowStyle := lipgloss.NewStyle()
	altStyle := lipgloss.NewStyle().Faint(true)
	for ri, row := range t.rows {
		sb.WriteString("  ")
		style := rowStyle
		if ri%2 == 1 {
			style = altStyle
		}
		for i, cell := range row {
			if i >= colCount {
				break
			}
			sb.WriteString(style.Render(padCell(cell, widths[i], t.isAlignRight(i))))
			if i < colCount-1 {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// displayWidth 计算字符串显示宽度（自动处理 ANSI 码和 CJK 字符）
func displayWidth(s string) int {
	return lipgloss.Width(s)
}

// padCell 填充单元格到指定宽度
func padCell(s string, width int, rightAlign bool) string {
	dw := displayWidth(s)
	pad := width - dw
	if pad <= 0 {
		return s
	}
	if rightAlign {
		return strings.Repeat(" ", pad) + s
	}
	return s + strings.Repeat(" ", pad)
}

// totalWidth 计算所有列的总显示宽度（含间距）
func totalWidth(widths []int, colCount int) int {
	total := 0
	for _, w := range widths {
		total += w
	}
	total += (colCount - 1) * 2 // 列间距
	return total
}

// Rows 返回行数
func (t *Table) Rows() int {
	return len(t.rows)
}
