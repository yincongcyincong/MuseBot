# DeepSeek Telegram Bot

本仓库提供了一个使用 **Golang** 构建的 **Telegram 机器人**，集成了 **DeepSeek API**，能够提供 AI 驱动的智能回复。该机器人支持 **流式响应**，使交互更自然和动态。

## 🚀 功能特点
- 🤖 **AI 回复**：使用 DeepSeek API 进行智能对话。
- ⏳ **流式输出**：实时发送回复，提升用户体验。
- 🎯 **命令处理**：支持自定义命令。
- 🏗 **简单部署**：可在本地运行或部署到云服务器。

## 🤖 使用示例
[使用演示视频](https://github.com/yincongcyincong/telegram-deepseek-bot/wiki/Usage-Video)

## 📌 运行要求
- [Go 1.24+](https://go.dev/dl/)
- [Telegram 机器人令牌](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek 授权令牌](https://api-docs.deepseek.com/zh-cn/)

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

## 🚀 运行
本地运行机器人：
   ```sh
   go run main.go
   或
   go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
   ```
使用 Docker 运行：
   ```sh
   docker pull jackyin0822/telegram-deepseek-bot:latest
   docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ 配置
可以通过环境变量配置机器人：

| 变量名称                  | 描述                                                         |
|---------------------------|------------------------------------------------------------|
| TELEGRAM_BOT_TOKEN (必填) | Telegram 机器人令牌                                          |
| DEEPSEEK_TOKEN (必填)     | DeepSeek API 密钥                                           |
| CUSTOM_URL                | 自定义 DeepSeek 地址                                        |
| DEEPSEEK_TYPE             | deepseek/其他（deepseek-r1-250120，doubao-1.5-pro-32k-250115，...） |
| VOLC_AK                   | Volcengine 图片模型 AK                                      |
| VOLC_SK                   | Volcengine 图片模型 SK                                      |

### CUSTOM_URL
如果你使用的是自建的 DeepSeek，可以设置 CUSTOM_URL 来将请求路由到你的自建 DeepSeek。

### DEEPSEEK_TYPE
- **deepseek**：直接使用 DeepSeek 服务（但可能不太稳定）。
- **其他类型**：详见 [文档](https://www.volcengine.com/docs/82379/1463946)。

## 📜 机器人命令

### /mode
选择 DeepSeek 模式，包括 chat、coder、reasoner。

### /balance
查询 DeepSeek 账户余额。

### /clear
清除所有与你的 DeepSeek 交互记录，这些记录用于帮助 DeepSeek 理解上下文。

### /retry
重试上一个问题。

### /photo
使用 Volcengine 图片模型生成图片，需要配置 **VOLC_AK** 和 **VOLC_SK**。详见 [文档](https://www.volcengine.com/docs/6444/1340578)。

### /help
显示帮助信息。

## 🚀 部署
### 使用 Docker 部署
1. **构建 Docker 镜像**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```
2. **运行 Docker 容器**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot telegram-deepseek-bot
   ```

## 🤝 贡献
欢迎提交 issue 和 pull request 来改进这个机器人！🚀

## 📜 许可证
MIT 许可证 © 2025 jack yin

