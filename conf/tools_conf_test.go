package conf

import (
	"flag"
	"testing"
)

func TestApplyParsedToolsFlags(t *testing.T) {
	oldToolsConf := ToolsConfInfo
	ToolsConfInfo = new(ToolsConf)
	defer func() {
		ToolsConfInfo = oldToolsConf
	}()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.String("allowed_commands", "", "comma-separated executable allowlist")
	flags.Int("command_timeout_sec", 60, "built-in command timeout in seconds")
	if err := flags.Parse([]string{"-allowed_commands", " git,ls ", "-command_timeout_sec", "30"}); err != nil {
		t.Fatal(err)
	}

	applyParsedToolsFlags(flags)

	if ToolsConfInfo.AllowedCommands != "git,ls" {
		t.Errorf("AllowedCommands = %q", ToolsConfInfo.AllowedCommands)
	}
	if ToolsConfInfo.CommandTimeoutSec != 30 {
		t.Errorf("CommandTimeoutSec = %d", ToolsConfInfo.CommandTimeoutSec)
	}
}

func TestEnvToolsConfAppliesDefaultsAndEnvironment(t *testing.T) {
	t.Setenv("MCP_CONF_PATH", "/tmp/mcp.json")
	t.Setenv("SKILLS_PATH", "/tmp/skills")
	t.Setenv("ALLOWED_COMMANDS", "git,ls")
	t.Setenv("COMMAND_TIMEOUT_SEC", "30")
	t.Setenv("FILE_ROOT_PATH", "/tmp/agent-files")

	oldToolsConf := ToolsConfInfo
	ToolsConfInfo = new(ToolsConf)
	defer func() {
		ToolsConfInfo = oldToolsConf
	}()

	EnvToolsConf()

	if *ToolsConfInfo.McpConfPath != "/tmp/mcp.json" {
		t.Errorf("McpConfPath = %q", *ToolsConfInfo.McpConfPath)
	}
	if *ToolsConfInfo.SkillPath != "/tmp/skills" {
		t.Errorf("SkillPath = %q", *ToolsConfInfo.SkillPath)
	}
	if ToolsConfInfo.AllowedCommands != "git,ls" {
		t.Errorf("AllowedCommands = %q", ToolsConfInfo.AllowedCommands)
	}
	if ToolsConfInfo.CommandTimeoutSec != 30 {
		t.Errorf("CommandTimeoutSec = %d", ToolsConfInfo.CommandTimeoutSec)
	}
	if *ToolsConfInfo.FileRootPath != "/tmp/agent-files" {
		t.Errorf("FileRootPath = %q", *ToolsConfInfo.FileRootPath)
	}
}

func TestEnvToolsConfDefaultsMissingFileRoot(t *testing.T) {
	oldToolsConf := ToolsConfInfo
	ToolsConfInfo = new(ToolsConf)
	defer func() {
		ToolsConfInfo = oldToolsConf
	}()

	EnvToolsConf()

	if ToolsConfInfo.FileRootPath == nil || *ToolsConfInfo.FileRootPath == "" {
		t.Fatal("expected a non-empty default file root")
	}
	if ToolsConfInfo.CommandTimeoutSec != 60 {
		t.Errorf("CommandTimeoutSec = %d, want 60", ToolsConfInfo.CommandTimeoutSec)
	}
}
