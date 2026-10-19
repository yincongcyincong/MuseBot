package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
)

const (
	taskCommandMaxOutputBytes = 16 << 10
	taskCommandMaxTimeoutSec  = 600
)

type taskShellResult struct {
	Status         string   `json:"status"`
	Args           []string `json:"args"`
	Cwd            string   `json:"cwd"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	ExitCode       int      `json:"exit_code"`
	Output         string   `json:"output,omitempty"`
}

func (l *LLM) executeTaskShellTool(ctx context.Context, property map[string]interface{}) (string, error) {
	args, err := taskCommandArgs(property)
	if err != nil {
		return "", err
	}

	executable, err := exec.LookPath(args[0])
	if err != nil {
		return "", err
	}
	if !taskCommandAllowed(executable) {
		return "", fmt.Errorf("command %q is not allowed", args[0])
	}

	cwdInput, _ := property["cwd"].(string)
	workingDir, relativeCwd, err := taskWorkingDirectory(cwdInput)
	if err != nil {
		return "", fmt.Errorf("invalid command working directory: %w", err)
	}

	timeout := taskCommandTimeout()
	if requestedTimeout, ok := property["timeout_sec"].(float64); ok && requestedTimeout > 0 {
		timeout = minInt(int(requestedTimeout), timeout)
	}
	if timeout <= 0 {
		return "", errors.New("command timeout must be positive")
	}

	detail := fmt.Sprintf("argv %s; timeout %ds", formatTaskCommandArgs(args), timeout)
	approved, err := requestTaskToolApproval(ctx, l, "shell", relativeCwd, detail)
	if err != nil {
		return "", err
	}
	if !approved {
		return taskToolDenialJSON(conf.BuiltinShellTool, relativeCwd, "The user denied this shell operation.")
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	command := exec.CommandContext(runCtx, executable, args[1:]...)
	command.Dir = workingDir
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()

	result := taskShellResult{
		Status:         "completed",
		Args:           args,
		Cwd:            relativeCwd,
		TimeoutSeconds: timeout,
		Output:         formatTaskCommandOutput(stdout.Bytes(), stderr.Bytes()),
	}
	switch {
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		result.Status = "timeout"
	case runErr != nil:
		result.Status = "failed"
		var exitError *exec.ExitError
		if errors.As(runErr, &exitError) {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	l.DirectSendMsg(i18n.GetMessage("task_tool_operation_approved", map[string]interface{}{
		"operation": "shell",
		"path":      relativeCwd,
	}), true)

	output, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func taskCommandArgs(property map[string]interface{}) ([]string, error) {
	rawArgs, ok := property["args"].([]interface{})
	if !ok || len(rawArgs) == 0 {
		return nil, errors.New("command args must be a non-empty string array")
	}

	args := make([]string, 0, len(rawArgs))
	for index, rawArg := range rawArgs {
		arg, ok := rawArg.(string)
		if !ok || (index == 0 && strings.TrimSpace(arg) == "") {
			return nil, errors.New("command args must contain strings and a non-empty executable")
		}
		args = append(args, arg)
	}
	return args, nil
}

func taskCommandAllowed(executable string) bool {
	for _, allowedCommand := range strings.Split(conf.ToolsConfInfo.AllowedCommands, ",") {
		allowedCommand = strings.TrimSpace(allowedCommand)
		if allowedCommand == "" {
			continue
		}
		if allowedCommand == executable || allowedCommand == filepath.Base(executable) {
			return true
		}
	}
	return false
}

func taskCommandTimeout() int {
	if conf.ToolsConfInfo.CommandTimeoutSec <= 0 {
		return 60
	}
	return minInt(conf.ToolsConfInfo.CommandTimeoutSec, taskCommandMaxTimeoutSec)
}

func taskWorkingDirectory(pathInput string) (string, string, error) {
	if strings.TrimSpace(pathInput) == "" {
		pathInput = "."
	}
	probePath := filepath.Join(pathInput, ".musebot-cwd-check")
	absolutePath, relativePath, err := taskFilePath(probePath)
	if err != nil {
		return "", "", err
	}
	workingDir := filepath.Dir(absolutePath)
	info, err := os.Stat(workingDir)
	if err != nil {
		return "", "", err
	}
	if !info.IsDir() {
		return "", "", errors.New("working directory is not a directory")
	}
	return workingDir, filepath.Dir(relativePath), nil
}

func formatTaskCommandArgs(args []string) string {
	encoded, err := json.Marshal(args)
	if err != nil {
		return strings.Join(args, " ")
	}
	return string(encoded)
}

func formatTaskCommandOutput(stdout, stderr []byte) string {
	output := bytes.TrimSpace(bytes.Join([][]byte{stdout, stderr}, []byte("\n--- stderr ---\n")))
	if len(output) <= taskCommandMaxOutputBytes {
		return string(output)
	}

	prefix := output[:taskCommandMaxOutputBytes/2]
	suffix := output[len(output)-taskCommandMaxOutputBytes/2:]
	return string(bytes.Join([][]byte{
		prefix,
		[]byte("\n... output truncated ...\n"),
		suffix,
	}, nil))
}

func taskToolDenialJSON(tool, target, message string) (string, error) {
	denial := struct {
		Status  string `json:"status"`
		Tool    string `json:"tool"`
		Target  string `json:"target"`
		Message string `json:"message"`
	}{
		Status:  "denied",
		Tool:    tool,
		Target:  target,
		Message: message,
	}
	output, err := json.Marshal(denial)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
