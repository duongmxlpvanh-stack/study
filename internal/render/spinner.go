package render

import (
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

// Spinner 终端加载动画，输出到 stderr，不影响 stdout 管道。
type Spinner struct {
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// 旋转帧字符
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// StartSpinner 启动一个旋转加载动画。
// 仅在 stderr 是终端时播放动画，否则静默（管道/重定向安全）。
// 调用方负责在操作完成后调用 spinner.Stop() 停止动画。
func StartSpinner(message string) *Spinner {
	s := &Spinner{
		message: message,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	// 非终端环境不启动动画
	if !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()) {
		return s
	}

	go s.run()
	return s
}

func (s *Spinner) run() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-s.stopCh:
			// 清除行
			fmt.Fprintf(os.Stderr, "\r\033[K")
			close(s.doneCh)
			return
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r\033[K%s %s",
				spinnerFrames[frame%len(spinnerFrames)],
				s.message,
			)
			frame++
		}
	}
}

// Stop 停止旋转动画并清除行。
func (s *Spinner) Stop() {
	if s.stopCh == nil {
		return
	}
	select {
	case <-s.stopCh:
		// 已停止
		return
	default:
		close(s.stopCh)
		<-s.doneCh
	}
}
