## 配置 `MCP_CONF_PATH` 环境变量以使用 MCP 服务器 (Go 二进制文件)

本文档将指导您如何在 `telegram-deepseek-bot` 项目中使用 Go 二进制文件时配置 `MCP_CONF_PATH` 环境变量，以便使用自定义的 MCP 服务器配置。

### 1. 创建 MCP 配置文件

首先，您需要创建一个 JSON 文件来定义您的 MCP 服务器。以下是一个示例配置，其中包含了 GitHub、Playwright、高德地图 (amap-mcp-server) 和高德地图 (amap-maps) 的 MCP 服务器设置。您可以根据您的需求修改此文件。

```json
{
    "mcpServers": {
       "github": {
          "command": "docker",
          "description": "执行 Git 操作并与 GitHub 集成，用于管理仓库、拉取请求、问题和工作流。",
          "args": [
             "run",
             "-i",
             "--rm",
             "-e",
             "GITHUB_PERSONAL_ACCESS_TOKEN",
             "ghcr.io/github/github-mcp-server"
          ],
          "env": {
             "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>"
          }
       },
       "playwright": {
          "description": "模拟浏览器行为，用于网页导航、数据抓取和网页自动化交互等任务。",
          "url": "http://localhost:8931/sse"
       },
       "amap-mcp-server": {
          "description": "提供地理服务，如位置查找、路线规划和地图导航。",
          "url": "http://localhost:8000/mcp"
       },
       "amap-maps": {
          "command": "npx",
          "description": "提供地理服务，如位置查找、路线规划和地图导航。",
          "args": [
             "-y",
             "@amap/amap-maps-mcp-server"
          ],
          "env": {
             "AMAP_MAPS_API_KEY": "<YOUR_TOKEN>"
          }
       }
    }
}
```

**请注意：**

* 将 **<YOUR\_TOKEN>** 替换为您的实际 GitHub 个人访问令牌和高德地图 API 密钥。
* `amap-mcp-server` 和 `playwright` MCP 服务器的 `url` 字段指向本地运行的服务。您需要确保这些服务正在运行并可访问。
* 您可以根据需要添加或删除 MCP 服务器配置。

将此文件保存到您选择的目录，例如，您可以将其命名为 **mcp\_config.json** 并放置在项目根目录下。

### 2. 设置 `MCP_CONF_PATH` 环境变量

对于 Go 二进制文件，设置 `MCP_CONF_PATH` 环境变量的方法与 Python 脚本类似，主要是在运行二进制文件之前，确保环境变量已被正确设置。

#### 方法一：直接在命令行中设置 (临时)

如果您只是想临时运行二进制文件并测试配置，可以在运行之前在命令行中设置环境变量：

**Linux/macOS:**

```bash
export MCP_CONF_PATH=/path/to/your/mcp_config.json
./telegram-deepseek-bot # 假设这是您的 Go 二进制文件
```

#### 方法二：在 Docker Compose 或 Dockerfile 中设置 (如果您使用 Docker)

如果您通过 Docker 部署 `telegram-deepseek-bot`，您可以在文件 `Dockerfile` 中设置环境变量。


**在 `Dockerfile` 中设置：**

```dockerfile
# ... 其他 Dockerfile 指令 ...

COPY mcp_config.json /app/mcp_config.json

ENV MCP_CONF_PATH /app/mcp_config.json

# ... 其他 Dockerfile 指令 ...
```

### 3. 运行 `telegram-deepseek-bot` Go 二进制文件

在设置好 `MCP_CONF_PATH` 环境变量后，您可以正常运行 `telegram-deepseek-bot` 的 Go 二进制文件。项目将加载您指定的 MCP 配置文件，并能够使用其中定义的 MCP 服务器。

例如：

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-use_tools=true
```

现在，您的 `telegram-deepseek-bot` 应该能够与您配置的 MCP 服务器进行交互。
---
