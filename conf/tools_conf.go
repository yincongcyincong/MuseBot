package conf

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/mcp-client-go/clients"
	mcputils "github.com/yincongcyincong/mcp-client-go/utils"
	"google.golang.org/genai"
)

const (
	BuiltinWebSearchTool = "web_search"
	BuiltinFileReadTool  = "file_read"
	BuiltinFileWriteTool = "file_write"
	BuiltinShellTool     = "shell"
	BuiltinConfigTool    = "bot_config"
)

type AgentInfo struct {
	Description string `json:"description"`

	Instructions string `json:"-"`
	IsSkill      bool   `json:"-"`
	IsBuiltin    bool   `json:"-"`

	DeepseekTool    []deepseek.Tool   `json:"-"`
	VolTool         []*model.Tool     `json:"-"`
	OpenAITools     []openai.Tool     `json:"-"`
	GeminiTools     []*genai.Tool     `json:"-"`
	OpenRouterTools []openrouter.Tool `json:"-"`
}

type ToolsConf struct {
	McpConfPath       *string `json:"mcp_conf_path"`
	SkillPath         *string `json:"skill_path"`
	AllowedCommands   string  `json:"allowed_commands"`
	CommandTimeoutSec int     `json:"command_timeout_sec"`
	FileRootPath      *string `json:"file_root_path"`
}

var (
	DeepseekTools = make([]deepseek.Tool, 0)
	VolTools      = make([]*model.Tool, 0)
	OpenAITools   = make([]openai.Tool, 0)
	GeminiTools   = make([]*genai.Tool, 0)

	TaskTools     = sync.Map{}
	SkillTools    = sync.Map{}
	ToolsConfInfo = new(ToolsConf)
)

func InitToolsConf() {
	flag.String("allowed_commands", "", "comma-separated executable allowlist")
	flag.Int("command_timeout_sec", 60, "built-in command timeout in seconds")

	ToolsConfInfo.McpConfPath = flag.String("mcp_conf_path", GetAbsPath("conf/mcp/mcp.json"), "mcp conf path")
	ToolsConfInfo.SkillPath = flag.String("skill_path", GetAbsPath("conf/skills"), "skill directory path")
	ToolsConfInfo.FileRootPath = flag.String("file_root_path", GetAbsPath("data/agent_files"), "built-in file operation root")
}

func applyParsedToolsFlags(flags *flag.FlagSet) {
	if allowedCommandsFlag := flags.Lookup("allowed_commands"); allowedCommandsFlag != nil {
		ToolsConfInfo.AllowedCommands = strings.TrimSpace(allowedCommandsFlag.Value.String())
	}
	if commandTimeoutFlag := flags.Lookup("command_timeout_sec"); commandTimeoutFlag != nil {
		timeout, err := strconv.Atoi(commandTimeoutFlag.Value.String())
		if err == nil {
			ToolsConfInfo.CommandTimeoutSec = timeout
		}
	}
}

func EnvToolsConf() {
	applyToolsDefaults()
	if os.Getenv("MCP_CONF_PATH") != "" {
		*ToolsConfInfo.McpConfPath = os.Getenv("MCP_CONF_PATH")
	}
	if os.Getenv("SKILLS_PATH") != "" {
		*ToolsConfInfo.SkillPath = os.Getenv("SKILLS_PATH")
	}
	if os.Getenv("ALLOWED_COMMANDS") != "" {
		ToolsConfInfo.AllowedCommands = os.Getenv("ALLOWED_COMMANDS")
	}
	if os.Getenv("COMMAND_TIMEOUT_SEC") != "" {
		timeout, err := strconv.Atoi(os.Getenv("COMMAND_TIMEOUT_SEC"))
		if err == nil {
			ToolsConfInfo.CommandTimeoutSec = timeout
		}
	}
	if os.Getenv("FILE_ROOT_PATH") != "" {
		*ToolsConfInfo.FileRootPath = os.Getenv("FILE_ROOT_PATH")
	}
}

