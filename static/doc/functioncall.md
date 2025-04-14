## Function Call Specification

The telegram-deepseek-bot integrates with the MCP client via [https://github.com/yincongcyincong/mcp-client-go](https://github.com/yincongcyincong/mcp-client-go) to automatically request MCP server data and interact with DeepSeek.

### Supported Services

| MCP Client         | Description               | Environment Variables                                                                 |
|--------------------|---------------------------|--------------------------------------------------------------------------------------|
| AMAP               | AMAP MCP Service          | AMAP_API_KEY: AMAP access token                                                      |
| GitHub             | GitHub MCP Service        | GITHUB_ACCESS_TOKEN: GitHub access token                                             |
| Victoria Metrics   | VM Metrics Service        | VMUrl: Single-node VM URL, VMInsertUrl: Cluster write URL, VMSelectUrl: Cluster query URL |

### Usage Instructions

1. **AMAP Service**:
    - Requires configuring the `AMAP_API_KEY` environment variable
    - Provides geocoding, reverse geocoding and other location-based services

2. **GitHub Service**:
    - Requires configuring the `GITHUB_ACCESS_TOKEN` environment variable
    - Supports repository information, user data queries and other features

3. **Victoria Metrics Service**:
    - Supports both single-node and cluster modes
    - Single-node mode only requires `VMUrl` configuration
    - Cluster mode requires separate configuration of write (`VMInsertUrl`) and query (`VMSelectUrl`) URLs
