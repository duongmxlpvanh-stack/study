package cli

import (
	"fmt"
	"strings"

	"study/internal/model"
	"study/internal/render"

	"github.com/spf13/cobra"
)

func newGenCmd() *cobra.Command {
	var (
		checkEnv  bool
		direct    bool
		subject   string
		sections  string
		count     int
		mode      string
		types     string
		difficulty string
		build     bool
		combined  bool
		figures   bool
		title     string
	)

	cmd := &cobra.Command{
		Use:     "gen [自然语言描述]",
		Aliases: []string{"g"},
		Short:   "AI 生成试题与讲义（PDF）",
		Long: render.Title("🤖 AI 生成试题与讲义") + `

通过自然语言描述需求，自动调用 LLM 生成 LaTeX 试题/讲义并编译为 PDF。

示例:
  study gen "出5道微积分导数的题"
  study gen "用概率论第一章出10道选择题，难度期末级别"
  study gen "生成高数第三章的复习讲义"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --check 模式：环境检查
			if checkEnv {
				return runCheckEnv()
			}

			// --direct 模式：直接传参给 Python
			if direct {
				return runDirect(subject, sections, count, mode, types, difficulty, build, combined, figures, title)
			}

			// 自然语言模式
			nl := strings.Join(args, " ")
			if nl == "" {
				// 无参数：显示帮助+可用科目
				cmd.Help()
				fmt.Println()
				return listAvailableSubjects()
			}

			return runNaturalLanguage(nl)
		},
	}

	cmd.Flags().BoolVar(&checkEnv, "check", false, "检查运行环境（Python, XeLaTeX）")
	cmd.Flags().BoolVar(&direct, "direct", false, "直接传参模式（跳过自然语言解析）")
	cmd.Flags().StringVarP(&subject, "subject", "s", "", "科目（约定文件名）")
	cmd.Flags().StringVar(&sections, "sections", "all", "章节，逗号分隔")
	cmd.Flags().IntVarP(&count, "count", "n", 5, "题目数量")
	cmd.Flags().StringVarP(&mode, "mode", "m", "exam", "模式: exam 或 study")
	cmd.Flags().StringVar(&types, "types", "", "题型分配，如 '选择:2,填空:3'")
	cmd.Flags().StringVar(&difficulty, "difficulty", "", "难度: 基础/期末/竞赛")
	cmd.Flags().BoolVarP(&build, "build", "b", true, "编译 PDF")
	cmd.Flags().BoolVar(&combined, "combined", false, "试卷与答案合订")
	cmd.Flags().BoolVar(&figures, "figures", false, "启用 TikZ 配图")
	cmd.Flags().StringVar(&title, "title", "", "自定义标题")

	return cmd
}

// runCheckEnv 检查运行环境
func runCheckEnv() error {
	if GenSvc == nil {
		fmt.Println(render.Red("❌ 生成服务未初始化"))
		return nil
	}

	issues := GenSvc.CheckEnv()
	if len(issues) == 0 {
		fmt.Println(render.Green("✅ 环境就绪 — Python 和 PDF 编译引擎可用"))
		if GenSvc.IsAvailable() {
			fmt.Printf("   Python: %s\n", render.Cyan("可用"))
		}
	} else {
		fmt.Println(render.Yellow("⚠️  环境检查发现问题:"))
		for _, issue := range issues {
			fmt.Printf("   %s %s\n", render.Red("✗"), issue)
		}
	}
	return nil
}

// runNaturalLanguage 自然语言模式
func runNaturalLanguage(nl string) error {
	if GenSvc == nil {
		return fmt.Errorf("生成服务未初始化，请设置 API key 环境变量")
	}
	if !GenSvc.IsAvailable() {
		return fmt.Errorf("未找到 Python，请安装 Python 3.10+\n   下载: https://www.python.org/downloads/")
	}

	// Step 1: LLM 解析意图
	fmt.Println(render.Cyan("🔍 正在分析你的需求..."))
	intent, err := GenSvc.ParseIntent(nl)
	if err != nil {
		return fmt.Errorf("解析需求失败: %w\n\n💡 提示：请更具体地描述，例如：\n   study gen \"出5道微积分导数的题\"\n   study gen \"用概率论第一章出10道选择题\"", err)
	}

	// Step 2: 显示解析结果，让用户确认
	fmt.Printf("\n%s\n", render.Bold("解析结果:"))
	fmt.Printf("   科目:     %s\n", render.Cyan(intent.Subject))
	fmt.Printf("   模式:     %s\n", intent.Mode)
	if len(intent.Sections) > 0 {
		fmt.Printf("   章节:     %s\n", strings.Join(intent.Sections, ", "))
	} else {
		fmt.Printf("   章节:     %s\n", render.Dim("全部"))
	}
	if intent.Count > 0 {
		fmt.Printf("   题数:     %d\n", intent.Count)
	}
	if len(intent.Types) > 0 {
		fmt.Printf("   题型:     %s\n", strings.Join(intent.Types, ", "))
	}
	if intent.Difficulty != "" {
		fmt.Printf("   难度:     %s\n", intent.Difficulty)
	}
	if intent.Combined {
		fmt.Printf("   合订:     是\n")
	}
	if intent.Build {
		fmt.Printf("   编译PDF:  是\n")
	}
	fmt.Println()

	// Step 3: 执行 Python 管线
	result, err := GenSvc.Run(intent)
	if err != nil {
		return fmt.Errorf("生成失败: %w", err)
	}

	if result.Success {
		fmt.Println(render.Green("\n✅ 生成完成!"))
		for _, p := range result.OutPaths {
			fmt.Printf("   📄 %s\n", p)
		}
	} else {
		fmt.Println(render.Red("\n❌ 生成出现问题"))
		for _, e := range result.Errors {
			if e != "" {
				fmt.Printf("   %s\n", render.Dim(e))
			}
		}
	}
	return nil
}

// runDirect 直接传参模式
func runDirect(subject, sections string, count int, mode, types, difficulty string,
	build, combined, figures bool, title string) error {

	if GenSvc == nil {
		return fmt.Errorf("生成服务未初始化")
	}
	if subject == "" {
		return fmt.Errorf("--direct 模式需要 --subject 参数")
	}

	var sectionList []string
	if sections != "all" && sections != "" {
		sectionList = strings.Split(sections, ",")
		for i := range sectionList {
			sectionList[i] = strings.TrimSpace(sectionList[i])
		}
	}

	var typeList []string
	if types != "" {
		typeList = strings.Split(types, ",")
	}

	intent := &model.GenIntent{
		Subject:    subject,
		Mode:       mode,
		Sections:   sectionList,
		Count:      count,
		Types:      typeList,
		Difficulty: difficulty,
		Build:      build,
		Combined:   combined,
		Figures:    figures,
		Title:      title,
		RawInput:   fmt.Sprintf("--direct --subject %s", subject),
	}

	result, err := GenSvc.Run(intent)
	if err != nil {
		return fmt.Errorf("生成失败: %w", err)
	}

	if result.Success {
		fmt.Println(render.Green("\n✅ 生成完成!"))
		for _, p := range result.OutPaths {
			fmt.Printf("   📄 %s\n", p)
		}
	}
	return nil
}

// listAvailableSubjects 列出可用科目
func listAvailableSubjects() error {
	if GenSvc == nil {
		return nil
	}
	subjects, err := GenSvc.LoadSubjectMeta()
	if err != nil {
		return err
	}

	fmt.Println(render.Section("可用科目"))
	for _, m := range subjects {
		sections, _ := GenSvc.ListSections(m.Key)
		fmt.Printf("  %-20s  %s  (%d 个章节)\n", render.Cyan(m.Key), m.Name, len(sections))
	}
	fmt.Printf("\n💡 用法: %s\n", render.Dim(`study gen "出5道<科目><章节>的题"`))
	return nil
}

