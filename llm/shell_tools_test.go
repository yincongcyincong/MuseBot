package llm

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestTaskShellToolRunsApprovedCommand(t *testing.T) {
	executable, err := exec.LookPath("true")
	assert.NoError(t, err)

	root := t.TempDir()
	previousRoot := conf.ToolsConfInfo.FileRootPath
	previousAllowed := conf.ToolsConfInfo.AllowedCommands
	previousTimeout := conf.ToolsConfInfo.CommandTimeoutSec
	conf.ToolsConfInfo.FileRootPath = &root
	conf.ToolsConfInfo.AllowedCommands = executable
	conf.ToolsConfInfo.CommandTimeoutSec = 5
	t.Cleanup(func() {
		conf.ToolsConfInfo.FileRootPath = previousRoot
		conf.ToolsConfInfo.AllowedCommands = previousAllowed
		conf.ToolsConfInfo.CommandTimeoutSec = previousTimeout
	})

	messageChan := make(chan *param.MsgInfo, 2)
	l := &LLM{MessageChan: messageChan, UserId: "user", ChatId: "chat"}
	done := make(chan string, 1)
	errors := make(chan error, 1)
	go func() {
		result, runErr := l.executeTaskShellTool(context.Background(), map[string]interface{}{
			"args": []interface{}{executable},
		})
		done <- result
		errors <- runErr
	}()

	approvalMessage := <-messageChan
	assert.NotNil(t, approvalMessage)
	assert.True(t, ResolveTaskToolApproval("user", "chat", "", true))

	result := <-done
	assert.NoError(t, <-errors)
	assert.Contains(t, result, `"status":"completed"`)
	assert.Contains(t, result, `"exit_code":0`)
	assert.Len(t, messageChan, 1)
}

func TestTaskShellToolRequiresAllowedCommand(t *testing.T) {
	executable, err := exec.LookPath("true")
	assert.NoError(t, err)

	previousAllowed := conf.ToolsConfInfo.AllowedCommands
	conf.ToolsConfInfo.AllowedCommands = ""
	t.Cleanup(func() {
		conf.ToolsConfInfo.AllowedCommands = previousAllowed
	})

	l := &LLM{}
	_, err = l.executeTaskShellTool(context.Background(), map[string]interface{}{
		"args": []interface{}{executable},
	})
	assert.ErrorContains(t, err, "is not allowed")
}
