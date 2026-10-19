package llm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/param"
)

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

func TestTaskFileWriteAndRead(t *testing.T) {
	root := t.TempDir()
	previousRoot := conf.ToolsConfInfo.FileRootPath
	conf.ToolsConfInfo.FileRootPath = &root
	t.Cleanup(func() {
		conf.ToolsConfInfo.FileRootPath = previousRoot
	})

	l := &LLM{}
	absolutePath, relativePath, err := taskFilePath("notes/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, "notes/file.txt", relativePath)

	content, err := l.writeTaskFile(absolutePath, relativePath, "hello", false)
	assert.NoError(t, err)
	assert.Contains(t, content, `"bytes":5`)

	_, err = l.writeTaskFile(absolutePath, relativePath, " world", true)
	assert.NoError(t, err)
	data, err := os.ReadFile(absolutePath)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(data))
}

func TestResolveTaskToolApprovalRequiresOwner(t *testing.T) {
	assert.False(t, ResolveTaskToolApproval("user", "chat", "missing", true))
	assert.False(t, ResolveTaskToolApproval("user", "chat", "", true))
}

func TestTaskFileApprovalFlow(t *testing.T) {
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
