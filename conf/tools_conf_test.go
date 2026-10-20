package conf

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/sashabaranov/go-openai"
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

func TestRebuildChatToolSetsIsIdempotent(t *testing.T) {
	oldUseTools := BaseConfInfo.UseTools
	oldDeepseekTools := DeepseekTools
	oldVolTools := VolTools
	oldOpenAITools := OpenAITools
	oldGeminiTools := GeminiTools
	BaseConfInfo.UseTools = true
	defer func() {
		BaseConfInfo.UseTools = oldUseTools
		DeepseekTools = oldDeepseekTools
		VolTools = oldVolTools
		OpenAITools = oldOpenAITools
		GeminiTools = oldGeminiTools
		TaskTools.Delete("test_mcp_server")
		TaskTools.Delete(BuiltinWebSearchTool)
		TaskTools.Delete("test_skill")
	}()

	TaskTools.Store("test_mcp_server", &AgentInfo{
		OpenAITools: []openai.Tool{{
			Type:     openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{Name: "rebuild_demo_tool"},
		}},
	})
	TaskTools.Store(BuiltinWebSearchTool, &AgentInfo{
		IsBuiltin: true,
		OpenAITools: []openai.Tool{{
			Type:     openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{Name: "rebuild_builtin_tool"},
		}},
	})
	TaskTools.Store("test_skill", &AgentInfo{
		IsSkill: true,
		OpenAITools: []openai.Tool{{
			Type:     openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{Name: "rebuild_skill_tool"},
		}},
	})

	countTool := func(name string) int {
		n := 0
		for _, tool := range OpenAITools {
			if tool.Function != nil && tool.Function.Name == name {
				n++
			}
		}
		return n
	}

	RebuildChatToolSets()
	RebuildChatToolSets()

	if got := countTool("rebuild_demo_tool"); got != 1 {
		t.Errorf("MCP tool appears %d times after repeated rebuilds, want 1", got)
	}
	if got := countTool("rebuild_builtin_tool"); got != 0 {
		t.Errorf("built-in agent leaked into chat tools %d times, want 0", got)
	}
	if got := countTool("rebuild_skill_tool"); got != 0 {
		t.Errorf("skill agent leaked into chat tools %d times, want 0", got)
	}

	BaseConfInfo.UseTools = false
	RebuildChatToolSets()
	if got := countTool("rebuild_demo_tool"); got != 0 {
		t.Errorf("tools kept after UseTools disabled: %d, want 0", got)
	}
}

func TestRuntimeBoundaryCommandsAndPaths(t *testing.T) {
	oldAllowed := ToolsConfInfo.AllowedCommands
	oldFileRoot := ToolsConfInfo.FileRootPath
	defer func() {
		ToolsConfInfo.AllowedCommands = oldAllowed
		ToolsConfInfo.FileRootPath = oldFileRoot
		RuntimeBoundaryFileOverride = ""
		ResetRuntimeBoundaryForTest()
	}()
	RuntimeBoundaryFileOverride = filepath.Join(t.TempDir(), "runtime_boundary.json")
	ResetRuntimeBoundaryForTest()
	ToolsConfInfo.AllowedCommands = ""

	if IsCommandAllowed("/usr/bin/git") {
		t.Fatal("git should not be allowed before approval")
	}
	AddAllowedCommand("/usr/bin/git")
	if !IsCommandAllowed("/usr/bin/git") {
		t.Fatal("git should be allowed after approval")
	}
	if !IsCommandAllowed("git") {
		t.Fatal("basename form should be allowed too")
	}
	// duplicate approvals are idempotent
	AddAllowedCommand("git")

	if len(AllowedPathRoots()) != 0 {
		t.Fatal("no path roots expected before approval")
	}
	root := t.TempDir()
	AddAllowedPathRoot(root)
	roots := AllowedPathRoots()
	if len(roots) != 1 || roots[0] != root {
		t.Fatalf("roots = %v, want [%s]", roots, root)
	}

	// approvals survive a reload (new process)
	ResetRuntimeBoundaryForTest()
	if !IsCommandAllowed("git") {
		t.Fatal("approved command should survive reload")
	}
	roots = AllowedPathRoots()
	if len(roots) != 1 || roots[0] != root {
		t.Fatalf("roots after reload = %v", roots)
	}
}
