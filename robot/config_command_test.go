package robot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestConversationCommandAllowed(t *testing.T) {
	oldAllowedCommands := conf.ToolsConfInfo.AllowedCommands
	defer func() {
		conf.ToolsConfInfo.AllowedCommands = oldAllowedCommands
	}()

	conf.ToolsConfInfo.AllowedCommands = "git,/usr/bin/printf"
	if !conversationCommandAllowed("/usr/local/bin/git") {
		t.Error("expected basename allowlist entry to match")
	}
	if !conversationCommandAllowed("/usr/bin/printf") {
		t.Error("expected absolute allowlist entry to match")
	}
	if conversationCommandAllowed("/bin/sh") {
		t.Error("expected non-allowlisted command to be rejected")
	}
}

func TestSecureConversationFilePathRejectsEscapeAndSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	oldFileRoot := conf.ToolsConfInfo.FileRootPath
	defer func() {
		conf.ToolsConfInfo.FileRootPath = oldFileRoot
	}()
	conf.ToolsConfInfo.FileRootPath = &root

	if _, err := secureConversationFilePath("../outside.txt", false); err == nil {
		t.Fatal("expected parent traversal to be rejected")
	}
	if _, err := secureConversationFilePath("link", true); err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}

func TestConversationFileWriteAndEdit(t *testing.T) {
	root := t.TempDir()
	oldFileRoot := conf.ToolsConfInfo.FileRootPath
	defer func() {
		conf.ToolsConfInfo.FileRootPath = oldFileRoot
	}()
	conf.ToolsConfInfo.FileRootPath = &root

	relativePath := "notes/a.md"
	if _, err := writeConversationFile(relativePath, "hello old world"); err != nil {
		t.Fatal(err)
	}
	if _, err := editConversationFile(relativePath, "old", "new"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "notes", "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello new world" {
		t.Fatalf("content = %q", content)
	}
}

func TestConversationFileRootCannotBeDeletedOrMoved(t *testing.T) {
	root := t.TempDir()
	oldFileRoot := conf.ToolsConfInfo.FileRootPath
	defer func() {
		conf.ToolsConfInfo.FileRootPath = oldFileRoot
	}()
	conf.ToolsConfInfo.FileRootPath = &root

	robotInfo := new(RobotInfo)
	if _, err := robotInfo.executeConversationFileAction(&param.ConversationConfigAction{
		Type: "file", Operation: "delete", Path: ".",
	}); err == nil {
		t.Error("expected deleting the file root to be rejected")
	}
	if _, err := moveConversationFile(".", "child"); err == nil {
		t.Error("expected moving the file root to be rejected")
	}
}

func TestRuntimeImmutableSystemBoundaryConfig(t *testing.T) {
	action := &param.ConversationConfigAction{
		ConfigType: "tools",
		Key:        "allowed_commands",
		Value:      []byte(`"sh"`),
	}
	robotInfo := new(RobotInfo)
	_, err := robotInfo.executeConfigAction(action)
	if err == nil {
		t.Fatal("expected runtime command allowlist change to be rejected")
	}
}
