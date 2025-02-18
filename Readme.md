# DeepSeek Telegram Bot

This repository provides a **Telegram bot** built with **Golang** that integrates with **DeepSeek API** to provide AI-powered responses. The bot supports **streaming replies**, making interactions feel more natural and dynamic.

## 🚀 Features
- 🤖 **AI Responses**: Uses DeepSeek API for chatbot replies.
- ⏳ **Streaming Output**: Sends responses in real-time to improve user experience.
- 🎯 **Command Handling**: Supports custom commands.
- 🏗 **Easy Deployment**: Run locally or deploy to a cloud server.

## 📌 Requirements
- [Go 1.24+](https://go.dev/dl/)
- [Telegram Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek Auth Token](https://api-docs.deepseek.com/zh-cn/)

## 📥 Installation
1. **Clone the repository**
   ```sh
   git clone https://github.com/yourusername/deepseek-telegram-bot.git
   cd deepseek-telegram-bot
    ```
2. **Install dependencies**
   ```sh
    go mod tidy
    ```

3. **Set up environment variables**
   ```sh
    export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
    export DEEPSEEK_TOKEN="your_deepseek_api_key"
    ```

## 🚀 Usage
Run the bot locally:
   ```sh
    go run main.go
    or
    go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
   ```
Use docker
   ```sh
     docker pull jackyin0822/telegram-deepseek-bot:latest
     docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot  jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ Configuration
You can configure the bot via environment variables:

| Variable Name       | 	Description            |
|---------------------|-------------------------|
| TELEGRAM_BOT_TOKEN	 | Your Telegram bot token |
| DEEPSEEK_TOKEN	     | DeepSeek Api Key        |
| MODE	               | sample / complex        |
| CUSTOM_URL	         | deepseek url            |

### MODE
sample: all deepseek config is default config.
complex: custome yourself deepseek config. now support choose deepseek mode (chat, coder, reasoner).

## Deployment
### Deploy with Docker
1. **Build the Docker image**
   ```sh
    docker build -t deepseek-telegram-bot .
   ```
   
2. **Run the container**
   ```sh
     docker run -d -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot telegram-deepseek-bot 
   ```

## Contributing
Feel free to submit issues and pull requests to improve this bot. 🚀

## License
MIT License © 2025 Your Name
