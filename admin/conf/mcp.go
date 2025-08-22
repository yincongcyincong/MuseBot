package conf

import "github.com/yincongcyincong/mcp-client-go/clients/param"

var MCPConf = &param.McpClientGoConfig{
	McpServers: map[string]*param.MCPConfig{
		"playwright": {
			Description: "Simulates browser behavior for tasks like web navigation, data scraping, and automated interactions with web pages.",
			Url:         "http://localhost:8931/sse",
		},
		"filesystem": {
			Command:     "npx",
			Description: "supports file operations such as reading, writing, deleting, renaming, moving, and listing files and directories.\n",
			Args: []string{
				"-y", "@modelcontextprotocol/server-filesystem", "/path/to/your/directory",
			},
		},
		"chrome-mcp-server": {
			Description: "Operate Chrome to perform different interactive operations.",
			Url:         "http://127.0.0.1:12306/mcp",
		},
		"mcp-server-commands": {
			Description: " execute local system commands through a backend service.",
			Command:     "npx",
			Args:        []string{"mcp-server-commands"},
		},
		"mysql": {
			Description: "manage MySQL server",
			Command:     "npx",
			Args:        []string{"-y", "@benborla29/mcp-server-mysql"},
			Env: map[string]string{
				"MYSQL_HOST":             "127.0.0.1",
				"MYSQL_PORT":             "3306",
				"MYSQL_USER":             "test",
				"MYSQL_PASS":             "test",
				"MYSQL_DB":               "test",
				"ALLOW_INSERT_OPERATION": "true",
				"ALLOW_UPDATE_OPERATION": "true",
				"ALLOW_DELETE_OPERATION": "true",
				"ALLOW_DDL_OPERATION":    "true",
			},
		},
		"binance": {
			Description: "get encrypt currency information from binance.",
			Command:     "npx",
			Args:        []string{"-y", "@snjyor/binance-mcp@latest"},
		},
		"github": {
			Description: "manage Github base on GitHub API.",
			Command:     "docker",
			Args:        []string{"run", "-i", "--rm", "-e", "GITHUB_PERSONAL_ACCESS_TOKEN", "ghcr.io/github/github-mcp-server"},
			Env: map[string]string{
				"GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>",
			},
		},
		"fetch": {
			Command:     "uvx",
			Args:        []string{"mcp-server-fetch"},
			Description: "Web content crawling",
		},
		"amap": {
			Command: "npx",
			Args:    []string{"-y", "@amap/amap-maps-mcp-server"},
			Env: map[string]string{
				"AMAP_MAPS_API_KEY": "",
			},
			Description: "get geo info from amap service",
		},
		"mcp-server-alipay": {
			Command: "npx",
			Args:    []string{"-y", "@alipay/mcp-server-alipay"},
			Env: map[string]string{
				"AP_APP_ID":     "2014...222",
				"AP_APP_KEY":    "MIIE...DZdM=",
				"AP_PUB_KEY":    "MIIB...DAQAB",
				"AP_NOTIFY_URL": "https://your-own-server",
				"AP_RETURN_URL": "https://success-page",
				"...其他参数":   "...其他值", // Note: Key from original data
			},
			Description: "request alipay payment api",
		},
		"wuying": {
			Command: "npx",
			Args:    []string{"-y", "wuying-agentbay-mcp-server"},
			Env: map[string]string{
				"APIKEY": "APIKEY",
			},
			Description: "AI Agent cloud service api",
		},
		"sequential-thinking": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
			Description: "use sequential thinking to cognize process of approaching problems or tasks in a step-by-step, logical order",
		},
		"xiyan": {
			Command: "python",
			Args:    []string{"-m", "xiyan_mcp_server"},
			Env: map[string]string{
				"YML": "PATH/TO/YML",
			},
			Description: "enables natural language queries to databases",
		},
		"baidu-map": {
			Command: "uvx",
			Args:    []string{"mcp-server-baidu-maps"},
			Env: map[string]string{
				"BAIDU_MAPS_API_KEY": "<YOUR_API_KEY>",
			},
			Description: "get geo info from baidu map service",
		},
		"tavily": {
			Command: "npx",
			Args:    []string{"-y", "tavily-mcp@0.1.4"},
			Env: map[string]string{
				"TAVILY_API_KEY": "your-api-key-here",
			},
			Disabled:    false,
			Description: "search engine optimized for LLMs and RAG, aimed at efficient, quick and persistent search result",
		},
		"oceanbase": {
			Command: "uvx",
			Args:    []string{"oceanbase_mcp_server"},
			Env: map[string]string{
				"OB_HOST":     "localhost",
				"OB_PORT":     "2881",
				"OB_USER":     "your_username",
				"OB_DATABASE": "your_database",
				"OB_PASSWORD": "your_password",
			},
			Description: "manage OceanBase",
		},
		"minimax": {
			Command: "uvx",
			Args:    []string{"minimax-mcp"},
			Env: map[string]string{
				"MINIMAX_API_KEY":       "<insert-your-api-key-here>",
				"MINIMAX_API_HOST":      "https://api.minimaxi.chat",
				"MINIMAX_MCP_BASE_PATH": "<local-output-dir-path>",
			},
			Description: "interaction with powerful Text to Speech and video/image generation APIs. ",
		},
		"edgeone": {
			Command:     "npx",
			Args:        []string{"edgeone-pages-mcp"},
			Description: "deploying HTML content, folder, and zip file to EdgeOne Pages and obtaining a publicly accessible URL.",
		},
		"time": {
			Command:     "uvx",
			Args:        []string{"mcp-server-time"},
			Description: "Processing time related functions",
		},
		"gitlab": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-gitlab"},
			Env: map[string]string{
				"GITLAB_API_URL":               "https://gitlab.com/api/v4",
				"GITLAB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>",
			},
			Description: "manage Gitlab base on Gitlab API.",
		},
		"flomo": {
			Command: "npx",
			Args:    []string{"-y", "@chatmcp/mcp-server-flomo"},
			Env: map[string]string{
				"FLOMO_API_URL": "https://flomoapp.com/iwh/xxx/xxx/",
			},
			Description: "write notes to Flomo",
		},
		"firecrawl": {
			Command: "npx",
			Args:    []string{"-y", "firecrawl-mcp"},
			Env: map[string]string{
				"FIRECRAWL_API_KEY": "YOUR_API_KEY_HERE",
			},
			Description: "Enhance your AI agents with Firecrawl's web scraping capabilities to extract data from any website.",
		},
		"perplexity": {
			Command: "npx",
			Args:    []string{"-y", "server-perplexity-ask"},
			Env: map[string]string{
				"PERPLEXITY_API_KEY": "YOUR_API_KEY_HERE",
			},
			Description: "allows any AI-powered tool to perform real-time, web-wide research using Perplexity's powerful search engine.",
		},
		"exa": {
			Command: "npx",
			Args:    []string{"exa-mcp-server"},
			Env: map[string]string{
				"EXA_API_KEY": "your-api-key-here",
			},
			Description: "lets AI assistants like Claude use the Exa AI Search API for web searches",
		},
		"figma": {
			Command: "npx",
			Args: []string{
				"-y",
				"figma-developer-mcp",
				"--figma-api-key=FIGMA_API_ACCESS_TOKEN",
				"--stdio",
			},
			Description: "access to your Figma files with this Model Context Protocol server.",
		},
		"blender": {
			Command:     "uvx",
			Args:        []string{"blender-mcp"},
			Description: " allowing AI to directly interact with and control Blender.",
		},
		"slack": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-slack"},
			Env: map[string]string{
				"SLACK_TEAM_ID":   "T01234567",
				"SLACK_BOT_TOKEN": "xoxb-your-bot-token",
			},
			Description: "DMs, Group DMs, Smart History fetch from slack",
		},
		"AbletonMCP": {
			Command:     "uvx",
			Args:        []string{"ableton-mcp"},
			Description: "allowing AI to directly interact with and control Ableton Live.",
		},
		"redis": {
			Command: "npx",
			Args: []string{
				"-y",
				"@modelcontextprotocol/server-redis",
				"redis://localhost:6379",
			},
			Description: "The Redis MCP Server is a natural language interface designed for agentic applications to efficiently manage and search data in Redis.",
		},
		"browserbase": {
			Command: "node",
			Args:    []string{"path/to/mcp-server-browserbase/browserbase/dist/index.js"},
			Env: map[string]string{
				"BROWSERBASE_API_KEY":    "<YOUR_BROWSERBASE_API_KEY>",
				"BROWSERBASE_PROJECT_ID": "<YOUR_BROWSERBASE_PROJECT_ID>",
			},
			Description: "The Model Context Protocol (MCP) is an open protocol that enables seamless integration between LLM applications and external data sources and tools.",
		},
		"magic": {
			Command: "npx",
			Args: []string{
				"-y",
				"@smithery/cli@latest",
				"install",
				"@21st-dev/magic-mcp",
				"--client",
				"windsurf",
			},
			Env: map[string]string{
				"TWENTY_FIRST_API_KEY": "your-api-key",
			},
			Description: "helps developers create beautiful, modern UI components instantly through natural language descriptions. ",
		},
		"everything": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-everything"},
			Description: "This MCP server attempts to exercise all the features of the MCP protocol. It is not intended to be a useful server, but rather a test server for builders of MCP clients. It implements prompts, tools, resources, sampling, and more to showcase MCP capabilities. ",
		},
		"gdrive": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-gdrive"},
			Description: " integrates with Google Drive to allow listing, reading, and searching files, as well as the ability to read and write to Google Sheets.",
		},
		"memory": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
			Description: "A basic implementation of persistent memory using a local knowledge graph. This lets AI remember information about the user across chats.",
		},
		"brave-search": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-brave-search"},
			Env: map[string]string{
				"BRAVE_API_KEY": "YOUR_API_KEY_HERE",
			},
			Description: " integrates the Brave Search API, providing, Web Search, Local Points of Interest Search, Image Search, Video Search and News",
		},
		"sentry": {
			Command: "uvx",
			Args: []string{
				"mcp-server-sentry",
				"--auth-token",
				"YOUR_SENTRY_TOKEN",
			},
			Description: "The Sentry MCP Server provides a secure way of bringing Sentry's full issue context into systems",
		},
		"google-maps": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-google-maps"},
			Env: map[string]string{
				"GOOGLE_MAPS_API_KEY": "<YOUR_API_KEY>",
			},
			Description: "使用 npx 启动 Google Maps MCP，依赖环境变量中的 API 密钥访问 Google 地图服务",
		},
		"puppeteer": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-puppeteer"},
			Description: "providing comprehensive Google Maps API integration with streamable HTTP transport support and LLM processing",
		},
		"chatsum": {
			Command: "path-to/bin/node",
			Args:    []string{"path-to/mcp-server-chatsum/build/index.js"},
			Env: map[string]string{
				"CHAT_DB_PATH": "path-to/mcp-server-chatsum/chatbot/data/chat.db",
			},
			Description: "summarize your chat messages",
		},
		"everart": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-everart"},
			Env: map[string]string{
				"EVERART_API_KEY": "your_key_here",
			},
			Description: "integrates with EverArt's AI models to generate both vector and raster images.",
		},
		"arxiv": {
			Command: "uv",
			Args: []string{
				"tool",
				"run",
				"arxiv-mcp-server",
				"--storage-path",
				"/path/to/paper/storage",
			},
			Description: "provides a bridge between AI assistants and arXiv's research repository",
		},
		"idb": {
			Command:     "npx",
			Args:        []string{"-y", "@noahlozevski/idb"},
			Description: " provides integration between MCP (Model Context Protocol) and Facebook's idb (iOS Development Bridge), enabling automated iOS device management.",
		},
		"graphlit": {
			Command: "npx",
			Args:    []string{"-y", "graphlit-mcp-server"},
			Env: map[string]string{
				"JIRA_EMAIL":                   "your-jira-email",
				"JIRA_TOKEN":                   "your-jira-token",
				"TWITTER_TOKEN":                "your-twitter-token",
				"LINEAR_API_KEY":               "your-linear-api-key",
				"NOTION_API_KEY":               "your-notion-api-key",
				"SLACK_BOT_TOKEN":              "your-slack-bot-token",
				"DISCORD_BOT_TOKEN":            "your-discord-bot-token",
				"NOTION_DATABASE_ID":           "your-notion-database-id",
				"GRAPHLIT_JWT_SECRET":          "your-jwt-secret",
				"GOOGLE_EMAIL_CLIENT_ID":       "your-google-client-id",
				"GRAPHLIT_ENVIRONMENT_ID":      "your-environment-id",
				"GRAPHLIT_ORGANIZATION_ID":     "your-organization-id",
				"GOOGLE_EMAIL_CLIENT_SECRET":   "your-google-client-secret",
				"GOOGLE_EMAIL_REFRESH_TOKEN":   "your-google-refresh-token",
				"GITHUB_PERSONAL_ACCESS_TOKEN": "your-github-pat",
			},
			Description: "enables integration between MCP clients and the Graphlit service. This document outlines the setup process.",
		},
		"xhs": {
			Command: "uvx",
			Args:    []string{"xhs_mcp_server@latest"},
			Env: map[string]string{
				"phone":     "YOUR_PHONE_NUMBER",
				"json_path": "PATH_TO_STORE_YOUR_COOKIES",
			},
			Description: "manage xiaohonshu(redNote) based on ai",
		},
		"awslabs.core-mcp-server": {
			Description: "As a core component of the MCP framework, this server provides foundational services and general support for other specific AWS-related MCP servers.",
			Command:     "uvx",
			Args:        []string{"awslabs.core-mcp-server@latest"},
			Env: map[string]string{
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
		},
		"awslabs.nova-canvas-mcp-server": {
			Description: "This server is closely linked to AWS's AI and Machine Learning capabilities, specifically integrating with the Amazon Nova Canvas tool to enable canvas-related interactions or data processing within generative AI applications.",
			Command:     "uvx",
			Args:        []string{"awslabs.nova-canvas-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":       "your-aws-profile",
				"AWS_REGION":        "us-east-1",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
		},
		"awslabs.bedrock-kb-retrieval-mcp-server": {
			Description: "Focused on the AWS AI and Machine Learning domain, this server deeply integrates with Amazon Bedrock Knowledge Base Retrieval functionality, allowing foundation models to access and retrieve information hosted within Bedrock knowledge bases.",
			Command:     "uvx",
			Args:        []string{"awslabs.bedrock-kb-retrieval-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":       "your-aws-profile",
				"AWS_REGION":        "us-east-1",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
		},
		"awslabs.aws-pricing-mcp-server": {
			Description: "This server is designed to provide real-time access to AWS service pricing information, helping users query or understand the cost details of various AWS services.",
			Command:     "uvx",
			Args:        []string{"awslabs.aws-pricing-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":       "your-aws-profile",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
		},
		"awslabs.cdk-mcp-server": {
			Description: "Closely coupled with AWS Cloud Development Kit (CDK) as part of the infrastructure and deployment tools, this server enables models to understand, generate, or assist with Infrastructure as Code operations using CDK for AWS.",
			Command:     "uvx",
			Args:        []string{"awslabs.cdk-mcp-server@latest"},
			Env: map[string]string{
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
		},
		"awslabs.aws-documentation-mcp-server": {
			Description: "This server provides real-time access to official AWS documentation, empowering foundation models to retrieve the latest AWS service information, technical guides, and best practices, thereby reducing hallucinations and improving accuracy.",
			Command:     "uvx",
			Args:        []string{"awslabs.aws-documentation-mcp-server@latest"},
			Env: map[string]string{
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.lambda-tool-mcp-server": {
			Description: "Integrating AWS Lambda tool functionalities, this server allows models to assist in managing, deploying, or interacting with AWS Lambda functions, simplifying serverless application development and operations.",
			Command:     "uvx",
			Args:        []string{"awslabs.lambda-tool-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":        "your-aws-profile",
				"AWS_REGION":         "us-east-1",
				"FUNCTION_PREFIX":    "your-function-prefix",
				"FUNCTION_LIST":      "your-first-function, your-second-function",
				"FUNCTION_TAG_KEY":   "your-tag-key",
				"FUNCTION_TAG_VALUE": "your-tag-value",
			},
		},
		"awslabs.terraform-mcp-server": {
			Description: "This server, as an infrastructure and deployment tool, is tightly integrated with Terraform, enabling models to understand, generate, or assist in managing and deploying AWS resources using Terraform scripts.",
			Command:     "uvx",
			Args:        []string{"awslabs.terraform-mcp-server@latest"},
			Env: map[string]string{
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.frontend-mcp-server": {
			Description: "This server focuses on frontend development tools and support, providing foundation models with knowledge and capabilities related to frontend technologies, frameworks, or deployment processes within the AWS cloud.",
			Command:     "uvx",
			Args:        []string{"awslabs.frontend-mcp-server@latest"},
			Env: map[string]string{
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.valkey-mcp-server": {
			Description: "This server integrates with MemoryDB for Valkey, an AWS data and analytics tool, allowing models to interact with Valkey databases, such as performing data queries or operations.",
			Command:     "uvx",
			Args:        []string{"awslabs.valkey-mcp-server@latest"},
			Env: map[string]string{
				"VALKEY_HOST":       "127.0.0.1",
				"VALKEY_PORT":       "6379",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.aws-location-mcp-server": {
			Description: "This server is designed to provide access to AWS Location services, enabling models to process geospatial data, maps, or leverage AWS's location-based service capabilities.",
			Command:     "uvx",
			Args:        []string{"awslabs.aws-location-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":       "your-aws-profile",
				"AWS_REGION":        "us-east-1",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.memcached-mcp-server": {
			Description: "This server integrates with ElastiCache for Memcached within AWS's data and analytics tools, allowing models to interact with Memcached caching systems for high-performance data access.",
			Command:     "uvx",
			Args:        []string{"awslabs.memcached-mcp-server@latest"},
			Env: map[string]string{
				"MEMCACHED_HOST":    "127.0.0.1",
				"MEMCACHED_PORT":    "11211",
				"FASTMCP_LOG_LEVEL": "ERROR",
			},
			Disabled: false,
		},
		"awslabs.git-repo-research-mcp-server": {
			Description: "Related to developer tools and support, this server focuses on Git repository research, enabling models to analyze code repositories hosted on platforms like GitHub, extract information, or assist with code reviews, potentially involving AWS-related codebases.",
			Command:     "uvx",
			Args:        []string{"awslabs.git-repo-research-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE":       "your-aws-profile",
				"AWS_REGION":        "us-east-1",
				"FASTMCP_LOG_LEVEL": "ERROR",
				"GITHUB_TOKEN":      "your-github-token",
			},
			Disabled: false,
		},
		"awslabs.cloudformation": {
			Description: "This server integrates with AWS CloudFormation, serving as an infrastructure and deployment tool that allows models to understand, generate, or assist in using CloudFormation templates to define and deploy resource stacks on the AWS cloud.",
			Command:     "uvx",
			Args:        []string{"awslabs.cfn-mcp-server@latest"},
			Env: map[string]string{
				"AWS_PROFILE": "your-aws-profile",
			},
			Disabled: false,
		},
	},
}
