# DeepSeek Telegram 机器人

本仓库提供了一个基于 **Golang** 构建的 **Telegram 机器人**，集成了 **DeepSeek API**，实现 AI 驱动的智能回复功能。机器人支持 **流式输出**，让交互更加自然流畅。  
[English Documentation](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/README.md)

## 🚀 功能特点

- 🤖 **AI 智能回复**：使用 DeepSeek API 进行智能对话。
- ⏳ **流式输出**：实时发送回复，提升用户体验。
- 🎯 **命令处理**：支持自定义命令。
- 🏗 **轻松部署**：可在本地运行或部署到云服务器。

## 🤖 使用示例

[使用视频](https://github.com/yincongcyincong/telegram-deepseek-bot/wiki/Usage-Video)

## 📌 环境要求

- [Go 1.24+](https://go.dev/dl/)
- [Telegram Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek 授权 Token](https://api-docs.deepseek.com/zh-cn/)

## 📥 安装

1. **克隆仓库**
   ```sh
   git clone https://github.com/yourusername/deepseek-telegram-bot.git
   cd deepseek-telegram-bot
   ```

2. **安装依赖**
   ```sh
   go mod tidy
   ```

3. **配置环境变量**
   ```sh
   export TELEGRAM_BOT_TOKEN="你的 Telegram 机器人 Token"
   export DEEPSEEK_TOKEN="你的 DeepSeek 授权 Token"
   ```

## 🚀 启动

本地运行：

```sh
go run main.go -telegram_bot_token=你的_telegram_token -deepseek_token=你的_deepseek_token
```

使用 Docker：

```sh
docker pull jackyin0822/telegram-deepseek-bot:latest
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="你的_telegram_token" -e DEEPSEEK_TOKEN="你的_deepseek_token" --name my-telegram-bot jackyin0822/telegram-deepseek-bot:latest
```

## ⚙️ 配置

通过环境变量配置机器人：

| 变量名                         | 描述                                                                                                  | 默认值                       |
|--------------------------------|-------------------------------------------------------------------------------------------------------|------------------------------|
| TELEGRAM_BOT_TOKEN (必填)      | Telegram 机器人 Token                                                                                  | -                            |
| DEEPSEEK_TOKEN (必填)          | DeepSeek API 授权 Token                                                                                | -                            |
| CUSTOM_URL                     | 自定义 DeepSeek URL                                                                                     | https://api.deepseek.com/    |
| DEEPSEEK_TYPE                   | DeepSeek 类型（deepseek-r1-250120，doubao-1.5-pro-32k-250115 等）                                       | deepseek                     |
| VOLC_AK                         | Volcengine 图像生成模型的 AK [文档](https://www.volcengine.com/docs/6444/1340578)                      | -                            |
| VOLC_SK                         | Volcengine 图像生成模型的 SK [文档](https://www.volcengine.com/docs/6444/1340578)                      | -                            |
| DB_TYPE                         | 数据库类型：sqlite3 / mysql                                                                            | sqlite3                      |
| DB_CONF                         | 数据库配置：sqlite3 文件路径或 MySQL 连接信息                                                          | ./data/telegram_bot.db       |
| ALLOWED_TELEGRAM_USER_IDS       | 允许使用机器人的 Telegram 用户 ID，多个用户用逗号分隔，空表示所有用户可使用                           | -                            |
| DEEPSEEK_PROXY                   | DeepSeek 代理                                                                                            | -                            |
| TELEGRAM_PROXY                   | Telegram 代理                                                                                            | -                            |

### CUSTOM_URL

如果使用自建的 DeepSeek 服务，可通过 CUSTOM_URL 指定自建服务的地址。

### DEEPSEEK_TYPE

- `deepseek`: 使用官方 DeepSeek 服务，但稳定性可能波动。  
- `others`: 使用其他模型，例如 [火山引擎](https://www.volcengine.com/docs/82379/1463946)。

### DB_TYPE

支持 sqlite3 和 MySQL：

- sqlite3：配置为数据库文件路径，例如 `./data/telegram_bot.db`
- MySQL：配置为 MySQL 连接信息，数据库需提前创建，例如：
  ```
  root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
  ```

## 🛠️ 命令

### /clear
清除与 DeepSeek 的所有对话记录，这些记录用于帮助 DeepSeek 理解上下文。

### /retry
重试上一个问题。

### /mode
选择 DeepSeek 模式，包括 `chat`、`coder` 和 `reasoner`。  
- `chat` 和 `coder` 对应 DeepSeek-V3  
- `reasoner` 对应 DeepSeek-R1  
<img width="374" alt="mode" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e" />

### /balance
查询 DeepSeek 账户的余额。  
<img width="374" alt="balance" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### /state
计算当前用户的 Token 使用量。  
<img width="374" alt="state" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132" />

### /photo
使用 Volcengine 图像生成模型创建图片，需配置 VOLC_AK 和 VOLC_SK。[文档](https://www.volcengine.com/docs/6444/1340578)  
<img width="374" alt="photo" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b" />

### /chat
在群组中使用 `/chat` 命令与机器人对话，无需将机器人设置为管理员。  
<img width="374" alt="chat" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### /help
查看帮助信息。  
<img width="374" alt="help" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />

## 🚀 部署

### 使用 Docker 部署

1. **构建 Docker 镜像**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```

2. **运行容器**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="你的_telegram_token" -e DEEPSEEK_TOKEN="你的_deepseek_token" --name my-telegram-bot telegram-deepseek-bot 
   ```

## 🤝 贡献

欢迎提交 Issues 和 Pull Requests 来改进此机器人！🚀

## 📜 许可证

MIT License © 2025 jack yin
