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
