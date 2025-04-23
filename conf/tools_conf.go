package conf

import (
    "context"
    "flag"
    "os"
    "strings"

    "github.com/cohesion-org/deepseek-go"
    "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
    "github.com/yincongcyincong/mcp-client-go/clients"
    "github.com/yincongcyincong/mcp-client-go/clients/amap"
    "github.com/yincongcyincong/mcp-client-go/clients/binance"
    "github.com/yincongcyincong/mcp-client-go/clients/filesystem"
    "github.com/yincongcyincong/mcp-client-go/clients/github"
    "github.com/yincongcyincong/mcp-client-go/clients/googlemap"
    "github.com/yincongcyincong/mcp-client-go/clients/notion"
    "github.com/yincongcyincong/mcp-client-go/clients/param"
    "github.com/yincongcyincong/mcp-client-go/clients/playwright"
    mcp_time "github.com/yincongcyincong/mcp-client-go/clients/time"
    "github.com/yincongcyincong/mcp-client-go/clients/victoriametrics"
    "github.com/yincongcyincong/mcp-client-go/utils"
    "github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
    AmapApiKey          *string
    GithubAccessToken   *string
    VMUrl               *string
    VMInsertUrl         *string
    VMSelectUrl         *string
    BinanceSwitch       *bool
    TimeZone            *string
    PlayWrightSSEServer *string
    PlayWrightSwitch    *bool
    FilecrawlApiKey     *string
    FilePath            *string
    GoogleMapApiKey     *string
    NotionAuthorization *string
    NotionVersion       *string

    AllTools *string

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
    BinanceSwitch = flag.Bool("binance_switch", false, "binance switch")
    TimeZone = flag.String("time_zone", "", "time zone")
    PlayWrightSwitch = flag.Bool("play_wright_switch", false, "playwright switch")
    PlayWrightSSEServer = flag.String("play_wright_sse_server", "", "playwright sw server")
    FilecrawlApiKey = flag.String("filecrawl_api_key", "", "filecrawl_api_key")
    FilePath = flag.String("file_path", "", "file path")
    GoogleMapApiKey = flag.String("google_map_api_key", "", "google_map_api_key")
    NotionAuthorization = flag.String("notion_authorization", "", "notion_authorization")
    NotionVersion = flag.String("notion_version", "", "notion_version")

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

    logger.Info("TOOLS_CONF", "AmapApiKey", *AmapApiKey)
    logger.Info("TOOLS_CONF", "AmapTools", *AllTools)
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

}

func InitTools() {
    ctx := context.Background()

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

    if *BinanceSwitch {
        mcpParams = append(mcpParams, binance.InitBinanceMCPClient(&binance.BinanceParam{},
            "", nil, nil, nil))
    }

    if *TimeZone != "" {
        mcpParams = append(mcpParams, mcp_time.InitTimeMCPClient(&mcp_time.TimeParma{
            LocalTimezone: *TimeZone,
        }, "", nil, nil, nil))
    }

    if *PlayWrightSwitch {
        if *PlayWrightSSEServer != "" {
            mcpParams = append(mcpParams, playwright.InitPlaywrightSSEMCPClient(*PlayWrightSSEServer,
                nil, "", nil, nil, nil))
        } else {
            mcpParams = append(mcpParams, playwright.InitPlaywrightMCPClient(&playwright.PlaywrightParam{},
                "", nil, nil, nil))
        }
    }

    if *FilePath != "" {
        mcpParams = append(mcpParams, filesystem.InitFilesystemMCPClient(&filesystem.FilesystemParam{
            Paths: strings.Split(*FilePath, ","),
        }, "", nil, nil, nil))
    }

    if *GoogleMapApiKey != "" {
        mcpParams = append(mcpParams, googlemap.InitGooglemapMCPClient(&googlemap.GoogleMapParam{
            GooglemapApiKey: *GoogleMapApiKey,
        }, "", nil, nil, nil))
    }

    if *NotionAuthorization != "" && *NotionVersion != "" {
        mcpParams = append(mcpParams, notion.InitNotionMCPClient(&notion.NotionParam{
            NotionVersion: *NotionVersion,
            Authorization: *NotionAuthorization,
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

    if *BinanceSwitch {
        InsertTools(binance.NpxBinanceMcpServer, allTools)
    }

    if *TimeZone != "" {
        InsertTools(mcp_time.UvTimeMcpServer, allTools)
    }

    if *PlayWrightSwitch {
        if *PlayWrightSSEServer != "" {
            InsertTools(playwright.SsePlaywrightMcpServer, allTools)
        } else {
            InsertTools(playwright.NpxPlaywrightMcpServer, allTools)
        }
    }

    if *FilePath != "" {
        InsertTools(filesystem.NpxFilesystemMcpServer, allTools)
    }

    if *GoogleMapApiKey != "" {
        InsertTools(googlemap.NpxGooglemapMcpServer, allTools)
    }

    if *NotionAuthorization != "" && *NotionVersion != "" {
        InsertTools(notion.NpxNotionMcpServer, allTools)
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
