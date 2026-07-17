package cli

import (
	"fmt"
	"strconv"

	"study/internal/model"
	"study/internal/render"

	"github.com/spf13/cobra"
)

func newWeakPointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wp",
		Short:   "管理薄弱知识点",
		Long:    "记录和管理自己\"没搞懂的知识点\"，支持紧急程度标签。",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "添加薄弱知识点",
		Example: "study wp add \"多元函数微分\" --level 紧急 --subj 高等数学",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content := args[0]
			level, _ := cmd.Flags().GetString("level")
			subject, _ := cmd.Flags().GetString("subj")

			urgency := model.Urgency(level)
			switch urgency {
			case model.UrgencyUrgent, model.UrgencyRelaxed, model.UrgencyPreExam:
				// 有效
			default:
				return fmt.Errorf("紧急程度必须是: 紧急 / 不急 / 考前看")
			}

			if err := WpSvc.Add(content, urgency, subject); err != nil {
				return err
			}
			afterWrite("添加薄弱点: %s", content)
			fmt.Printf(render.Green("✅ 已添加薄弱点: %s [%s]\n"), content, urgency)
			return nil
		},
	})
	// 标记级别
	wpAddCmd := cmd.Commands()[0] // add 子命令
	wpAddCmd.Flags().StringP("level", "l", "紧急", "紧急程度: 紧急 / 不急 / 考前看")
	wpAddCmd.Flags().StringP("subj", "s", "", "关联科目")

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "列出薄弱知识点（按紧急程度排列）",
		RunE: func(cmd *cobra.Command, args []string) error {
			wps, err := WpSvc.List()
			if err != nil {
				return err
			}
			if len(wps) == 0 {
				fmt.Println(render.Dim("暂无薄弱知识点，使用 study wp add 添加"))
				return nil
			}

			// 分组显示：紧急优先
			groups := map[model.Urgency][]model.WeakPoint{
				model.UrgencyUrgent:  {},
				model.UrgencyPreExam: {},
				model.UrgencyRelaxed: {},
			}
			for _, w := range wps {
				groups[w.Urgency] = append(groups[w.Urgency], w)
			}

			fmt.Println(render.Section("🎯 薄弱知识点"))
			idx := 1
			for _, urgency := range []model.Urgency{model.UrgencyUrgent, model.UrgencyPreExam, model.UrgencyRelaxed} {
				list := groups[urgency]
				if len(list) == 0 {
					continue
				}
				var label string
				switch urgency {
				case model.UrgencyUrgent:
					label = render.Red("🔴 紧急")
				case model.UrgencyPreExam:
					label = render.Yellow("🟡 考前看")
				case model.UrgencyRelaxed:
					label = render.Dim("🟢 不急")
				}
				fmt.Printf("  %s (%d条)\n", label, len(list))
				for _, w := range list {
					subj := ""
					if w.Subject != "" {
						subj = " " + render.Cyan(w.Subject)
					}
					fmt.Printf("    %d. %s%s\n", idx, w.Content, subj)
					idx++
				}
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "del",
		Aliases: []string{"d", "rm"},
		Short:   "删除薄弱知识点（按序号）",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("请输入序号数字")
			}
			if err := WpSvc.Delete(idx); err != nil {
				return err
			}
			afterWrite("删除薄弱点")
			fmt.Println(render.Green("✅ 已删除薄弱点"))
			return nil
		},
	})

	return cmd
}
