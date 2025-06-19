package conf

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/mcp-client-go/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"google.golang.org/genai"
)

type AgentInfo struct {
	Description string
	ToolsName   []string

	DeepseekTool    []deepseek.Tool
	VolTool         []*model.Tool
	OpenAITools     []openai.Tool
	GeminiTools     []*genai.Tool
	OpenRouterTools []openrouter.Tool
}

var (
	McpConfPath *string

	DeepseekTools   = make([]deepseek.Tool, 0)
	VolTools        = make([]*model.Tool, 0)
	OpenAITools     = make([]openai.Tool, 0)
	GeminiTools     = make([]*genai.Tool, 0)
	OpenRouterTools = make([]openrouter.Tool, 0)

	TaskTools = map[string]*AgentInfo{}
)

func InitToolsConf() {
	McpConfPath = flag.String("mcp_conf_path", "./conf/mcp/mcp.json", "mcp conf path")
}

func EnvToolsConf() {
	if os.Getenv("MCP_CONF_PATH") != "" {
		*McpConfPath = os.Getenv("MCP_CONF_PATH")
	}

	logger.Info("TOOLS_CONF", "McpConfPath", *McpConfPath)
}

func InitTools() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		for name, tool := range TaskTools {
			if len(tool.DeepseekTool) == 0 || len(tool.VolTool) == 0 {
				delete(TaskTools, name)
			}
		}
	}()

	mcpParams, err := clients.InitByConfFile(*McpConfPath)
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

		if *UseTools {
			DeepseekTools = append(DeepseekTools, dpTools...)
			VolTools = append(VolTools, volTools...)
			OpenAITools = append(OpenAITools, oaTools...)
			GeminiTools = append(GeminiTools, gmTools...)
			OpenRouterTools = append(OpenRouterTools, orTools...)
		}

		if c.Conf.Description != "" {
			TaskTools[clientName] = &AgentInfo{
				Description:     c.Conf.Description,
				DeepseekTool:    dpTools,
				VolTool:         volTools,
				GeminiTools:     gmTools,
				OpenAITools:     OpenAITools,
				OpenRouterTools: OpenRouterTools,
				ToolsName:       []string{clientName},
			}
		}
	}
}
