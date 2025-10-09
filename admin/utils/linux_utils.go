//go:build linux || darwin
// +build linux darwin

package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/yincongcyincong/MuseBot/utils"
)

func StartDetachedProcess(argsStr string) error {
	lines := strings.Split(argsStr, "\n")
	execName := "MuseBot"

	exePath := filepath.Join(utils.GetAbsPath(""), execName)

	var args []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			args = append(args, trimmed)
		}
	}

	switch runtime.GOOS {
	case "darwin":
		cmdStr := fmt.Sprintf("%s %s", exePath, strings.Join(args, " "))
		script := fmt.Sprintf(`tell application "Terminal"
	activate
	do script "%s"
end tell`, cmdStr)

		cmd := exec.Command("osascript", "-e", script)
		return cmd.Start()

	default: // Linux 或其他
		cmd := exec.Command(exePath, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // 独立进程组
		}
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Start()
	}
}
