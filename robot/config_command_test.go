package robot

import (
	"encoding/json"
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

type coerceTestConfig struct {
	Switch  bool    `json:"switch"`
	Count   int     `json:"count"`
	Ratio   float64 `json:"ratio"`
	Caption string  `json:"caption"`
}

func TestSetConversationConfigFieldCoercesScalarStrings(t *testing.T) {
	cfg := &coerceTestConfig{}

	// models frequently quote booleans and numbers
	if err := setConversationConfigField(cfg, "switch", json.RawMessage(`"false"`)); err != nil {
		t.Fatalf("quoted bool: %v", err)
	}
	if cfg.Switch {
		t.Error("switch should be false")
	}
	if err := setConversationConfigField(cfg, "count", json.RawMessage(`"10"`)); err != nil {
		t.Fatalf("quoted int: %v", err)
	}
	if cfg.Count != 10 {
		t.Errorf("count = %d, want 10", cfg.Count)
	}
	if err := setConversationConfigField(cfg, "ratio", json.RawMessage(`"1.5"`)); err != nil {
		t.Fatalf("quoted float: %v", err)
	}
	if cfg.Ratio != 1.5 {
		t.Errorf("ratio = %v, want 1.5", cfg.Ratio)
	}

	// unquoted scalar into a string field
	if err := setConversationConfigField(cfg, "caption", json.RawMessage(`5`)); err != nil {
		t.Fatalf("unquoted string: %v", err)
	}
	if cfg.Caption != "5" {
		t.Errorf("caption = %q, want \"5\"", cfg.Caption)
	}

	// proper JSON types still work
	if err := setConversationConfigField(cfg, "switch", json.RawMessage(`true`)); err != nil {
		t.Fatalf("plain bool: %v", err)
	}
	if !cfg.Switch {
		t.Error("switch should be true")
	}

	// garbage for a bool is still an error
	if err := setConversationConfigField(cfg, "switch", json.RawMessage(`"maybe"`)); err == nil {
		t.Error("expected error for non-boolean string")
	}
}
