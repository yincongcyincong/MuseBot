package llm

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
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

func TestTaskShellToolDeniedBoundaryApproval(t *testing.T) {
	executable, err := exec.LookPath("true")
	assert.NoError(t, err)

	root := t.TempDir()
	previousRoot := conf.ToolsConfInfo.FileRootPath
	previousAllowed := conf.ToolsConfInfo.AllowedCommands
	conf.ToolsConfInfo.FileRootPath = &root
	conf.ToolsConfInfo.AllowedCommands = ""
	t.Cleanup(func() {
		conf.ToolsConfInfo.FileRootPath = previousRoot
		conf.ToolsConfInfo.AllowedCommands = previousAllowed
	})

	messageChan := make(chan *param.MsgInfo, 2)
	l := &LLM{MessageChan: messageChan, UserId: "user", ChatId: "chat"}
	done := make(chan error, 1)
	go func() {
		_, runErr := l.executeTaskShellTool(context.Background(), map[string]interface{}{
			"args": []interface{}{executable},
		})
		done <- runErr
	}()

	// the command is not allowlisted: an allow_command approval is requested
	approvalMessage := <-messageChan
	assert.NotNil(t, approvalMessage)
	assert.True(t, ResolveTaskToolApproval("user", "chat", "", false))

	err = <-done
	assert.ErrorContains(t, err, "was not approved")
}

// ---- task file tools tests ----

func TestMain(m *testing.M) {
	i18n.InitI18n()
	os.Exit(m.Run())
}

func TestTaskFilePathRejectsEscape(t *testing.T) {
	root := t.TempDir()
	previousRoot := conf.ToolsConfInfo.FileRootPath
	conf.ToolsConfInfo.FileRootPath = &root
	t.Cleanup(func() {
		conf.ToolsConfInfo.FileRootPath = previousRoot
	})

	for _, path := range []string{"", "/tmp/musebot", "../outside", "a/../../outside"} {
		_, _, err := taskFilePath(path)
		assert.Error(t, err)
	}
}

func TestTaskFilePathRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	previousRoot := conf.ToolsConfInfo.FileRootPath
	conf.ToolsConfInfo.FileRootPath = &root
	t.Cleanup(func() {
		conf.ToolsConfInfo.FileRootPath = previousRoot
	})

	assert.NoError(t, os.Symlink(filepath.Join(root, "target"), filepath.Join(root, "link")))
	_, _, err := taskFilePath("link/file.txt")
	assert.Error(t, err)
}

func TestResolveTaskToolApprovalRequiresOwner(t *testing.T) {
	assert.False(t, ResolveTaskToolApproval("user", "chat", "missing", true))
	assert.False(t, ResolveTaskToolApproval("user", "chat", "", true))
}

func TestTaskToolApprovalFlow(t *testing.T) {
	messageChan := make(chan *param.MsgInfo, 1)
	l := &LLM{MessageChan: messageChan, UserId: "user", ChatId: "chat"}

	done := make(chan error, 1)
	approved := make(chan bool, 1)
	go func() {
		value, err := requestTaskToolApproval(context.Background(), l, "write", "notes/file.txt", "test")
		approved <- value
		done <- err
	}()

	message := <-messageChan
	assert.NotNil(t, message)
	value, loaded := taskToolApprovals.Load("user\x00chat")
	assert.True(t, loaded)
	approval := value.(*taskToolApproval)
	assert.Len(t, approval.Token, 8)
	assert.True(t, ResolveTaskToolApproval("user", "chat", "", true))
	assert.False(t, ResolveTaskToolApproval("user", "chat", "", true))
	assert.False(t, ResolveTaskToolApproval("user", "chat", approval.Token, true))

	assert.NoError(t, <-done)
	assert.True(t, <-approved)
}

func TestTaskFilePathSupportsRuntimeApprovedRoots(t *testing.T) {
	sandbox := t.TempDir()
	approved := t.TempDir()
	approvedFile := filepath.Join(approved, "notes.txt")
	if err := os.WriteFile(approvedFile, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	oldRoot := conf.ToolsConfInfo.FileRootPath
	defer func() {
		conf.ToolsConfInfo.FileRootPath = oldRoot
		conf.RuntimeBoundaryFileOverride = ""
		resetRuntimeBoundaryForTest()
	}()
	conf.ToolsConfInfo.FileRootPath = &sandbox
	conf.RuntimeBoundaryFileOverride = filepath.Join(t.TempDir(), "boundary.json")
	resetRuntimeBoundaryForTest()

	// absolute path outside every root fails with the boundary sentinel
	if _, _, err := taskFilePath(approvedFile); !errors.Is(err, errTaskPathOutsideRoots) {
		t.Fatalf("err = %v, want errTaskPathOutsideRoots", err)
	}

	// once the user approves the root, the same path resolves
	conf.AddAllowedPathRoot(approved)
	absolute, relative, err := taskFilePath(approvedFile)
	if err != nil {
		t.Fatalf("approved root path failed: %v", err)
	}
	// the resolved path is the symlink-free form (e.g. /private/var on macOS)
	if want := utils.EvalExistingPathPrefix(approvedFile); absolute != want {
		t.Errorf("absolute = %q, want %q", absolute, want)
	}
	if relative != "notes.txt" {
		t.Errorf("relative = %q, want notes.txt", relative)
	}

	// subpaths of the approved root are covered too
	subdir := filepath.Join(approved, "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := taskFilePath(filepath.Join(subdir, "a.txt")); err != nil {
		t.Fatalf("subpath should resolve: %v", err)
	}

	// paths outside all roots are still rejected
	if _, _, err := taskFilePath(filepath.Join(t.TempDir(), "other.txt")); !errors.Is(err, errTaskPathOutsideRoots) {
		t.Fatalf("err = %v, want errTaskPathOutsideRoots", err)
	}
}

func resetRuntimeBoundaryForTest() {
	conf.ResetRuntimeBoundaryForTest()
}