func applyToolsDefaults() {
	if ToolsConfInfo.McpConfPath == nil {
		ToolsConfInfo.McpConfPath = stringPtr(GetAbsPath("conf/mcp/mcp.json"))
	}
	if ToolsConfInfo.SkillPath == nil {
		ToolsConfInfo.SkillPath = stringPtr(GetAbsPath("conf/skills"))
	}
	if ToolsConfInfo.FileRootPath == nil {
		ToolsConfInfo.FileRootPath = stringPtr(GetAbsPath("data/agent_files"))
	}
	if ToolsConfInfo.CommandTimeoutSec <= 0 {
		ToolsConfInfo.CommandTimeoutSec = 60
	}
}

func stringPtr(value string) *string {
	return &value
}

func InitTools() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		var keysToDelete []any

		TaskTools.Range(func(key, value any) bool {
			aInfo := value.(*AgentInfo)
			if !aInfo.IsSkill && (len(aInfo.DeepseekTool) == 0 || len(aInfo.VolTool) == 0) {
				keysToDelete = append(keysToDelete, key)
			}
			return true
		})

		for _, key := range keysToDelete {
			TaskTools.Delete(key)
		}
	}()

	RegisterBuiltinTools()

	mcpParams, err := clients.InitByConfFile(*ToolsConfInfo.McpConfPath)
	if err != nil {
		logger.Error("init mcp file fail", "err", err)
	}

	errs := clients.RegisterMCPClient(ctx, mcpParams)
	if len(errs) > 0 {
		for mcpServer, err := range errs {
			logger.Error("register mcp client error", "server", mcpServer, "error", err)
		}
	}

	for _, mcpParam := range mcpParams {
		InsertTools(mcpParam.Name)
	}

	InitSkills()
}

