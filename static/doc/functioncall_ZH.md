## 功能调用说明

telegram-deepseek-bot 通过 https://github.com/yincongcyincong/mcp-client-go 与 mcp 客户端集成  
自动请求 mcp 服务器并获取 mcp 数据，用于与 deepseek 交互

### 支持的服务

| mcp 客户端          | 	功能说明         | 环境变量配置                                                     |
|------------------|---------------|------------------------------------------------------------|
| 高德地图(amap)       | 高德地图 mcp 服务   | AMAP_API_KEY: 高德地图访问令牌                                     |
| GitHub           | GitHub mcp 服务 | GITHUB_ACCESS_TOKEN: 使用 GitHub 的 access token              |
| Victoria Metrics | VM 指标服务       | VMUrl: 单节点 VM 地址, VMInsertUrl: 集群写入地址, VMSelectUrl: 集群查询地址 |

### 使用说明

1. **高德地图服务**：
    - 需要配置 `AMAP_API_KEY` 环境变量
    - 提供地理编码、逆地理编码等服务

2. **GitHub 服务**：
    - 配置 `GITHUB_ACCESS_TOKEN` 环境变量
    - 提供仓库信息、用户数据等查询

3. **Victoria Metrics**：
    - 支持单节点和集群模式
    - 单节点模式只需配置 `VMUrl`
    - 集群模式需要分别配置写入(`VMInsertUrl`)和查询(`VMSelectUrl`)地址

