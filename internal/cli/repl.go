package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"study/internal/render"
)

// RunREPL 启动交互式命令行
func RunREPL() {
	fmt.Println(render.Title("📋 study 交互模式"))
	fmt.Println(render.Dim("输入命令，输入 help 查看帮助，输入 exit 退出"))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(render.Bold("study> "))
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 退出
		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println(render.Dim("再见 👋"))
			return
		}

		// 帮助
		if input == "help" || input == "h" || input == "?" {
			printREPLHelp()
			continue
		}

		// 解析并执行
		executeREPLCommand(input)
	}
}

func printREPLHelp() {
	fmt.Println(render.Section("可用命令"))
	fmt.Println()
	fmt.Println(render.Dim("  学习记录"))
	fmt.Println("  log <内容>              记录学习进度")
	fmt.Println("  log list / ll           浏览历史记录")
	fmt.Println()
	fmt.Println(render.Dim("  考试管理"))
	fmt.Println("  exam list / el          查看考试倒计时")
	fmt.Println("  exam add <名称> <日期>    添加考试")
	fmt.Println("  exam del <序号>          删除考试")
	fmt.Println()
	fmt.Println(render.Dim("  薄弱知识点"))
	fmt.Println("  wp list / wl            查看薄弱点列表")
	fmt.Println("  wp add <内容> -l <级别>  添加薄弱点")
	fmt.Println("  wp del <序号>            删除薄弱点")
	fmt.Println()
	fmt.Println(render.Dim("  科目管理"))
	fmt.Println("  subj list / sl          查看科目列表")
	fmt.Println("  subj add <名称>          添加科目")
	fmt.Println("  subj open <名称>         打开资料文件夹")
	fmt.Println()
	fmt.Println(render.Dim("  日记"))
	fmt.Println("  diary / dj              打开今天的日记")
	fmt.Println("  diary list / dl         最近日记列表")
	fmt.Println("  diary search <关键词>    全文搜索")
	fmt.Println()
	fmt.Println(render.Dim("  全局"))
	fmt.Println("  overview / ov / st      查看仪表板")
	fmt.Println("  heatmap / hm            学习热力图")
	fmt.Println("  streak / sk             连续学习统计")
	fmt.Println()
	fmt.Println(render.Dim("  系统"))
	fmt.Println("  help / h / ?            显示此帮助")
	fmt.Println("  exit / quit / q         退出")
	fmt.Println()
	fmt.Println(render.Dim("  Google 服务"))
	fmt.Println("  google login / gl       登录 Google 账号")
	fmt.Println("  google logout / glo     退出 Google 账号")
	fmt.Println("  google status / gst     查看认证状态")
	fmt.Println("  drive upload / dru      上传文件到 Drive")
	fmt.Println("  drive list / drl        列出 Drive 文件")
	fmt.Println("  drive status / drs      查看 Drive 状态")
	fmt.Println()
}

// executeREPLCommand 解析并执行 REPL 命令
func executeREPLCommand(input string) {
	parts := tokenize(input)
	if len(parts) == 0 {
		return
	}

	// 别名映射
	aliasMap := map[string][]string{
		"ll": {"log", "list"},
		"el": {"exam", "list"},
		"wl": {"wp", "list"},
		"sl": {"subj", "list"},
		"dl": {"diary", "list"},
		"ov": {"overview"},
		"st": {"overview"},
		"hm": {"heatmap"},
		"sk": {"streak"},
		"dj": {"diary"},
		"gl":  {"google", "login"},
		"glo": {"google", "logout"},
		"gst": {"google", "status"},
		"dr":  {"drive"},
		"dru": {"drive", "upload"},
		"drl": {"drive", "list"},
		"drs": {"drive", "status"},
	}

	// 展开别名
	if expanded, ok := aliasMap[parts[0]]; ok {
		parts = append(expanded, parts[1:]...)
	}

	// 复用全局 root command
	cmd := GetRootCmd()
	cmd.SetArgs(parts)
	if err := cmd.Execute(); err != nil {
		fmt.Println(render.Red(fmt.Sprintf("❌ %v", err)))
	}
	fmt.Println()
}

// tokenize 分词，支持引号包裹含空格的参数
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false

	for _, ch := range input {
		switch {
		case ch == '"':
			inQuote = !inQuote
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