func RegisterBuiltinTools() {
	searchTool := mcp.Tool{
		Name:        BuiltinWebSearchTool,
		Description: "Search the public web. Use for current facts, unknown terminology, market data, documentation, prior work, and source links.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Focused web search query",
				},
			},
			Required: []string{"query"},
		},
	}

	readTool := mcp.Tool{
		Name:        BuiltinFileReadTool,
		Description: "Read a UTF-8 text file from the task file sandbox after explicit user approval.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path relative to the task file sandbox",
				},
			},
			Required: []string{"path"},
		},
	}

	writeTool := mcp.Tool{
		Name:        BuiltinFileWriteTool,
		Description: "Create, overwrite, or append to a UTF-8 text file in the task file sandbox after explicit user approval.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path relative to the task file sandbox",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "UTF-8 text content to write",
				},
				"append": map[string]any{
					"type":        "boolean",
					"description": "Append to the file instead of overwriting it",
				},
			},
			Required: []string{"path", "content"},
		},
	}

	shellTool := mcp.Tool{
		Name:        BuiltinShellTool,
		Description: "Run one allowlisted command as an argv array in the task file sandbox after explicit user approval. Shell interpolation is not used.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"args": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Executable and arguments, for example [\"git\", \"status\"]",
				},
				"cwd": map[string]any{
					"type":        "string",
					"description": "Optional working directory relative to the task file sandbox",
				},
				"timeout_sec": map[string]any{
					"type":        "integer",
					"description": "Optional timeout in seconds, capped by the server configuration",
				},
			},
			Required: []string{"args"},
		},
	}

	TaskTools.Store(BuiltinWebSearchTool, &AgentInfo{
		Description:     "Required first-step web research agent. Call web_search to collect facts and source links before planning or producing the final result.",
		IsBuiltin:       true,
		DeepseekTool:    mcputils.TransToolsToDPFunctionCall([]mcp.Tool{searchTool}),
		VolTool:         mcputils.TransToolsToVolFunctionCall([]mcp.Tool{searchTool}),
		OpenAITools:     mcputils.TransToolsToChatGPTFunctionCall([]mcp.Tool{searchTool}),
		GeminiTools:     mcputils.TransToolsToGeminiFunctionCall([]mcp.Tool{searchTool}),
		OpenRouterTools: mcputils.TransToolsToOpenRouterFunctionCall([]mcp.Tool{searchTool}),
	})

	TaskTools.Store(BuiltinFileReadTool, &AgentInfo{
		Description:     "File access agent. Call file_read to inspect an approved text file in the task file sandbox.",
		IsBuiltin:       true,
		DeepseekTool:    mcputils.TransToolsToDPFunctionCall([]mcp.Tool{readTool}),
		VolTool:         mcputils.TransToolsToVolFunctionCall([]mcp.Tool{readTool}),
		OpenAITools:     mcputils.TransToolsToChatGPTFunctionCall([]mcp.Tool{readTool}),
		GeminiTools:     mcputils.TransToolsToGeminiFunctionCall([]mcp.Tool{readTool}),
		OpenRouterTools: mcputils.TransToolsToOpenRouterFunctionCall([]mcp.Tool{readTool}),
	})

	TaskTools.Store(BuiltinFileWriteTool, &AgentInfo{
		Description:     "File mutation agent. Call file_write only for user-approved creation, overwrite, or append operations in the task file sandbox.",
		IsBuiltin:       true,
		DeepseekTool:    mcputils.TransToolsToDPFunctionCall([]mcp.Tool{writeTool}),
		VolTool:         mcputils.TransToolsToVolFunctionCall([]mcp.Tool{writeTool}),
		OpenAITools:     mcputils.TransToolsToChatGPTFunctionCall([]mcp.Tool{writeTool}),
		GeminiTools:     mcputils.TransToolsToGeminiFunctionCall([]mcp.Tool{writeTool}),
		OpenRouterTools: mcputils.TransToolsToOpenRouterFunctionCall([]mcp.Tool{writeTool}),
	})

	TaskTools.Store(BuiltinShellTool, &AgentInfo{
		Description:     "Command execution agent. Call shell only with the exact argv needed, and wait for user approval before execution.",
		IsBuiltin:       true,
		DeepseekTool:    mcputils.TransToolsToDPFunctionCall([]mcp.Tool{shellTool}),
		VolTool:         mcputils.TransToolsToVolFunctionCall([]mcp.Tool{shellTool}),
		OpenAITools:     mcputils.TransToolsToChatGPTFunctionCall([]mcp.Tool{shellTool}),
		GeminiTools:     mcputils.TransToolsToGeminiFunctionCall([]mcp.Tool{shellTool}),
		OpenRouterTools: mcputils.TransToolsToOpenRouterFunctionCall([]mcp.Tool{shellTool}),
	})

	configTool := mcp.Tool{
		Name:        BuiltinConfigTool,
		Description: "Manage MuseBot itself: runtime settings, MCP servers, skills, cron jobs, allowlisted commands, and task files. Use it whenever the task asks to view or change the bot's own configuration. Requires an administrator user.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"type":        map[string]any{"type": "string", "description": "Action type: config, mcp, skill, cron, command, file"},
				"operation":   map[string]any{"type": "string", "description": "config: list|get|set; mcp: list|set|delete|enable|disable|sync; skill: list|create|reload|set_path|search|install; cron: list|get|create|update|enable|disable|delete|clear; command: run; file: list|read|write|edit|mkdir|delete|move"},
				"config_type": map[string]any{"type": "string", "description": "config get/set target: base, audio, llm, photo, rag, video, register, tools"},
				"key":         map[string]any{"type": "string", "description": "config key in json tag form, e.g. smart_mode"},
				"value":       map[string]any{"description": "config value to set"},
				"name":        map[string]any{"type": "string", "description": "mcp/skill/cron name"},
				"query":       map[string]any{"type": "string", "description": "skill search/install capability query"},
				"description": map[string]any{"type": "string", "description": "mcp/skill description"},
				"instructions": map[string]any{
					"type":        "string",
					"description": "skill instructions (SKILL.md body)",
				},
				"cron":    map[string]any{"type": "string", "description": "six-field cron expression for cron create/update"},
				"prompt":  map[string]any{"type": "string", "description": "cron prompt"},
				"args":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "command argv array, e.g. [\"git\",\"status\"]"},
				"timeout": map[string]any{"type": "integer", "description": "command timeout seconds"},
				"path":    map[string]any{"type": "string", "description": "file path relative to the configured file root, or command working directory"},
				"content": map[string]any{"type": "string", "description": "file write content"},
				"old_text": map[string]any{
					"type":        "string",
					"description": "exact text to replace for file edit",
				},
				"new_text": map[string]any{"type": "string", "description": "replacement text for file edit"},
				"target":   map[string]any{"type": "string", "description": "move destination for file move"},
				"id":       map[string]any{"type": "integer", "description": "cron id for cron update/enable/disable/delete"},
			},
			Required: []string{"type", "operation"},
		},
	}

	TaskTools.Store(BuiltinConfigTool, &AgentInfo{
		Description: "MuseBot configuration agent. Assign it when the task needs to view or change MuseBot settings, MCP servers, skills, cron jobs, run allowlisted commands, or manage task files. It calls bot_config and enforces admin-only access.",
		IsBuiltin:   true,
		Instructions: "You manage MuseBot itself through the bot_config tool. Pass exactly the keys \"type\" and \"operation\". " +
			"Use config(list|get|set) with config_type and key for settings; bot_name, http_host, db_type, db_conf, the command allowlist and the file root cannot change at runtime. " +
			"Use mcp(list|set|delete|enable|disable|sync); set needs name plus a complete server object. " +
			"Use skill(list|create|reload|set_path|search|install). Use cron(list|get|create|update|enable|disable|delete|clear) with six-field cron, prompt and numeric id as applicable. " +
			"Use command(run) with an argv array, never a shell string. Use file(list|read|write|edit|mkdir|delete|move) with paths relative to the configured file root; write needs content, edit needs exact old_text and new_text, move needs target. " +
			"Never invent configuration keys, operations, or URLs, and never return secrets.",
		DeepseekTool:    mcputils.TransToolsToDPFunctionCall([]mcp.Tool{configTool}),
		VolTool:         mcputils.TransToolsToVolFunctionCall([]mcp.Tool{configTool}),
		OpenAITools:     mcputils.TransToolsToChatGPTFunctionCall([]mcp.Tool{configTool}),
		GeminiTools:     mcputils.TransToolsToGeminiFunctionCall([]mcp.Tool{configTool}),
		OpenRouterTools: mcputils.TransToolsToOpenRouterFunctionCall([]mcp.Tool{configTool}),
	})
}

