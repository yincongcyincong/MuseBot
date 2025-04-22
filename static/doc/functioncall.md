## Function Call Specification

The telegram-deepseek-bot integrates with the MCP client
via [https://github.com/yincongcyincong/mcp-client-go](https://github.com/yincongcyincong/mcp-client-go) to
automatically request MCP server data and interact with DeepSeek.

### Supported Services

| MCP Client       | Description             | Environment Variables                                                                     |
|------------------|-------------------------|-------------------------------------------------------------------------------------------|
| AMAP             | AMAP MCP Service        | AMAP_API_KEY: AMAP access token                                                           |
| GitHub           | GitHub MCP Service      | GITHUB_ACCESS_TOKEN: GitHub access token                                                  |
| Victoria Metrics | VM Metrics Service      | VMUrl: Single-node VM URL, VMInsertUrl: Cluster write URL, VMSelectUrl: Cluster query URL |
| Time             | Time MCP Service        | TIME_ZONE: local time zone                                                                |
| Binance          | Binance MCP Service     | BINANCE_SWITCH                                                                            |
| Play Wright      | Play Wright MCP Service | PLAY_WRIGHT_SWITCH                                                                        |
| File System      | File MCP Service        | FILE_PATH: multi computer directory, split by ','                                         |
| FILE CRAWL       | FILECRAWL MCP Service   | FILECRAWL_API_KEY                                                                         |


### Usage Instructions

1. **AMAP Service**:
    - Requires configuring the `AMAP_API_KEY` environment variable
    - Provides geocoding, reverse geocoding, IP location, and route planning services

2. **GitHub Service**:
    - Requires configuring the `GITHUB_ACCESS_TOKEN` environment variable
    - Supports repository information, user profile queries, commit history, and organization data

3. **Victoria Metrics Service**:
    - Supports both single-node and cluster modes
    - Single-node mode requires `VMUrl` configuration
    - Cluster mode requires separate configuration of write (`VMInsertUrl`) and query (`VMSelectUrl`) URLs

4. **Time Service**:
    - Requires setting `TIME_ZONE` environment variable (e.g., `Asia/Shanghai`, `UTC`)
    - Returns the current local time based on the configured timezone
    - Useful for time-based automation and context-aware responses

5. **Binance Service**:
    - Requires enabling `BINANCE_SWITCH` (e.g., `true`)
    - Fetches real-time cryptocurrency data, including prices, tickers, and volume
    - Supports querying by symbol, e.g., BTC, ETH, etc.

6. **Play Wright Service**:
    - Requires enabling `PLAY_WRIGHT_SWITCH` (e.g., `true`)
    - Allows automated browser interactions such as scraping, screenshots, or headless browsing
    - Suitable for automating interaction with web content not accessible via API

7. **File System Service**:
    - Requires setting `FILE_PATH` environment variable with comma-separated paths (e.g., `/mnt/data1,/mnt/data2`)
    - Enables reading from local or network-mounted directories
    - Supports searching, reading, or listing files for multi-device setups

8. **FILE CRAWL Service**:
    - Requires setting `FILECRAWL_API_KEY`
    - Enables crawling and indexing of files from given paths or URLs
    - Useful for building searchable file databases or document retrieval systems
