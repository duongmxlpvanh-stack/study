package render

import (
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

// Typewriter 逐字打印效果。
// 仅在 stdout 是终端时播放动画，否则直接打印整行（管道安全）。
func Typewriter(text string, delay time.Duration) {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		fmt.Println(text)
		return
	}

	for _, ch := range text {
		fmt.Print(string(ch))
		os.Stdout.Sync()
		time.Sleep(delay)
	}
	fmt.Println()
}
