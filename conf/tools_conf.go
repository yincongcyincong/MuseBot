package conf

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/mcp-client-go/utils"
	"google.golang.org/genai"
)

type AgentInfo struct {
	Description string `json:"description"`

	Instructions string `json:"-"`
	IsSkill      bool   `json:"-"`

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

func InsertTools(clientName string) {
	c, err := clients.GetMCPClient(clientName)
	if err != nil {
		logger.Error("get client fail", "err", err)
	} else {
		dpTools := utils.TransToolsToDPFunctionCall(c.Tools)
		volTools := utils.TransToolsToVolFunctionCall(c.Tools)
		oaTools := utils.TransToolsToChatGPTFunctionCall(c.Tools)
		gmTools := utils.TransToolsToGeminiFunctionCall(c.Tools)
		orTools := utils.TransToolsToOpenRouterFunctionCall(c.Tools)

		if BaseConfInfo.UseTools {
			DeepseekTools = append(DeepseekTools, dpTools...)
			VolTools = append(VolTools, volTools...)
			OpenAITools = append(OpenAITools, oaTools...)
			GeminiTools = append(GeminiTools, gmTools...)
		}

		if c.Conf.Description != "" {
			TaskTools.Store(clientName, &AgentInfo{
				Description:     c.Conf.Description,
				DeepseekTool:    dpTools,
				VolTool:         volTools,
				GeminiTools:     gmTools,
				OpenAITools:     oaTools,
				OpenRouterTools: orTools,
			})
		}
	}
}
