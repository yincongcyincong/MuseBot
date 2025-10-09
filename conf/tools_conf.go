package conf

import (
	"context"
	"flag"
	"os"
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

	DeepseekTool    []deepseek.Tool   `json:"-"`
	VolTool         []*model.Tool     `json:"-"`
	OpenAITools     []openai.Tool     `json:"-"`
	GeminiTools     []*genai.Tool     `json:"-"`
	OpenRouterTools []openrouter.Tool `json:"-"`
}

var (
	McpConfPath *string

	DeepseekTools   = make([]deepseek.Tool, 0)
	VolTools        = make([]*model.Tool, 0)
	OpenAITools     = make([]openai.Tool, 0)
	GeminiTools     = make([]*genai.Tool, 0)
	OpenRouterTools = make([]openrouter.Tool, 0)

	TaskTools = sync.Map{}
)

func InitToolsConf() {
	McpConfPath = flag.String("mcp_conf_path", GetAbsPath("conf/mcp/mcp.json"), "mcp conf path")
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
		var keysToDelete []any

		TaskTools.Range(func(key, value any) bool {
			aInfo := value.(*AgentInfo)
			if len(aInfo.DeepseekTool) == 0 || len(aInfo.VolTool) == 0 {
				keysToDelete = append(keysToDelete, key)
			}
			return true
		})

		for _, key := range keysToDelete {
			TaskTools.Delete(key)
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

		if *BaseConfInfo.UseTools {
			DeepseekTools = append(DeepseekTools, dpTools...)
			VolTools = append(VolTools, volTools...)
			OpenAITools = append(OpenAITools, oaTools...)
			GeminiTools = append(GeminiTools, gmTools...)
			OpenRouterTools = append(OpenRouterTools, orTools...)
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