func InsertTools(clientName string) {
	if clientName == BuiltinWebSearchTool || clientName == BuiltinFileReadTool ||
		clientName == BuiltinFileWriteTool || clientName == BuiltinShellTool ||
		clientName == BuiltinConfigTool {
		logger.Error("built-in tool name is reserved", "tool", clientName)
		return
	}

	c, err := clients.GetMCPClient(clientName)
	if err != nil {
		logger.Error("get client fail", "err", err)
	} else {
		dpTools := mcputils.TransToolsToDPFunctionCall(c.Tools)
		volTools := mcputils.TransToolsToVolFunctionCall(c.Tools)
		oaTools := mcputils.TransToolsToChatGPTFunctionCall(c.Tools)
		gmTools := mcputils.TransToolsToGeminiFunctionCall(c.Tools)
		orTools := mcputils.TransToolsToOpenRouterFunctionCall(c.Tools)

		TaskTools.Store(clientName, &AgentInfo{
			Description:     fmt.Sprintf("MCP server %s exposes the function tools listed in this agent's schema.", clientName),
			DeepseekTool:    dpTools,
			VolTool:         volTools,
			GeminiTools:     gmTools,
			OpenAITools:     oaTools,
			OpenRouterTools: orTools,
		})

		RebuildChatToolSets()
	}
}

// RebuildChatToolSets rebuilds the global function-call tool lists from the
// registered MCP agents. It must be called whenever TaskTools changes so that
// repeated syncs, updates or deletions replace the tool sets instead of
// appending duplicates or keeping stale tools.
func RebuildChatToolSets() {
	dpTools := make([]deepseek.Tool, 0)
	volTools := make([]*model.Tool, 0)
	oaTools := make([]openai.Tool, 0)
	gmTools := make([]*genai.Tool, 0)

	if BaseConfInfo.UseTools {
		TaskTools.Range(func(_, value any) bool {
			info := value.(*AgentInfo)
			if info.IsSkill || info.IsBuiltin {
				// skills and built-in tools are task agents, not chat tools
				return true
			}
			dpTools = append(dpTools, info.DeepseekTool...)
			volTools = append(volTools, info.VolTool...)
			oaTools = append(oaTools, info.OpenAITools...)
			gmTools = append(gmTools, info.GeminiTools...)
			return true
		})
	}

	DeepseekTools = dpTools
	VolTools = volTools
	OpenAITools = oaTools
	GeminiTools = gmTools
}
