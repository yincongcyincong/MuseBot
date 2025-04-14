package conf

import (
	"context"
	"flag"
	"github.com/yincongcyincong/mcp-client-go/clients/victoriametrics"
	"os"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/mcp-client-go/clients/amap"
	"github.com/yincongcyincong/mcp-client-go/clients/github"
	"github.com/yincongcyincong/mcp-client-go/clients/param"
	"github.com/yincongcyincong/mcp-client-go/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	AmapApiKey        *string
	GithubAccessToken *string
	VMUrl             *string
	VMInsertUrl       *string
	VMSelectUrl       *string
	AllTools          *string

	DeepseekTools = make([]deepseek.Tool, 0)
	VolTools      = make([]*model.Tool, 0)
)

func InitToolsConf() {
	AmapApiKey = flag.String("amap_api_key", "", "amap api key")
	AllTools = flag.String("allow_tools", "*", "allow tools")
	GithubAccessToken = flag.String("github_access_token", "", "github access token")
	VMUrl = flag.String("vm_url", "", "vm url")
	VMInsertUrl = flag.String("vm_insert_url", "", "vm insert url")
	VMSelectUrl = flag.String("vm_select_url", "", "vm select url")

	if os.Getenv("AMAP_API_KEY") != "" {
		*AmapApiKey = os.Getenv("AMAP_API_KEY")
	}

	if os.Getenv("ALLOW_TOOLS") != "" {
		*AllTools = os.Getenv("ALLOW_TOOLS")
	}

	if os.Getenv("GITHUB_ACCESS_TOKEN") != "" {
		*GithubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	}

	if os.Getenv("VM_URL") != "" {
		*VMUrl = os.Getenv("VM_URL")
	}

	if os.Getenv("VM_INSERT_URL") != "" {
		*VMInsertUrl = os.Getenv("VM_INSERT_URL")
	}

	if os.Getenv("VM_SELECT_URL") != "" {
		*VMSelectUrl = os.Getenv("VM_SELECT_URL")
	}

	logger.Info("TOOLS_CONF", "AmapApiKey", *AmapApiKey)
	logger.Info("TOOLS_CONF", "AmapTools", *AllTools)
	logger.Info("TOOLS_CONF", "GithubAccessToken", *GithubAccessToken)

}

func InitTools() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	allTools := make(map[string]bool)
	for _, tool := range strings.Split(*AllTools, ",") {
		allTools[tool] = true
	}

	mcpParams := make([]*param.MCPClientConf, 0)
	if *AmapApiKey != "" {
		mcpParams = append(mcpParams,
			amap.InitAmapMCPClient(&amap.AmapParam{
				AmapApiKey: *AmapApiKey,
			}, "", nil, nil, nil))
	}

	if *GithubAccessToken != "" {
		mcpParams = append(mcpParams,
			github.InitModelContextProtocolGithubMCPClient(&github.GithubParam{
				GithubAccessToken: *GithubAccessToken,
			}, "", nil, nil, nil))
	}

	if *VMUrl != "" || *VMInsertUrl != "" || *VMSelectUrl != "" {
		mcpParams = append(mcpParams, victoriametrics.InitVictoriaMetricsMCPClient(&victoriametrics.VictoriaMetricsParam{
			VMUrl:       *VMUrl,
			VMInsertUrl: *VMInsertUrl,
			VMSelectUrl: *VMSelectUrl,
		}, "", nil, nil, nil))
	}

	err := clients.RegisterMCPClient(ctx, mcpParams)
	if len(err) > 0 {
		logger.Error("register mcp client error", "errors", err)
	}

	if *AmapApiKey != "" {
		InsertTools(amap.NpxAmapMapsMcpServer, allTools)
	}

	if *GithubAccessToken != "" {
		InsertTools(github.NpxModelContextProtocolGithubServer, allTools)
	}

	if *VMUrl != "" || *VMInsertUrl != "" || *VMSelectUrl != "" {
		InsertTools(victoriametrics.NpxVictoriaMetricsMcpServer, allTools)
	}

}

func InsertTools(clientName string, allTools map[string]bool) {
	c, err := clients.GetMCPClient(clientName)
	if err != nil {
		logger.Error("get client fail", "err", err)
	} else {
		dpTools := utils.TransToolsToDPFunctionCall(c.Tools)
		volTools := utils.TransToolsToVolFunctionCall(c.Tools)

		if allTools["*"] {
			DeepseekTools = append(DeepseekTools, dpTools...)
			VolTools = append(VolTools, volTools...)
		} else {
			for _, dpTool := range dpTools {
				if _, ok := allTools[dpTool.Function.Name]; ok {
					DeepseekTools = append(DeepseekTools, dpTool)
				}
			}

			for _, volTool := range volTools {
				if _, ok := allTools[volTool.Function.Name]; ok {
					VolTools = append(VolTools, volTool)
				}
			}
		}
	}
}
