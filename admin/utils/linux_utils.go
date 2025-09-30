//go:build linux || darwin
// +build linux darwin

package utils

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	
	"github.com/yincongcyincong/MuseBot/utils"
)

func StartDetachedProcess(argsStr string) error {
	lines := strings.Split(argsStr, "\n")
	execName := "MuseBot"
	
	exePath := filepath.Join(utils.GetAbsPath(""), execName)
	
	var args []string
	args = append(args, exePath)
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			args = append(args, trimmed)
		}
	}
	
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Linux/macOS: 新进程组
	}
	
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	return cmd.Start()
}
