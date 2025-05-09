package conf

import (
    "context"
    "flag"
    "github.com/cohesion-org/deepseek-go"
    "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
    "github.com/yincongcyincong/mcp-client-go/clients"
    "github.com/yincongcyincong/mcp-client-go/clients/airbnb"
    "github.com/yincongcyincong/mcp-client-go/clients/aliyun"
    "github.com/yincongcyincong/mcp-client-go/clients/amap"
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
    "os"
    "strings"
    "time"
)

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
    logger.Info("TOOLS_CONF", "AliyunAccessKeyID", *AliyunAccessKeyID)
    logger.Info("TOOLS_CONF", "AliyunAccessKeySecret", *AliyunAccessKeySecret)
    logger.Info("TOOLS_CONF", "AirBnbSwitch", *AirBnbSwitch)
    logger.Info("TOOLS_CONF", "BinanceSwitch", *BinanceSwitch)
    logger.Info("TOOLS_CONF", "TwitterApiKey", *TwitterApiKey)
    logger.Info("TOOLS_CONF", "TwitterApiSecretKey", *TwitterApiSecretKey)
    logger.Info("TOOLS_CONF", "TwitterAccessTokenSecret", *TwitterAccessTokenSecret)
    logger.Info("TOOLS_CONF", "TwitterAccessToken", *TwitterAccessToken)

}

func InitTools() {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

    if *AliyunAccessKeyID != "" && *AliyunAccessKeySecret != "" {
        mcpParams = append(mcpParams, aliyun.InitAliyunMCPClient(&aliyun.AliyunParams{
            AliyunAccessKeyID:     *AliyunAccessKeySecret,
            AliyunAccessKeySecret: *AliyunAccessKeySecret,
        }, "", nil, nil, nil))
    }

    if *AirBnbSwitch {
        mcpParams = append(mcpParams, airbnb.InitAirbnbMCPClient(&airbnb.AirbnbParam{}, "", nil, nil, nil))
    }

    if *BitCoinSwitch {
        mcpParams = append(mcpParams, bitcoin.InitBitcoinMCPClient(&bitcoin.BitcoinParam{}, "", nil, nil, nil))
    }

    if *TwitterApiSecretKey != "" && *TwitterAccessToken != "" && *TwitterAccessTokenSecret != "" && *TwitterApiKey != "" {
        mcpParams = append(mcpParams, twitter.InitTwitterMCPClient(&twitter.TwitterParam{
            ApiKey:            *TwitterApiKey,
            ApiSecretKey:      *TwitterApiSecretKey,
            AccessToken:       *TwitterAccessToken,
            AccessTokenSecret: *TwitterAccessTokenSecret,
        }, "", nil, nil, nil))
    }

    if *WhatsappPath != "" && *WhatsappPythonMainFile != "" {
        mcpParams = append(mcpParams, whatsapp.InitWhatsappMCPClient(&whatsapp.WhaPsAppParam{
            WhatsappPath:   *WhatsappPath,
            PythonMainFile: *WhatsappPythonMainFile,
        },
            "", nil, nil, nil))
    }

    errs := clients.RegisterMCPClient(ctx, mcpParams)
    if len(errs) > 0 {
        for mcpServer, err := range errs {
            logger.Error("register mcp client error", "server", mcpServer, "error", err)
        }
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

    if *AliyunAccessKeyID != "" && *AliyunAccessKeySecret != "" {
        InsertTools(aliyun.UvxAliyunMcpServer, allTools)
    }

    if *AirBnbSwitch {
        InsertTools(airbnb.NpxAirbnbMcpServer, allTools)
    }

    if *BitCoinSwitch {
        InsertTools(bitcoin.NpxBitcoinMcpServer, allTools)
    }

    if *TwitterApiKey != "" && *TwitterApiSecretKey != "" && *TwitterAccessToken != "" && *TwitterAccessTokenSecret != "" {
        InsertTools(twitter.NpxTwitterMcpServer, allTools)
    }

    if *WhatsappPath != "" && *WhatsappPythonMainFile != "" {
        InsertTools(whatsapp.UvWhatsAppMcpServer, allTools)
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
