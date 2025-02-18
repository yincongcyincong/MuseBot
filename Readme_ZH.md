# DeepSeek Telegram Bot

本项目是一个基于 **Golang** 构建的 **Telegram 机器人**，集成了 **DeepSeek API**，提供 AI 驱动的智能回复。支持 **流式输出**，使交互更加自然流畅。

## 🚀 功能特点
- 🤖 **AI 智能回复**：利用 DeepSeek API 提供聊天机器人服务。
- ⏳ **流式响应**：实时发送回复，提升用户体验。
- 🎯 **命令处理**：支持自定义命令。
- 🏗 **简易部署**：可本地运行，也可部署到云服务器。

## 🤖 使用示例
[使用演示视频](https://github.com/yincongcyincong/telegram-deepseek-bot/wiki/Usage-Video)

## 📌 运行环境
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

## 🚀 运行方式
本地运行：
   ```sh
   go run main.go
   ```
或使用命令行参数：
   ```sh
   go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
   ```

使用 Docker 运行：
   ```sh
   docker pull jackyin0822/telegram-deepseek-bot:latest
   docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ 配置
可以通过环境变量配置机器人：

| 变量名称            | 描述                          |
|--------------------|-----------------------------|
| TELEGRAM_BOT_TOKEN | 你的 Telegram 机器人 Token  |
| DEEPSEEK_TOKEN    | DeepSeek API 认证 Token     |
| MODE              | 运行模式（sample / complex） |
| CUSTOM_URL        | 自定义 DeepSeek API 地址    |

### 运行模式（MODE）
- **sample**：使用 DeepSeek 默认配置。
- **complex**：允许自定义 DeepSeek 配置，目前支持选择 DeepSeek 模式（chat、coder、reasoner）。

<img width="374" alt="DeepSeek 模式" src="https://github.com/user-attachments/assets/2d1bc0be-d4a2-4908-bede-b351f2a10423" />

## 🚀 部署
### 使用 Docker 部署
1. **构建 Docker 镜像**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```

2. **运行容器**
   ```sh
   docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot deepseek-telegram-bot
   ```

## 💡 贡献
欢迎提交 issue 和 pull request 以改进本项目！🚀

## 📜 许可证
MIT License © 2025 Jack Yin
