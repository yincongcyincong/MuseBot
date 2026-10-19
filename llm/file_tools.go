package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yincongcyincong/MuseBot/i18n"

	"github.com/yincongcyincong/MuseBot/conf"
)

const taskFileMaxBytes = 1 << 20

type taskFileReadResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Bytes   int    `json:"bytes"`
}

type taskFileWriteResult struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	Append bool   `json:"append"`
}

func (l *LLM) executeTaskFileTool(ctx context.Context, name string, property map[string]interface{}) (string, error) {
	pathInput, _ := property["path"].(string)
	absolutePath, relativePath, err := taskFilePath(pathInput)
	if err != nil {
		return "", err
	}

	switch name {
	case conf.BuiltinFileReadTool:
		info, err := os.Stat(absolutePath)
		if err != nil {
			return "", err
		}
		detail := fmt.Sprintf("size %d bytes", info.Size())
		approved, err := requestTaskToolApproval(ctx, l, "read", relativePath, detail)
		if err != nil {
			return "", err
		}
		if !approved {
			return taskToolDenialJSON(name, relativePath, "The user denied this file operation.")
		}
		return l.readTaskFile(absolutePath, relativePath)
	case conf.BuiltinFileWriteTool:
		content, _ := property["content"].(string)
		appendFile, _ := property["append"].(bool)
		if len(content) > taskFileMaxBytes {
			return "", fmt.Errorf("file content exceeds %d bytes", taskFileMaxBytes)
		}

		mode := "overwrite"
		if appendFile {
			mode = "append"
		}
		detail := fmt.Sprintf("mode %s; content %d bytes; preview %.200q", mode, len(content), content)
		approved, err := requestTaskToolApproval(ctx, l, "write", relativePath, detail)
		if err != nil {
			return "", err
		}
		if !approved {
			return taskToolDenialJSON(name, relativePath, "The user denied this file operation.")
		}
		return l.writeTaskFile(absolutePath, relativePath, content, appendFile)
	default:
		return "", fmt.Errorf("unsupported file tool %s", name)
	}
}

func (l *LLM) readTaskFile(absolutePath, relativePath string) (string, error) {
	file, err := os.Open(absolutePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("file path is not a regular file")
	}
	if info.Size() > taskFileMaxBytes {
		return "", fmt.Errorf("file exceeds %d bytes", taskFileMaxBytes)
	}

	content, err := io.ReadAll(io.LimitReader(file, taskFileMaxBytes+1))
	if err != nil {
		return "", err
	}
	if len(content) > taskFileMaxBytes {
		return "", fmt.Errorf("file exceeds %d bytes", taskFileMaxBytes)
	}

	result := taskFileReadResult{Path: relativePath, Content: string(content), Bytes: len(content)}
	output, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	l.DirectSendMsg(i18n.GetMessage("task_tool_operation_approved", map[string]interface{}{
		"operation": "read",
		"path":      relativePath,
	}), true)
	return string(output), nil
}

func (l *LLM) writeTaskFile(absolutePath, relativePath, content string, appendFile bool) (string, error) {
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o700); err != nil {
		return "", err
	}
	if info, err := os.Lstat(absolutePath); err == nil {
		if !info.Mode().IsRegular() {
			return "", errors.New("file path is not a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	flags := os.O_WRONLY | os.O_CREATE
	if appendFile {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(absolutePath, flags, 0o600)
	if err != nil {
		return "", err
	}
	defer file.Close()

	written, err := file.WriteString(content)
	if err != nil {
		return "", err
	}

	result := taskFileWriteResult{Path: relativePath, Bytes: written, Append: appendFile}
	output, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	l.DirectSendMsg(i18n.GetMessage("task_tool_operation_approved", map[string]interface{}{
		"operation": "write",
		"path":      relativePath,
	}), true)
	return string(output), nil
}

func taskFilePath(pathInput string) (string, string, error) {
	pathInput = strings.TrimSpace(pathInput)
	if pathInput == "" || filepath.IsAbs(pathInput) {
		return "", "", errors.New("file path must be relative to the task file sandbox")
	}

	relativePath := filepath.Clean(pathInput)
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", "", errors.New("file path cannot escape the task file sandbox")
	}

	root := conf.ToolsConfInfo.FileRootPath
	if root == nil || *root == "" {
		return "", "", errors.New("task file sandbox is not configured")
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(absoluteRoot, 0o700); err != nil {
		return "", "", err
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
