## 功能调用规范（Function Call Specification）

`telegram-deepseek-bot` 通过集成 MCP 客户端（[https://github.com/yincongcyincong/mcp-client-go](https://github.com/yincongcyincong/mcp-client-go)）自动请求 MCP 服务端数据并与 DeepSeek 进行交互。

### 支持的服务

| MCP 客户端          | 描述            | 环境变量                                                                                               |
| ---------------- | ------------- | -------------------------------------------------------------------------------------------------- |
| AMAP             | 高德地图 MCP 服务   | `AMAP_API_KEY`: 高德地图访问令牌                                                                           |
| GitHub           | GitHub MCP 服务 | `GITHUB_ACCESS_TOKEN`: GitHub 访问令牌                                                                 |
| Victoria Metrics | VM 指标服务       | `VMUrl`: 单节点地址，`VMInsertUrl`: 集群写入地址，`VMSelectUrl`: 集群查询地址                                         |
| Time             | 时间服务          | `TIME_ZONE`: 本地时区                                                                                  |
| Binance          | 币安服务          | `BINANCE_SWITCH`: 启用开关                                                                             |
| Play Wright      | 浏览器自动化服务      | `PLAY_WRIGHT_SWITCH`: 启用开关                                                                         |
| File System      | 文件系统服务        | `FILE_PATH`: 多设备路径，使用英文逗号分隔                                                                        |
| File Crawl       | 文件爬取服务        | `FILECRAWL_API_KEY`: 文件抓取的 API 密钥                                                                  |
| GoogleMap        | 谷歌地图服务        | `GOOGLE_MAP_API_KEY`: 谷歌地图访问密钥                                                                     |
| Notion           | Notion 服务     | `NOTION_AUTHORIZATION`, `NOTION_VERSION`: 授权令牌与 API 版本                                             |
| Aliyun           | 阿里云服务         | `ALIYUN_ACCESS_KEY_ID`, `ALIYUN_ACCESS_KEY_SECRET`: 阿里云密钥对                                         |
| Airbnb           | Airbnb 服务     | `AIRBNB_SWITCH`: 启用开关                                                                              |
| Twitter          | 推特服务          | `TWITTER_API_KEY`, `TWITTER_API_KEY_SECRET`, `TWITTER_ACCESS_TOKEN`, `TWITTER_ACCESS_TOKEN_SECRET` |
| Bitcoin          | 比特币服务         | `BITCOIN_SWITCH`: 启用开关                                                                             |
| Whatsapp         | Whatsapp 服务   | `WHATSAPP_PATH`, `WHATSAPP_PYTHON_MAIN_FILE`: 服务路径与主程序路径                                           |

### 使用说明

1. **AMAP 服务**：

    * 需要设置 `AMAP_API_KEY`
    * 支持地理编码、逆地理编码、IP定位、路径规划等功能

2. **GitHub 服务**：

    * 需要设置 `GITHUB_ACCESS_TOKEN`
    * 可获取仓库信息、用户资料、提交历史和组织信息等

3. **Victoria Metrics 服务**：

    * 支持单节点和集群模式
    * 单节点模式需配置 `VMUrl`
    * 集群模式需分别配置写入地址 `VMInsertUrl` 和查询地址 `VMSelectUrl`

4. **Time 服务**：

    * 需配置 `TIME_ZONE`（如 `Asia/Shanghai`、`UTC`）
    * 返回当前本地时间
    * 适用于基于时间的自动化和上下文响应

5. **Binance 服务**：

    * 需设置 `BINANCE_SWITCH` 为 `true` 启用
    * 获取实时加密货币数据，如价格、行情、交易量等
    * 支持根据币种（如 BTC, ETH）查询

6. **Play Wright 服务**：

    * 需设置 `PLAY_WRIGHT_SWITCH` 为 `true` 启用
    * 支持自动化浏览器操作，如抓取、截图、无头浏览等
    * 适用于无法通过 API 获取数据的网页交互

7. **File System 服务**：

    * 需设置 `FILE_PATH`（如 `/mnt/data1,/mnt/data2`）
    * 支持从本地或挂载路径读取文件
    * 可用于搜索、读取、列出多设备上的文件

8. **File CRAWL 服务**：

    * 需设置 `FILECRAWL_API_KEY`
    * 可从路径或 URL 抓取并索引文件
    * 适合构建可搜索的文档数据库或检索系统

9. **GoogleMap 服务**：

    * 需设置 `GOOGLE_MAP_API_KEY`
    * 支持地理编码、地点搜索、导航等功能
    * 适用于基于位置的应用与地图服务整合

10. **Notion 服务**：

    * 需设置 `NOTION_AUTHORIZATION` 和 `NOTION_VERSION`
    * 支持页面、数据库、块等的读写操作
    * 可实现自动化页面创建、内容更新、与其他系统的数据同步

11. **Aliyun 服务**：

    * 需设置 `ALIYUN_ACCESS_KEY_ID` 与 `ALIYUN_ACCESS_KEY_SECRET`
    * 可使用阿里云计算、存储、消息等服务
    * 适合接入阿里云提供的各类云服务功能

12. **Airbnb 服务**：

    * 需设置 `AIRBNB_SWITCH`（如 `true`）
    * 接入 Airbnb 数据或服务
    * 适合自动同步房源信息或管理预订工作流

13. **Twitter 服务**：

    * 需设置所有 Twitter API 凭证
    * 支持发推文、获取时间线、管理账号等操作
    * 适合构建推特机器人、数据分析工具或自动化平台

14. **Bitcoin 服务**：

    * 需设置 `BITCOIN_SWITCH`（如 `true`）
    * 可进行交易追踪、钱包监控、区块链查询等
    * 适用于涉及比特币网络或加密应用的场景

15. **Whatsapp 服务**：

    * 需设置 `WHATSAPP_PATH` 与 `WHATSAPP_PYTHON_MAIN_FILE`
    * 实现自动消息发送、聊天机器人、通知功能
    * 可用于客服自动化、通知推送、个人助手等场景

---
