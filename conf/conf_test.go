package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCustomConf(t *testing.T) {
	dir := t.TempDir()

	origToken := BaseConfInfo.TelegramBotToken
	origType := BaseConfInfo.Type
	origSmartMode := BaseConfInfo.SmartMode
	origAdminIds := BaseConfInfo.AdminUserIds
	origProvider := ToolsConfInfo.WebSearchProvider
	origBotName := BaseConfInfo.BotName
	defer func() {
		BaseConfInfo.TelegramBotToken = origToken
		BaseConfInfo.Type = origType
		BaseConfInfo.SmartMode = origSmartMode
		BaseConfInfo.AdminUserIds = origAdminIds
		ToolsConfInfo.WebSearchProvider = origProvider
		BaseConfInfo.BotName = origBotName
	}()

	// flat file: every top-level key is a base config key
	flat := filepath.Join(dir, "flat.json")
	flatContent := `{
		"telegram_bot_token": "flat-token",
		"type": "openai",
		"smart_mode": true,
		"admin_user_ids": {"123456": true}
	}`
	if err := os.WriteFile(flat, []byte(flatContent), 0644); err != nil {
		t.Fatalf("write flat conf fail: %v", err)
	}

	BaseConfInfo.TelegramBotToken = ""
	BaseConfInfo.Type = ""
	BaseConfInfo.SmartMode = false
	BaseConfInfo.AdminUserIds = map[string]bool{}
	t.Setenv("CONFIG", flat)
	loadCustomConf("")

	if BaseConfInfo.TelegramBotToken != "flat-token" {
		t.Errorf("expected telegram_bot_token 'flat-token', got %q", BaseConfInfo.TelegramBotToken)
	}
	if BaseConfInfo.Type != "openai" {
		t.Errorf("expected type 'openai', got %q", BaseConfInfo.Type)
	}
	if !BaseConfInfo.SmartMode {
		t.Error("expected smart_mode true")
	}
	if !BaseConfInfo.AdminUserIds["123456"] {
		t.Error("expected admin_user_ids to contain 123456")
	}

	// sectioned file: base keys nest under "base", other sections map to
	// their own conf structs
	nested := filepath.Join(dir, "nested.json")
	nestedContent := `{
		"base": {
			"bot_name": "NestedBot",
			"telegram_bot_token": "nested-token"
		},
		"tools": {
			"web_search_provider": "parallel"
		}
	}`
	if err := os.WriteFile(nested, []byte(nestedContent), 0644); err != nil {
		t.Fatalf("write nested conf fail: %v", err)
	}

	BaseConfInfo.TelegramBotToken = ""
	ToolsConfInfo.WebSearchProvider = ""
	t.Setenv("CONFIG", "")
	loadCustomConf(nested)

	if BaseConfInfo.BotName != "NestedBot" {
		t.Errorf("expected bot_name 'NestedBot', got %q", BaseConfInfo.BotName)
	}
	if BaseConfInfo.TelegramBotToken != "nested-token" {
		t.Errorf("expected telegram_bot_token 'nested-token', got %q", BaseConfInfo.TelegramBotToken)
	}
	if ToolsConfInfo.WebSearchProvider != "parallel" {
		t.Errorf("expected web_search_provider 'parallel', got %q", ToolsConfInfo.WebSearchProvider)
	}

	// empty flag and env must not change config
	t.Setenv("CONFIG", "")
	BaseConfInfo.TelegramBotToken = "unchanged"
	loadCustomConf("")
	if BaseConfInfo.TelegramBotToken != "unchanged" {
		t.Error("empty CONFIG must not change config")
	}
}

func TestApplyLegacyConfKeys(t *testing.T) {
	// the misspelled key from older versions migrates to the current name
	m := map[string]interface{}{"max_qa_pari": float64(50)}
	applyLegacyConfKeys(m)
	if m["max_qa_pair"] != float64(50) {
		t.Errorf("expected max_qa_pair migrated to 50, got %v", m["max_qa_pair"])
	}
	if _, ok := m["max_qa_pari"]; ok {
		t.Error("legacy max_qa_pari key should be removed after migration")
	}

	// the current name always wins
	m = map[string]interface{}{"max_qa_pair": float64(10), "max_qa_pari": float64(50)}
	applyLegacyConfKeys(m)
	if m["max_qa_pair"] != float64(10) {
		t.Errorf("expected max_qa_pair kept at 10, got %v", m["max_qa_pair"])
	}

	// files without the legacy key are untouched
	m = map[string]interface{}{"bot_name": "x"}
	applyLegacyConfKeys(m)
	if _, ok := m["max_qa_pair"]; ok {
		t.Error("max_qa_pair must not be invented when absent")
	}
}
