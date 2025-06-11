package conf

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/mcp-client-go/clients/airbnb"
	"github.com/yincongcyincong/mcp-client-go/clients/aliyun"
	"github.com/yincongcyincong/mcp-client-go/clients/amap"
	"github.com/yincongcyincong/mcp-client-go/clients/baidumap"
	"github.com/yincongcyincong/mcp-client-go/clients/binance"
	"github.com/yincongcyincong/mcp-client-go/clients/bitcoin"
	"github.com/yincongcyincong/mcp-client-go/clients/filesystem"
	"github.com/yincongcyincong/mcp-client-go/clients/github"
	"github.com/yincongcyincong/mcp-client-go/clients/googlemap"
	"github.com/yincongcyincong/mcp-client-go/clients/notion"
	"github.com/yincongcyincong/mcp-client-go/clients/param"
	"github.com/yincongcyincong/mcp-client-go/clients/playwright"
	mcp_time "github.com/yincongcyincong/mcp-client-go/clients/time"
	"github.com/yincongcyincong/mcp-client-go/clients/twitter"
	"github.com/yincongcyincong/mcp-client-go/clients/victoriametrics"
	"github.com/yincongcyincong/mcp-client-go/clients/whatsapp"
	"github.com/yincongcyincong/mcp-client-go/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"google.golang.org/genai"
)

type AgentInfo struct {
	Description  string
	DeepseekTool []deepseek.Tool
	VolTool      []*model.Tool
	ToolsName    []string
}

var (
	AmapApiKey               *string
	GithubAccessToken        *string
	VMUrl                    *string
	VMInsertUrl              *string
	VMSelectUrl              *string
	BinanceSwitch            *bool
	TimeZone                 *string
	PlayWrightSSEServer      *string
	PlayWrightSwitch         *bool
	FilecrawlApiKey          *string
	FilePath                 *string
	GoogleMapApiKey          *string
	NotionAuthorization      *string
	NotionVersion            *string
	AliyunAccessKeyID        *string
	AliyunAccessKeySecret    *string
	AirBnbSwitch             *bool
	BitCoinSwitch            *bool
	TwitterApiKey            *string
	TwitterApiSecretKey      *string
	TwitterAccessToken       *string
	TwitterAccessTokenSecret *string
	WhatsappPath             *string
	WhatsappPythonMainFile   *string
	BaidumapApiKey           *string

	DeepseekTools   = make([]deepseek.Tool, 0)
	VolTools        = make([]*model.Tool, 0)
	OpenAITools     = make([]openai.Tool, 0)
	GeminiTools     = make([]*genai.Tool, 0)
	OpenRouterTools = make([]openrouter.Tool, 0)

	TaskTools = map[string]*AgentInfo{
		"map-agent": {
			Description:  "Provides geographic services such as location lookup, route planning, and map navigation.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{amap.NpxAmapMapsMcpServer, googlemap.NpxGooglemapMcpServer, baidumap.NpxBaidumapMcpServer},
		},
		"git-agent": {
			Description:  "Performs Git operations and integrates with GitHub to manage repositories, pull requests, issues, and workflows.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{github.NpxModelContextProtocolGithubServer},
		},
		"browser-agent": {
			Description:  "Simulates browser behavior for tasks like web navigation, data scraping, and automated interactions with web pages.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{playwright.SsePlaywrightMcpServer},
		},
		"vm-agent": {
			Description:  "Manages virtual machines, including starting, stopping, rebooting, monitoring status, and executing remote commands.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{victoriametrics.NpxVictoriaMetricsMcpServer},
		},
		"time-agent": {
			Description:  "Handles time-related functions such as retrieving the current time, converting time zones, and scheduling tasks.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{mcp_time.UvTimeMcpServer},
		},
		"encrypt-agent": {
			Description:  "Provides cryptocurrency-related functions including wallet management, real-time market data, and blockchain transactions.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{bitcoin.NpxBitcoinMcpServer, binance.NpxBinanceMcpServer},
		},
		"twitter-agent": {
			Description:  "Integrates with Twitter to post tweets, fetch user data, read timelines, and analyze trends.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{twitter.NpxTwitterMcpServer},
		},
		"whatapp-agent": {
			Description:  "Connects to WhatsApp for sending messages, reading conversations, and managing contacts.",
			DeepseekTool: make([]deepseek.Tool, 0),
			VolTool:      make([]*model.Tool, 0),
			ToolsName:    []string{whatsapp.UvWhatsAppMcpServer},
		},
	}
)

