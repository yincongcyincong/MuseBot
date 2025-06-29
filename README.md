## group

telegram-group: https://t.me/+WtaMcDpaMOlhZTE1 , or you can have a try robot `GWSbot`.
every body have **3000** token to try this bot, please give me a star!

# DeepSeek Telegram Bot

This repository provides a **Telegram bot** built with **Golang** that integrates with **DeepSeek API** to provide
AI-powered responses. The bot supports **streaming replies**, making interactions feel more natural and dynamic.
[中文文档](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/Readme_ZH.md)

## 🚀 Features

- 🤖 **AI Responses**: Uses DeepSeek API for chatbot replies.
- ⏳ **Streaming Output**: Sends responses in real-time to improve user experience.
- 🏗 **Easy Deployment**: Run locally or deploy to a cloud server.
- 👀 **Identify Image**: use image to communicate with deepseek,
  see [doc](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/imageconf.md).
- 🎺 **Support Voice**: use voice to communicate with deepseek,
  see [doc](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/audioconf.md).
- 🐂 **Function Call**: transform mcp protocol to function call,
  see [doc](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/functioncall.md).
- 🌊 **RAG**: Support Rag to fill context,
  see [doc](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/rag.md).
- ⛰️**OpenRouter**: Support OpenRouter more than 400 llms, see [doc](https://openrouter.ai/docs/quickstart).

## 🤖 Text Example

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f6b5cdc7-836f-410f-a784-f7074a672c0e" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/621861a4-88d1-4796-bf35-e64698ab1b7b" />

## 🎺 Multimodal Example

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b4057dce-9ea9-4fcc-b7fa-bcc297482542" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/67ec67e0-37a4-4998-bee0-b50463b87125" />

## 📌 Requirements

- [Go 1.24+](https://go.dev/dl/)
- [Telegram Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek Auth Token](https://api-docs.deepseek.com/zh-cn/)

## 📥 Installation

1. **Clone the repository**
   ```sh
   git clone https://github.com/yincongcyincong/telegram-deepseek-bot.git
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

| Variable Name                  | 	Description                                                                                                                   | Default Value             |
|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| TELEGRAM_BOT_TOKEN (required)	 | Your Telegram bot token                                                                                                        | -                         |
| DEEPSEEK_TOKEN	  (required)    | DeepSeek Api Key                                                                                                               | -                         |
| OPENAI_TOKEN	                  | Open AI Token                                                                                                                  | -                         |
| GEMINI_TOKEN	                  | Gemini Token                                                                                                                   | -                         |
| OPEN_ROUTER_TOKEN	             | OpenRouter Token  [doc](https://openrouter.ai/docs/quickstart)                                                                 | -                         |
| VOL_TOKEN	                     | Vol Token  [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                                       | -                         |
| CUSTOM_URL	                    | custom deepseek url                                                                                                            | https://api.deepseek.com/ |
| TYPE	                          | deepseek/openai/gemini/openrouter/vol                                                                                          | deepseek                  |
| VOLC_AK	                       | volcengine photo model ak     [doc](https://www.volcengine.com/docs/6444/1340578)                                              | -                         |
| VOLC_SK	                       | volcengine photo model sk      [doc](https://www.volcengine.com/docs/6444/1340578)                                             | -                         |
| Ernie_AK	                      | ernie ak     [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                                     | -                         |
| Ernie_SK	                      | ernie sk      [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                                    | -                         |
| DB_TYPE                        | sqlite3 / mysql                                                                                                                | sqlite3                   |
| DB_CONF	                       | ./data/telegram_bot.db / root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local                        | ./data/telegram_bot.db    |
| ALLOWED_TELEGRAM_USER_IDS	     | telegram user id, only these users can use bot, using "," splite. empty means all use can use it. 0 means all user is banned   | -                         |
| ALLOWED_TELEGRAM_GROUP_IDS	    | telegram chat id, only these chat can use bot, using "," splite. empty means all group can use it. 0 means all group is banned | -                         |
| DEEPSEEK_PROXY	                | deepseek proxy                                                                                                                 | -                         |
| TELEGRAM_PROXY	                | telegram proxy                                                                                                                 | -                         |
| LANG	                          | en / zh                                                                                                                        | en                        |
| TOKEN_PER_USER	                | The tokens that each user can use                                                                                              | 10000                     |
| ADMIN_USER_IDS	                | admin user, can use some admin commands                                                                                        | -                         |
| NEED_AT_BOT	                   | is it necessary to trigger an at robot in the group                                                                            | false                     |
| MAX_USER_CHAT	                 | max existing chat per user                                                                                                     | 2                         |
| VIDEO_TOKEN	                   | volcengine Api key[doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                                | -                         |
| HTTP_PORT	                     | http server port                                                                                                               | 36060                     |
| USE_TOOLS	                     | if normal conversation  use function call tools or not                                                                         | false                     |

### CUSTOM_URL

If you are using a self-deployed DeepSeek, you can set CUSTOM_URL to route requests to your self-deployed DeepSeek.

### DEEPSEEK_TYPE

deepseek: directly use deepseek service. but it's not very stable
others: see [doc](https://www.volcengine.com/docs/82379/1463946)

### DB_TYPE

support sqlite3 or mysql

### DB_CONF

if DB_TYPE is sqlite3, give a file path, such as `./data/telegram_bot.db`
if DB_TYPE is mysql, give a mysql link, such as
`root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`, database must be created.

### LANG

choose a language for bot, English (`en`), Chinese (`zh`), Russian (`ru`).

### other config

[deepseek_conf](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/deepseekconf.md)
[photo_conf](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/photoconf.md)
[video_conf](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/videoconf.md)
[audio_conf](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/audioconf.md)

## Command

### /clear

clear all of your communication record with deepseek. this record use for helping deepseek to understand the context.

### /retry

retry last question.

### /mode

chose deepseek mode, include chat, coder, reasoner
chat and coder means DeepSeek-V3, reasoner means DeepSeek-R1.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e" />

### /balance

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### /state

calculate one user token usage.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132" />

### /photo

using volcengine photo model create photo, deepseek don't support to create photo now. VOLC_AK and VOLC_SK is
necessary.[doc](https://www.volcengine.com/docs/6444/1340578)
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b" />

### /video

create video. `DEEPSEEK_TOKEN` must be volcengine Api key. deepseek don't support to create video
now. [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16" />

### /chat

allows the bot to chat through /chat command in groups, without the bot being set as admin of the group.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### /help

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />

### /task

multi agent communicate with each other!

## Admin Command

### /addtoken

add token for user.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/12d98272-0718-4c9b-bc5c-e0a92e6c8664" />

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
