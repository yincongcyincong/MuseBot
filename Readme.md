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
- [DeepSeek API Key](https://www.deepseek.com/)

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
    export DEEPSEEK_API_KEY="your_deepseek_api_key"
    ```

## 🚀 Usage
Run the bot locally:
   ```sh
    go run main.go
   ```
## ⚙️ Configuration
You can configure the bot via environment variables:

/ Variable Name /	Description
TELEGRAM_BOT_TOKEN	Your Telegram bot token
DEEPSEEK_API_KEY	DeepSeek API key

## Deployment
### Deploy with Docker
1. **Build the Docker image**
   ```sh
    docker build -t deepseek-telegram-bot .
   ```
   
2. **Run the container**
   ```sh
    docker run -e TELEGRAM_BOT_TOKEN="your_telegram_bot_token" -e DEEPSEEK_API_KEY="your_deepseek_api_key" deepseek-telegram-bot
   ```

## Contributing
Feel free to submit issues and pull requests to improve this bot. 🚀

## License
MIT License © 2025 Your Name
