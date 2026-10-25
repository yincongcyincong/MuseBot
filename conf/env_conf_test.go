package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyEnvConf(t *testing.T) {
	orig := *BaseConfInfo
	origCustomPath := BaseConfInfo.CustomPath
	defer func() {
		*BaseConfInfo = orig
		BaseConfInfo.CustomPath = origCustomPath
	}()

	allowed, group, admin := "", "", ""

	t.Run("qq onebot primary env name", func(t *testing.T) {
		BaseConfInfo.QQOneBotSendToken = ""
		t.Setenv("QQ_ONEBOT_SEND_TOKEN", "primary")
		t.Setenv("QQ_ONE_BOT_SEND_TOKEN", "alias")
		applyEnvConf(&allowed, &group, &admin)
		if BaseConfInfo.QQOneBotSendToken != "primary" {
			t.Errorf("expected primary env to win, got %q", BaseConfInfo.QQOneBotSendToken)
		}
		os.Unsetenv("QQ_ONE_BOT_SEND_TOKEN")
	})

	t.Run("qq onebot alias env name", func(t *testing.T) {
		BaseConfInfo.QQOneBotSendToken = ""
		os.Unsetenv("QQ_ONEBOT_SEND_TOKEN")
		t.Setenv("QQ_ONE_BOT_SEND_TOKEN", "alias")
		applyEnvConf(&allowed, &group, &admin)
		if BaseConfInfo.QQOneBotSendToken != "alias" {
			t.Errorf("expected alias env to apply, got %q", BaseConfInfo.QQOneBotSendToken)
		}
		os.Unsetenv("QQ_ONE_BOT_SEND_TOKEN")
	})

	t.Run("custom path env", func(t *testing.T) {
		BaseConfInfo.CustomPath = ""
		t.Setenv("CUSTOM_PATH", "/tmp/ollama")
		applyEnvConf(&allowed, &group, &admin)
		if BaseConfInfo.CustomPath != "/tmp/ollama" {
			t.Errorf("expected custom_path /tmp/ollama, got %q", BaseConfInfo.CustomPath)
		}
	})

	t.Run("smart mode env false", func(t *testing.T) {
		BaseConfInfo.SmartMode = true
		t.Setenv("SMART_MODE", "false")
		applyEnvConf(&allowed, &group, &admin)
		if BaseConfInfo.SmartMode {
			t.Error("expected SMART_MODE=false to disable smart mode")
		}
	})

	t.Run("allowed user ids env", func(t *testing.T) {
		allowed, group, admin = "", "", ""
		t.Setenv("ALLOWED_USER_IDS", "1,2,3")
		applyEnvConf(&allowed, &group, &admin)
		if allowed != "1,2,3" {
			t.Errorf("expected allowed ids 1,2,3, got %q", allowed)
		}
	})
}

func TestEnvToolsConfSkillPath(t *testing.T) {
	// other tests in this package never call InitToolsConf, so the pointer
	// may still be nil here
	applyToolsDefaults()
	origPath := *ToolsConfInfo.SkillPath
	origProvider := ToolsConfInfo.WebSearchProvider
	defer func() {
		*ToolsConfInfo.SkillPath = origPath
		ToolsConfInfo.WebSearchProvider = origProvider
	}()

	t.Run("primary singular env", func(t *testing.T) {
		t.Setenv("SKILL_PATH", "/skills-primary")
		os.Unsetenv("SKILLS_PATH")
		EnvToolsConf()
		if *ToolsConfInfo.SkillPath != "/skills-primary" {
			t.Errorf("expected /skills-primary, got %q", *ToolsConfInfo.SkillPath)
		}
	})

	t.Run("legacy plural env fallback", func(t *testing.T) {
		os.Unsetenv("SKILL_PATH")
		t.Setenv("SKILLS_PATH", "/skills-legacy")
		EnvToolsConf()
		if *ToolsConfInfo.SkillPath != "/skills-legacy" {
			t.Errorf("expected /skills-legacy, got %q", *ToolsConfInfo.SkillPath)
		}
	})

	t.Run("unset leaves value untouched", func(t *testing.T) {
		os.Unsetenv("SKILL_PATH")
		os.Unsetenv("SKILLS_PATH")
		*ToolsConfInfo.SkillPath = "/skills-keep"
		EnvToolsConf()
		if *ToolsConfInfo.SkillPath != "/skills-keep" {
			t.Errorf("expected /skills-keep, got %q", *ToolsConfInfo.SkillPath)
		}
	})
}

func TestEnvPhotoConfUsePreLLM(t *testing.T) {
	orig := PhotoConfInfo.UsePreLLM
	defer func() { PhotoConfInfo.UsePreLLM = orig }()

	t.Setenv("USE_PRE_LLM", "true")
	EnvPhotoConf()
	if !PhotoConfInfo.UsePreLLM {
		t.Error("expected USE_PRE_LLM=true to enable UsePreLLM")
	}

	t.Setenv("USE_PRE_LLM", "false")
	EnvPhotoConf()
	if PhotoConfInfo.UsePreLLM {
		t.Error("expected USE_PRE_LLM=false to disable UsePreLLM")
	}
}

func TestLoadCustomConfLegacyKey(t *testing.T) {
	orig := BaseConfInfo.MaxQAPair
	defer func() { BaseConfInfo.MaxQAPair = orig }()

	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")
	content := `{"telegram_bot_token": "t", "max_qa_pari": 50}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write legacy conf fail: %v", err)
	}

	t.Setenv("CONFIG", "")
	BaseConfInfo.MaxQAPair = 0
	loadCustomConf(path)
	if BaseConfInfo.MaxQAPair != 50 {
		t.Errorf("expected legacy max_qa_pari to set MaxQAPair=50, got %d", BaseConfInfo.MaxQAPair)
	}
}
