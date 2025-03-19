# DeepSeek Telegram Bot

This repository provides a **Telegram bot** built with **Golang** that integrates with **DeepSeek API** to provide AI-powered responses. The bot supports **streaming replies**, making interactions feel more natural and dynamic.      
[中文文档](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/Readme_ZH.md)

## 🚀 Features
- 🤖 **AI Responses**: Uses DeepSeek API for chatbot replies.
- ⏳ **Streaming Output**: Sends responses in real-time to improve user experience.
- 🎯 **Command Handling**: Supports custom commands.
- 🏗 **Easy Deployment**: Run locally or deploy to a cloud server.

## 🤖 Usage Example
[usage video](https://github.com/yincongcyincong/telegram-deepseek-bot/wiki/Usage-Video)


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
    go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
   ```
Use docker
   ```sh
     docker pull jackyin0822/telegram-deepseek-bot:latest
     docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot  jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ Configuration
You can configure the bot via environment variables:

| Variable Name                  | 	Description                                                                                            |
|--------------------------------|---------------------------------------------------------------------------------------------------------|
| TELEGRAM_BOT_TOKEN (required)	 | Your Telegram bot token                                                                                 |
| DEEPSEEK_TOKEN	  (required)    | DeepSeek Api Key                                                                                        |
| CUSTOM_URL	                    | custom deepseek url                                                                                     |
| DEEPSEEK_TYPE	                 | deepseek/others(deepseek-r1-250120，doubao-1.5-pro-32k-250115，...)                                       |
| VOLC_AK	                       | volcengine photo model ak                                                                               |
| VOLC_SK	                       | volcengine photo model sk                                                                               |
| DB_TYPE                        | sqlite3 / mysql                                                                                         |
| DB_CONF	                       | ./data/telegram_bot.db / root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local |
| ALLOWED_TELEGRAM_USER_IDS	  | telegram user id, only these users can use bot, using "," splite. empty means all use can use it.       |

### CUSTOM_URL
If you are using a self-deployed DeepSeek, you can set CUSTOM_URL to route requests to your self-deployed DeepSeek.

### DEEPSEEK_TYPE
deepseek: directly use deepseek service. but it's not very stable    
others: see [doc](https://www.volcengine.com/docs/82379/1463946)    

### DB_TYPE
support sqlite3 or mysql

### DB_CONF
if DB_TYPE is sqlite3, give a file path, such as `/data/telegram_bot.db`     
if DB_TYPE is mysql, give a mysql link, such as `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`, database must be created.      

## Command 

### /mode
chose deepseek mode, include chat, coder, reasoner      
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e" />

### /balance
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### /clear
clear all of your communication record with deepseek. this record use for helping deepseek to understand the context.

### /retry
retry last question.

### /photo
using volcengine photo model create photo，VOLC_AK and VOLC_SK is necessary.[doc](https://www.volcengine.com/docs/6444/1340578)      
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b" />

### /chat
allows the bot to chat through /chat command in groups, without the bot being set as admin of the group.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### /help
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />


## Deployment
### Deploy with Docker
1. **Build the Docker image**
   ```sh
    docker build -t deepseek-telegram-bot .
   ```
   
2. **Run the container**
   ```sh
     docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot telegram-deepseek-bot 
   ```

## Contributing
Feel free to submit issues and pull requests to improve this bot. 🚀

## License
MIT License © 2025 jack yin
