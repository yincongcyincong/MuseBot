# DeepSeek Telegram Bot

本仓库提供了一个基于 **Golang** 构建的 **Telegram 机器人**，集成了 **DeepSeek API**，可以提供 AI 驱动的智能回复。该机器人支持 **流式回复**，使交互更加自然和动态。

## 🚀 功能特点
- 🤖 **AI 智能回复**：使用 DeepSeek API 进行聊天机器人回复。
- ⏳ **流式输出**：实时发送回复，提升用户体验。
- 🎯 **命令处理**：支持自定义命令。
- 🏗 **易于部署**：可本地运行或部署到云服务器。

## 🤖 使用示例
[使用演示视频](https://github.com/yincongcyincong/telegram-deepseek-bot/wiki/Usage-Video)

## 📌 运行环境要求
- [Go 1.24+](https://go.dev/dl/)
- [Telegram 机器人 Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek 认证 Token](https://api-docs.deepseek.com/zh-cn/)

## 📥 安装步骤
1. **克隆仓库**
   ```sh
   git clone https://github.com/yourusername/deepseek-telegram-bot.git
   cd deepseek-telegram-bot
   ```
2. **安装依赖**
   ```sh
   go mod tidy
   ```
3. **设置环境变量**
   ```sh
   export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
   export DEEPSEEK_TOKEN="your_deepseek_api_key"
   ```

## 🚀 启动机器人
本地运行：
   ```sh
   go run main.go
   或
   go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
   ```
使用 Docker 运行：
   ```sh
   docker pull jackyin0822/telegram-deepseek-bot:latest
   docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot  jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ 配置
机器人支持通过环境变量进行配置：

| 变量名称             | 描述                                                                 |
|----------------------|--------------------------------------------------------------------|
| TELEGRAM_BOT_TOKEN  | 你的 Telegram 机器人 Token                                       |
| DEEPSEEK_TOKEN      | DeepSeek API Key                                                  |
| CUSTOM_URL          | 自定义 DeepSeek API 地址                                          |
| DEEPSEEK_TYPE       | deepseek/others(deepseek-r1-250120，doubao-1.5-pro-32k-250115，...) |

### CUSTOM_URL
如果你使用的是自部署的 DeepSeek，可以通过 CUSTOM_URL 设置请求地址。

### DEEPSEEK_TYPE
- **deepseek**: 直接使用官方 DeepSeek 服务，但可能不太稳定。
- **others**: 其他类型，详情参考 [文档](https://www.volcengine.com/docs/82379/1463946)。

## 机器人命令

### /mode
选择 DeepSeek 模式，包括 chat、coder、reasoner。

### /balance
查询 DeepSeek 账户余额。

### /clear
清除你与 DeepSeek 的全部聊天记录（用于帮助 DeepSeek 理解上下文）。

### /retry
重新尝试上一个问题。

### /help
查看帮助信息。

## 🛠 部署
### 使用 Docker 部署
1. **构建 Docker 镜像**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```

2. **运行容器**
   ```sh
   docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot telegram-deepseek-bot
   ```

## 🤝 贡献
欢迎提交 Issue 和 Pull Request 来优化这个机器人！🚀

## 📜 许可证
MIT License © 2025 Jack Yin