func InitToolsConf() {
	AmapApiKey = flag.String("amap_api_key", "", "amap api key")
	GithubAccessToken = flag.String("github_access_token", "", "github access token")
	VMUrl = flag.String("vm_url", "", "vm url")
	VMInsertUrl = flag.String("vm_insert_url", "", "vm insert url")
	VMSelectUrl = flag.String("vm_select_url", "", "vm select url")
	BinanceSwitch = flag.Bool("binance_switch", false, "binance switch")
	TimeZone = flag.String("time_zone", "", "time zone")
	PlayWrightSwitch = flag.Bool("play_wright_switch", false, "playwright switch")
	PlayWrightSSEServer = flag.String("play_wright_sse_server", "", "playwright sw server")
	FilecrawlApiKey = flag.String("filecrawl_api_key", "", "filecrawl_api_key")
	FilePath = flag.String("file_path", "", "file path")
	GoogleMapApiKey = flag.String("google_map_api_key", "", "google_map_api_key")
	NotionAuthorization = flag.String("notion_authorization", "", "notion_authorization")
	NotionVersion = flag.String("notion_version", "", "notion_version")
	AliyunAccessKeyID = flag.String("aliyun_access_key_id", "", "aliyun_access_key_id")
	AliyunAccessKeySecret = flag.String("aliyun_access_key_secret", "", "aliyun_access_key_secret")
	AirBnbSwitch = flag.Bool("airbnb_switch", false, "airbnb switch")
	BitCoinSwitch = flag.Bool("bitcoin_switch", false, "bitcoin switch")
	TwitterApiKey = flag.String("twitter_api_key", "", "twitter_api_key")
	TwitterApiSecretKey = flag.String("twitter_api_secret_key", "", "twitter_api_secret_key")
	TwitterAccessToken = flag.String("twitter_access_token", "", "twitter_access_token")
	TwitterAccessTokenSecret = flag.String("twitter_access_token_secret", "", "twitter_access_token_secret")
	WhatsappPath = flag.String("whatapp_path", "", "whatapp_path")
	WhatsappPythonMainFile = flag.String("wahtsapp_python_main_file", "", "wahtsapp_python_main_file")
	BaidumapApiKey = flag.String("baidumap_api_key", "", "baidumap_api_key")

	if os.Getenv("AMAP_API_KEY") != "" {
		*AmapApiKey = os.Getenv("AMAP_API_KEY")
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

	if os.Getenv("BINANCE_SWITCH") != "" {
		*BinanceSwitch = true
	}

	if os.Getenv("TIME_ZONE") != "" {
		*TimeZone = os.Getenv("TIME_ZONE")
	}

	if os.Getenv("PLAY_WRIGHT_SSE_SERVER") != "" {
		*PlayWrightSSEServer = os.Getenv("PLAY_WRIGHT_SSE_SERVER")
	}

	if os.Getenv("PLAY_WRIGHT_SWITCH") != "" {
		*PlayWrightSwitch = true
	}

	if os.Getenv("FILECRAWL_API_KEY") != "" {
		*FilecrawlApiKey = os.Getenv("FILECRAWL_API_KEY")
	}

	if os.Getenv("FILE_PATH") != "" {
		*FilePath = os.Getenv("FILE_PATH")
	}

	if os.Getenv("GOOGLE_MAP_API_KEY") != "" {
		*GoogleMapApiKey = os.Getenv("GOOGLE_MAP_API_KEY")
	}

	if os.Getenv("NOTION_AUTHORIZATION") != "" {
		*NotionAuthorization = os.Getenv("NOTION_AUTHORIZATION")
	}

	if os.Getenv("NOTION_VERSION") != "" {
		*NotionVersion = os.Getenv("NOTION_VERSION")
	}

	if os.Getenv("ALIYUN_ACCESS_KEY_ID") != "" {
		*AliyunAccessKeyID = os.Getenv("ALIYUN_ACCESS_KEY_ID")
	}

	if os.Getenv("ALIYUN_ACCESS_KEY_SECRET") != "" {
		*AliyunAccessKeySecret = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	}

	if os.Getenv("AIRBNB_SWITCH") != "" {
		*AirBnbSwitch = true
	}

	if os.Getenv("BITCOIN_SWITCH") != "" {
		*BitCoinSwitch = true
	}

	if os.Getenv("TWITTER_API_KEY") != "" {
		*TwitterApiKey = os.Getenv("TWITTER_API_KEY")
	}

	if os.Getenv("TWITTER_API_KEY_SECRET") != "" {
		*TwitterApiSecretKey = os.Getenv("TWITTER_API_KEY_SECRET")
	}

	if os.Getenv("TWITTER_ACCESS_TOKEN") != "" {
		*TwitterAccessToken = os.Getenv("TWITTER_ACCESS_TOKEN")
	}

	if os.Getenv("TWITTER_ACCESS_TOKEN_SECRET") != "" {
		*TwitterAccessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	}

	if os.Getenv("WHATSAPP_PATH") != "" {
		*WhatsappPath = os.Getenv("WHATSAPP_PATH")
	}

	if os.Getenv("WHATSAPP_PYTHON_MAIN_FILE") != "" {
		*WhatsappPythonMainFile = os.Getenv("WHATSAPP_PYTHON_MAIN_FILE")
	}

	if os.Getenv("BAIDUMAP_API_KEY") != "" {
		*BaidumapApiKey = os.Getenv("BAIDUMAP_API_KEY")
	}

	logger.Info("TOOLS_CONF", "AmapApiKey", *AmapApiKey)
	logger.Info("TOOLS_CONF", "GithubAccessToken", *GithubAccessToken)
	logger.Info("TOOLS_CONF", "VMUrl", *VMUrl)
	logger.Info("TOOLS_CONF", "VMInsertUrl", *VMInsertUrl)
	logger.Info("TOOLS_CONF", "VMSelectUrl", *VMSelectUrl)
	logger.Info("TOOLS_CONF", "BinanceSwitch", *BinanceSwitch)
	logger.Info("TOOLS_CONF", "TimeZone", *TimeZone)
	logger.Info("TOOLS_CONF", "PlayWrightSwitch", *PlayWrightSwitch)
	logger.Info("TOOLS_CONF", "PlayWrightSSEServer", *PlayWrightSSEServer)
	logger.Info("TOOLS_CONF", "FilePath", *FilePath)
	logger.Info("TOOLS_CONF", "GoogleMapApiKey", *GoogleMapApiKey)
	logger.Info("TOOLS_CONF", "NotionAuthorization", *NotionAuthorization)
	logger.Info("TOOLS_CONF", "NotionVersion", *NotionVersion)
	logger.Info("TOOLS_CONF", "AliyunAccessKeyID", *AliyunAccessKeyID)
	logger.Info("TOOLS_CONF", "AliyunAccessKeySecret", *AliyunAccessKeySecret)
	logger.Info("TOOLS_CONF", "AirBnbSwitch", *AirBnbSwitch)
	logger.Info("TOOLS_CONF", "BinanceSwitch", *BinanceSwitch)
	logger.Info("TOOLS_CONF", "TwitterApiKey", *TwitterApiKey)
	logger.Info("TOOLS_CONF", "TwitterApiSecretKey", *TwitterApiSecretKey)
	logger.Info("TOOLS_CONF", "TwitterAccessTokenSecret", *TwitterAccessTokenSecret)
	logger.Info("TOOLS_CONF", "TwitterAccessToken", *TwitterAccessToken)
	logger.Info("TOOLS_CONF", "BaidumapApiKey", *BaidumapApiKey)

}

func InitTools() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mcpParams := make([]*param.MCPClientConf, 0)
	if *AmapApiKey != "" {
		mcpParams = append(mcpParams,
			amap.InitAmapMCPClient(&amap.AmapParam{
				AmapApiKey: *AmapApiKey,
			}))
	}

	if *GithubAccessToken != "" {
		mcpParams = append(mcpParams,
			github.InitModelContextProtocolGithubMCPClient(&github.GithubParam{
				GithubAccessToken: *GithubAccessToken,
			}))
	}

	if *VMUrl != "" || *VMInsertUrl != "" || *VMSelectUrl != "" {
		mcpParams = append(mcpParams, victoriametrics.InitVictoriaMetricsMCPClient(&victoriametrics.VictoriaMetricsParam{
			VMUrl:       *VMUrl,
			VMInsertUrl: *VMInsertUrl,
			VMSelectUrl: *VMSelectUrl,
		}))
	}

	if *BinanceSwitch {
		mcpParams = append(mcpParams, binance.InitBinanceMCPClient(&binance.BinanceParam{}))
	}

	if *TimeZone != "" {
		mcpParams = append(mcpParams, mcp_time.InitTimeMCPClient(&mcp_time.TimeParam{
			LocalTimezone: *TimeZone,
		}))
	}

	if *PlayWrightSwitch {
		if *PlayWrightSSEServer != "" {
			mcpParams = append(mcpParams, playwright.InitPlaywrightSSEMCPClient(
				&playwright.PlaywrightParam{
					BaseUrl: *PlayWrightSSEServer,
				}))
		} else {
			mcpParams = append(mcpParams, playwright.InitPlaywrightMCPClient(&playwright.PlaywrightParam{}))
		}
	}

	if *FilePath != "" {
		mcpParams = append(mcpParams, filesystem.InitFilesystemMCPClient(&filesystem.FilesystemParam{
			Paths: strings.Split(*FilePath, ","),
		}))
	}

	if *GoogleMapApiKey != "" {
		mcpParams = append(mcpParams, googlemap.InitGooglemapMCPClient(&googlemap.GoogleMapParam{
			GooglemapApiKey: *GoogleMapApiKey,
		}))
	}

	if *NotionAuthorization != "" && *NotionVersion != "" {
		mcpParams = append(mcpParams, notion.InitNotionMCPClient(&notion.NotionParam{
			NotionVersion: *NotionVersion,
			Authorization: *NotionAuthorization,
		}))
	}

	if *AliyunAccessKeyID != "" && *AliyunAccessKeySecret != "" {
		mcpParams = append(mcpParams, aliyun.InitAliyunMCPClient(&aliyun.AliyunParams{
			AliyunAccessKeyID:     *AliyunAccessKeySecret,
			AliyunAccessKeySecret: *AliyunAccessKeySecret,
		}))
	}

	if *AirBnbSwitch {
		mcpParams = append(mcpParams, airbnb.InitAirbnbMCPClient(&airbnb.AirbnbParam{}))
	}

	if *BitCoinSwitch {
		mcpParams = append(mcpParams, bitcoin.InitBitcoinMCPClient(&bitcoin.BitcoinParam{}))
	}

	if *TwitterApiSecretKey != "" && *TwitterAccessToken != "" && *TwitterAccessTokenSecret != "" && *TwitterApiKey != "" {
		mcpParams = append(mcpParams, twitter.InitTwitterMCPClient(&twitter.TwitterParam{
			ApiKey:            *TwitterApiKey,
			ApiSecretKey:      *TwitterApiSecretKey,
			AccessToken:       *TwitterAccessToken,
			AccessTokenSecret: *TwitterAccessTokenSecret,
		}))
	}

	if *WhatsappPath != "" && *WhatsappPythonMainFile != "" {
		mcpParams = append(mcpParams, whatsapp.InitWhatsappMCPClient(&whatsapp.WhaPsAppParam{
			WhatsappPath:   *WhatsappPath,
			PythonMainFile: *WhatsappPythonMainFile,
		}))
	}

	if *BaidumapApiKey != "" {
		mcpParams = append(mcpParams, baidumap.InitBaidumapMCPClient(&baidumap.BaidumapParam{
			BaidumapApiKey: *BaidumapApiKey,
		}))
	}

	errs := clients.RegisterMCPClient(ctx, mcpParams)
	if len(errs) > 0 {
		for mcpServer, err := range errs {
			logger.Error("register mcp client error", "server", mcpServer, "error", err)
		}
	}

	if *AmapApiKey != "" {
		InsertTools(amap.NpxAmapMapsMcpServer)
	}

	if *GithubAccessToken != "" {
		InsertTools(github.NpxModelContextProtocolGithubServer)
	}

	if *VMUrl != "" || *VMInsertUrl != "" || *VMSelectUrl != "" {
		InsertTools(victoriametrics.NpxVictoriaMetricsMcpServer)
	}

	if *BinanceSwitch {
		InsertTools(binance.NpxBinanceMcpServer)
	}

	if *TimeZone != "" {
		InsertTools(mcp_time.UvTimeMcpServer)
	}

	if *PlayWrightSwitch {
		if *PlayWrightSSEServer != "" {
			InsertTools(playwright.SsePlaywrightMcpServer)
		} else {
			InsertTools(playwright.NpxPlaywrightMcpServer)
		}
	}

	if *FilePath != "" {
		InsertTools(filesystem.NpxFilesystemMcpServer)
	}

	if *GoogleMapApiKey != "" {
		InsertTools(googlemap.NpxGooglemapMcpServer)
	}

	if *NotionAuthorization != "" && *NotionVersion != "" {
		InsertTools(notion.NpxNotionMcpServer)
	}

	if *AliyunAccessKeyID != "" && *AliyunAccessKeySecret != "" {
		InsertTools(aliyun.UvxAliyunMcpServer)
	}

	if *AirBnbSwitch {
		InsertTools(airbnb.NpxAirbnbMcpServer)
	}

	if *BitCoinSwitch {
		InsertTools(bitcoin.NpxBitcoinMcpServer)
	}

	if *TwitterApiKey != "" && *TwitterApiSecretKey != "" && *TwitterAccessToken != "" && *TwitterAccessTokenSecret != "" {
		InsertTools(twitter.NpxTwitterMcpServer)
	}

	if *WhatsappPath != "" && *WhatsappPythonMainFile != "" {
		InsertTools(whatsapp.UvWhatsAppMcpServer)
	}

	if *BaidumapApiKey != "" {
		InsertTools(baidumap.NpxBaidumapMcpServer)
	}

	for name, tool := range TaskTools {
		if len(tool.DeepseekTool) == 0 || len(tool.VolTool) == 0 {
			delete(TaskTools, name)
		}
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

		if *UseTools {
			DeepseekTools = append(DeepseekTools, dpTools...)
			VolTools = append(VolTools, volTools...)
			OpenAITools = append(OpenAITools, oaTools...)
			GeminiTools = append(GeminiTools, gmTools...)
			OpenRouterTools = append(OpenRouterTools)
		}
		for _, tool := range TaskTools {
			for _, n := range tool.ToolsName {
				if n == clientName {
					tool.DeepseekTool = dpTools
					tool.VolTool = volTools
				}
			}
		}
	}
}
