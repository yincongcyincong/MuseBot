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
	"github.com/yincongcyincong/MuseBot/utils"
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

	// a command outside the allowlist triggers an interactive approval: the
	// user can add it to the allowlist and run it right away, so no startup
	// configuration is required
	boundaryApproved := false
	if !conf.IsCommandAllowed(executable) {
		detail := fmt.Sprintf("command %q is not in the allowlist; approving adds it to the allowlist and runs it now: argv %s",
			args[0], formatTaskCommandArgs(args))
		approved, err := requestTaskToolApproval(ctx, l, "allow_command", args[0], detail)
		if err != nil {
			return "", err
		}
		if !approved {
			return "", fmt.Errorf("command %q was not approved by the user", args[0])
		}
		conf.AddAllowedCommand(executable)
		boundaryApproved = true
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

	// the boundary approval above already confirmed this exact argv
	if !boundaryApproved {
		detail := fmt.Sprintf("argv %s; timeout %ds", formatTaskCommandArgs(args), timeout)
		approved, err := requestTaskToolApproval(ctx, l, "shell", relativeCwd, detail)
		if err != nil {
			return "", err
		}
		if !approved {
			return taskToolDenialJSON(conf.BuiltinShellTool, relativeCwd, "The user denied this shell operation.")
		}
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
		"detail":    formatTaskCommandArgs(args),
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

// errTaskPathOutsideRoots reports that the requested path is not inside any
// allowed sandbox root; callers may offer an interactive approval to add it.
var errTaskPathOutsideRoots = errors.New("file path is outside the allowed sandbox roots")

func taskFilePath(pathInput string) (string, string, error) {
	pathInput = strings.TrimSpace(pathInput)
	if pathInput == "" {
		return "", "", errors.New("file path is required")
	}

	roots := conf.AllowedPathRoots()
	if len(roots) == 0 {
		return "", "", errors.New("task file sandbox is not configured")
	}

	// pick the root the path belongs to; relative inputs always use the first
	// (configured) root
	rootIndex := 0
	relativePath := ""
	if filepath.IsAbs(pathInput) {
		absoluteInput, err := filepath.Abs(pathInput)
		if err != nil {
			return "", "", err
		}
		found := -1
		for index, root := range roots {
			relative, ok := pathWithinRoot(absoluteInput, root)
			if ok {
				found = index
				relativePath = relative
				break
			}
		}
		if found < 0 {
			return "", "", errTaskPathOutsideRoots
		}
		rootIndex = found
	} else {
		relativePath = filepath.Clean(pathInput)
		if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
			return "", "", errors.New("file path cannot escape the task file sandbox")
		}
	}

	absoluteRoot, err := filepath.Abs(roots[rootIndex])
	if err != nil {
		return "", "", err
	}
	// the configured root is created on demand; runtime-approved roots must exist
	if rootIndex == 0 {
		if err := os.MkdirAll(absoluteRoot, 0o700); err != nil {
			return "", "", err
		}
	}
	evaluatedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", "", err
	}

	current := evaluatedRoot
	for _, segment := range strings.Split(relativePath, string(filepath.Separator)) {
		if segment == "" {
			continue
		}
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", "", err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", "", errors.New("symbolic links are not allowed in the task file sandbox")
		}
	}

	return current, relativePath, nil
}

// pathWithinRoot reports whether absolute path lies under root and returns
// the path relative to the root.
func pathWithinRoot(absolutePath, root string) (string, bool) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	evaluatedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", false
	}
	// resolve symlinked ancestors of the input (e.g. /var -> /private/var)
	// before the prefix match; symlinks inside the root are still rejected by
	// the segment walk in taskFilePath
	evaluatedInput := utils.EvalExistingPathPrefix(absolutePath)
	for _, candidate := range []string{absolutePath, evaluatedInput} {
		if candidate == evaluatedRoot {
			return ".", true
		}
		prefix := evaluatedRoot + string(filepath.Separator)
		if strings.HasPrefix(candidate, prefix) {
			return filepath.Clean(strings.TrimPrefix(candidate, prefix)), true
		}
	}
	return "", false
}
