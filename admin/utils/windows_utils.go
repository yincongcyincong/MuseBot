//go:build windows
// +build windows

package utils

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

func StartDetachedProcess(argsStr string) error {
	lines := strings.Split(argsStr, "\n")
	execName := "MuseBot.exe"
	
	exePath := filepath.Join(botUtils.GetAbsPath(""), execName)
	
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
		CreationFlags: syscall.CREATE_NEW_CONSOLE | syscall.CREATE_NEW_PROCESS_GROUP,
	}
	
	return cmd.Start()
}
