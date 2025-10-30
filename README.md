# MuseBot

This repository provides a **Chat bot** (Telegram, Discord, Slack, Lark（飞书），钉钉, 企业微信, QQ, 微信) that integrates
with **LLM API** to provide
AI-powered responses. The bot supports **openai** **deepseek** **gemini** **openrouter** LLMs, making interactions feel
more natural and dynamic.       
[中文文档](https://github.com/yincongcyincong/MuseBot/blob/main/README_ZH.md)       
[Китайская документация](https://github.com/yincongcyincong/MuseBot/blob/main/README_RU.md)

## Usage Video

easiest way to use: https://www.youtube.com/watch?v=4UHoKRMfNZg     
deepseek: https://www.youtube.com/watch?v=kPtNdLjKVn0   
gemini: https://www.youtube.com/watch?v=7mV9RYvdE6I    
chatgpt: https://www.youtube.com/watch?v=G_DZYMvd5Ug

## 🚀 Features

- 🤖 **AI Responses**: Uses LLM API for chatbot replies.
- ⏳ **Streaming Output**: Sends responses in real-time to improve user experience.
- 🏗 **Easy Deployment**: Run locally or deploy to a cloud server.
- 👀 **Identify Image**: use image to communicate with LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/imageconf.md).
- 🎺 **Support Voice**: use voice to communicate with LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md).
- 🐂 **Function Call**: transform mcp protocol to function call,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall.md).
- 🌊 **RAG**: Support Rag to fill context,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag.md).
- 🌞 **AdminPlatform**: Use platform to manage MuseBot,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin.md).
- 🌛 **Register**: With the service registration module, robot instances can be automatically registered to the
  registration center [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register.md)
- 🌈 **Metrics**: Support Metrics for monitoring,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/metrics.md).

## 📸 Support Platform

| Platform             | Supported | Description                                                                                                           | Docs / Links                                                                           |
|----------------------|:---------:|-----------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| 🟦 **Telegram**      |     ✅     | Supports Telegram bot (go-telegram-bot-api based, handles commands, inline buttons, ForceReply, etc.)                 | [Docs](https://github.com/yincongcyincong/MuseBot)                                     |
| 🌈 **Discord**       |     ✅     | Supports Discord bot                                                                                                  | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/discord.md)     |
| 🌛 **Web API**       |     ✅     | Provides HTTP/Web API for interacting with LLM (great for custom frontends/backends)                                  | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api.md)     |
| 🔷 **Slack**         |     ✅     | Supports Slack (Socket Mode / Events API / Block Kit interactions)                                                    | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/slack.md)       |
| 🟣 **Lark (Feishu)** |     ✅     | Supports Lark long connection & message handling (based on larksuite SDK, with image/audio download & message update) | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/lark.md)        |
| 🆙 **DingDing**      |     ✅     | Supports Dingding long connection                                                                                     | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/dingding.md)    |
| ⚡️ **Work WeChat**   |     ✅     | Support Work WeChat http callback to trigger LLM                                                                      | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/com_wechat.md)  |
| 🌞 **QQ**            |     ✅     | Support QQ http callback to trigger LLM                                                                               | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq.md)          |
| 🚇 **Wechat**        |     ✅     | Support Wechat http callback to trigger LLM                                                                           | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat.md)      |

## Supported Large Language Models

| Model               | Provider     | Text Generation | Image Generation | Video Generation | Recognize Photo | TTS | Link                                                                                                          |
|---------------------|--------------|-----------------|:----------------:|:----------------:|----------------:|----:|---------------------------------------------------------------------------------------------------------------|
| 🌟 **Gemini**       | Google       | ✅               |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://gemini.google.com/app)                                                                          |
| 💬 **ChatGPT**      | OpenAI       | ✅               |        ✅         |        ❌         |               ✅ |   ✅ | [doc](https://chat.openai.com)                                                                                |
| 🐦 **Doubao**       | ByteDance    | ✅               |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://www.volcengine.com/)                                                                            |
| 🐦 **Qwen**         | Aliyun       | ✅               |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://bailian.console.aliyun.com/?spm=5176.12818093_47.overview_recent.1.663b2cc9wXXcVC&tab=api#/api) |
| 🧠 **DeepSeek**     | DeepSeek     | ✅               |        ❌         |        ❌         |               ❌ |   ❌ | [doc](https://www.deepseek.com/)                                                                              |
| ⚙️ **302.AI**       | 302.AI       | ✅               |        ✅         |        ✅         |               ✅ |   ❌ | [doc](https://302.ai/)                                                                                        |
| 🌐 **OpenRouter**   | OpenRouter   | ✅               |        ✅         |        ❌         |               ✅ |   ❌ | [doc](https://openrouter.ai/)                                                                                 |
| 🌐 **ChatAnywhere** | ChatAnywhere | ✅               |        ✅         |        ❌         |               ✅ |   ❌ | [doc](https://api.chatanywhere.tech/#/)                                                                       |

## 🤖 Text Example

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f6b5cdc7-836f-410f-a784-f7074a672c0e" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/621861a4-88d1-4796-bf35-e64698ab1b7b" />

## 🎺 Multimodal Example

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b4057dce-9ea9-4fcc-b7fa-bcc297482542" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/67ec67e0-37a4-4998-bee0-b50463b87125" />

## 📥 Installation

1. **Clone the repository**
   ```sh
   git clone https://github.com/yincongcyincong/MuseBot.git
   cd MuseBot
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
     docker pull jackyin0822/musebot:latest
     docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot  jackyin0822/MuseBot:latest
   ```

   ```sh
    ALIYUN:
    docker pull crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
   ```

command: (doc)[https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/param_conf.md]

## ⚙️ Configuration

You can configure the bot via environment variables:

Here’s the **English version** of your environment variable table:
If you use parameter. Please use lower letter and underscore. for example: ./MuseBot -telegram_bot_token=xxx

| Variable Name                     | Description                                                                                  | Default Value                                          |
|-----------------------------------|----------------------------------------------------------------------------------------------|--------------------------------------------------------|
| **TELEGRAM_BOT_TOKEN** (required) | Telegram bot token                                                                           | -                                                      |
| **DISCORD_BOT_TOKEN**             | Discord bot token                                                                            | -                                                      |
| **SLACK_BOT_TOKEN**               | Slack bot token                                                                              | -                                                      |
| **SLACK_APP_TOKEN**               | Slack app-level token                                                                        | -                                                      |
| **LARK_APP_ID**                   | Lark (Feishu) App ID                                                                         | -                                                      |
| **LARK_APP_SECRET**               | Lark (Feishu) App Secret                                                                     | -                                                      |
| **DING_CLIENT_ID**                | DingTalk App Key / Client ID                                                                 | -                                                      |
| **DING_CLIENT_SECRET**            | DingTalk App Secret                                                                          | -                                                      |
| **DING_TEMPLATE_ID**              | DingTalk template message ID                                                                 | -                                                      |
| **COM_WECHAT_TOKEN**              | WeCom (Enterprise WeChat) token                                                              | -                                                      |
| **COM_WECHAT_ENCODING_AES_KEY**   | WeCom EncodingAESKey                                                                         | -                                                      |
| **COM_WECHAT_CORP_ID**            | WeCom CorpID                                                                                 | -                                                      |
| **COM_WECHAT_SECRET**             | WeCom App Secret                                                                             | -                                                      |
| **COM_WECHAT_AGENT_ID**           | WeCom Agent ID                                                                               | -                                                      |
| **WECHAT_APP_ID**                 | WeChat Official Account AppID                                                                | -                                                      |
| **WECHAT_APP_SECRET**             | WeChat Official Account AppSecret                                                            | -                                                      |
| **WECHAT_ENCODING_AES_KEY**       | WeChat Official Account EncodingAESKey                                                       | -                                                      |
| **WECHAT_TOKEN**                  | WeChat Official Account Token                                                                | -                                                      |
| **WECHAT_ACTIVE**                 | Whether to enable WeChat message listening (true/false)                                      | false                                                  |
| **QQ_APP_ID**                     | QQ Open Platform AppID                                                                       | -                                                      |
| **QQ_APP_SECRET**                 | QQ Open Platform AppSecret                                                                   | -                                                      |
| **QQ_ONEBOT_RECEIVE_TOKEN**       | Token for ONEBOT → MuseBot event messages                                                    | MuseBot                                                |
| **QQ_ONEBOT_SEND_TOKEN**          | Token for MuseBot → ONEBOT message sending                                                   | MuseBot                                                |
| **QQ_ONEBOT_HTTP_SERVER**         | ONEBOT HTTP server address                                                                   | [http://127.0.0.1:3000](http://127.0.0.1:3000)         |
| **DEEPSEEK_TOKEN** (required)     | DeepSeek API key                                                                             | -                                                      |
| **OPENAI_TOKEN**                  | OpenAI API key                                                                               | -                                                      |
| **GEMINI_TOKEN**                  | Google Gemini API token                                                                      | -                                                      |
| **OPEN_ROUTER_TOKEN**             | OpenRouter token [doc](https://openrouter.ai/docs/quickstart)                                | -                                                      |
| **ALIYUN_TOKEN**                  | Aliyun Bailian token [doc](https://bailian.console.aliyun.com/#/doc/?type=model&url=2840915) | -                                                      |
| **AI_302_TOKEN**                  | 302.AI token [doc](https://302.ai/)                                                          | -                                                      |
| **VOL_TOKEN**                     | Volcano Engine general token [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)   | -                                                      |
| **VOLC_AK**                       | Volcano Engine multimedia access key [doc](https://www.volcengine.com/docs/6444/1340578)     | -                                                      |
| **VOLC_SK**                       | Volcano Engine multimedia secret key [doc](https://www.volcengine.com/docs/6444/1340578)     | -                                                      |
| **ERNIE_AK**                      | Baidu ERNIE large model AK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)     | -                                                      |
| **ERNIE_SK**                      | Baidu ERNIE large model SK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)     | -                                                      |
| **CUSTOM_URL**                    | Custom DeepSeek API endpoint                                                                 | [https://api.deepseek.com/](https://api.deepseek.com/) |
| **TYPE**                          | LLM type (deepseek/openai/gemini/openrouter/vol/302-ai/chatanywhere)                         | deepseek                                               |
| **MEDIA_TYPE**                    | Media generation source (openai/gemini/vol/openrouter/aliyun/302-ai)                         | vol                                                    |
| **DB_TYPE**                       | Database type (sqlite3/mysql)                                                                | sqlite3                                                |
| **DB_CONF**                       | Database config path or connection string                                                    | ./data/muse_bot.db                                     |
| **LLM_PROXY**                     | LLM network proxy (e.g. [http://127.0.0.1:7890](http://127.0.0.1:7890))                      | -                                                      |
| **ROBOT_PROXY**                   | Bot network proxy (e.g. [http://127.0.0.1:7890](http://127.0.0.1:7890))                      | -                                                      |
| **LANG**                          | Language (en/zh)                                                                             | en                                                     |
| **TOKEN_PER_USER**                | Max tokens allowed per user                                                                  | 10000                                                  |
| **MAX_USER_CHAT**                 | Maximum concurrent chats per user                                                            | 2                                                      |
| **HTTP_HOST**                     | MuseBot HTTP server port                                                                     | :36060                                                 |
| **USE_TOOLS**                     | Enable function-calling tools (true/false)                                                   | false                                                  |
| **MAX_QA_PAIR**                   | Max number of question-answer pairs to keep as context                                       | 100                                                    |
| **CHARACTER**                     | AI personality description                                                                   | -                                                      |
| **CRT_FILE**                      | HTTPS certificate file path                                                                  | -                                                      |
| **KEY_FILE**                      | HTTPS private key file path                                                                  | -                                                      |
| **CA_FILE**                       | HTTPS CA certificate file path                                                               | -                                                      |
| **ADMIN_USER_IDS**                | Comma-separated list of admin user IDs                                                       | -                                                      |
| **ALLOWED_USER_IDS**              | Comma-separated user IDs allowed to use the bot; empty = all allowed; 0 = all banned         | -                                                      |
| **ALLOWED_GROUP_IDS**             | Comma-separated group IDs allowed to use the bot; empty = all allowed; 0 = all banned        | -                                                      |
| **BOT_NAME**                      | Bot name                                                                                     | MuseBot                                                |
| **CHAT_ANY_WHERE_TOKEN**          | ChatAnyWhere platform token                                                                  | -                                                      |

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

[deepseek_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/deepseekconf.md)        
[photo_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/photoconf.md)      
[video_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/videoconf.md)      
[audio_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md)

## Command

### /clear $clear

clear all of your communication record with deepseek. this record use for helping deepseek to understand the context.

### /retry $retry

retry last question.

### /txt_type /photo_type /video_type /rec_type $txt_type $photo_type $video_type $rec_type
choose txt/photo/video/recognize model type.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b001e178-4c2a-4e4f-a679-b60be51a776b" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/ad7c3b84-b471-418b-8fe7-05af53893842" />

### /txt_model /img_model /video_model /rec_model $txt_model $img_model $video_model $rec_model
choose txt/photo/video/recognize model.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/882f7766-c237-45e7-b0d1-9035fc65ff73" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/276af04a-d602-470e-b2c1-ba22e16225b0" />


### /mode $mode
show current model type and model.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/47fb4043-7385-4f81-b8f9-83f8352b81f9" />


### /balance $balance

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### /state $state

calculate one user token usage.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132" />

### /photo /edit_photo $photo $edit_photo
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b05fcadc-800e-40fb-b9a1-8aea44851550" />

/edit_photo will update you photo base on your description.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b26c123a-8a61-4329-ba31-9b371bd9251c" />

### /video $video
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16" />

### /chat $chat

allows the bot to chat through /chat command in groups,
without the bot being set as admin of the group.        
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### /help $help

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />

### /task $task

multi agent communicate with each other!

### /change_photo $change_photo

only for tencent app (wechat, qq, work wechat)          
change photo base on your prompt.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/81e1eb85-ddb6-4a2b-b6bd-73da0d276036" />

### /rec_photo $rec_photo

only for tencent app (wechat, qq, work wechat)    
recognize photo base on your prompt.    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b67a2be2-cc5e-4985-90f3-d72c7a9bf4c1" />

### /save_voice $save_voice

only for tencent app (wechat, qq, work wechat)
save your voice to pc.
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/74499d85-4405-43d3-836e-2977de08cb31" />

## Deployment

### Deploy with Docker

1. **Build the Docker image**
   ```sh
    docker build -t MuseBot .
   ```

2. **Run the container**
   ```sh
     docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-bot MuseBot
   ```

## Contributing

Feel free to submit issues and pull requests to improve this bot. 🚀

## group

telegram-group: https://t.me/+WtaMcDpaMOlhZTE1 , or you can have a try robot `Guanwushan_bot`.
every body have **10000** token to try this bot, please give me a star!

QQ群：1031411708

## License

MIT License © 2025 jack yin
