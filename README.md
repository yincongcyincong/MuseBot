<div align="center" id="MuseBot">

<a href="https://github.com/yincongcyincong/MuseBot" title="TrendRadar">
  <img src="/static/logo.png" alt="MuseBot Banner" width="50%">
</a>

🚀 Connect your communication app to AI in just one minute

[![GitHub Stars](https://img.shields.io/github/stars/yincongcyincong/MuseBot?style=flat-square&logo=github&color=yellow)](https://github.com/yincongcyincong/MuseBot/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/yincongcyincong/MuseBot?style=flat-square&logo=github&color=blue)](https://github.com/yincongcyincong/MuseBot/network/members)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/yincongcyincong/MuseBot/blob/main/LICENSE)
[![Docker](https://img.shields.io/badge/Docker-deploy-2496ED?style=flat-square&logo=docker&logoColor=white)](https://hub.docker.com/search?q=musebot)
[![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-auto-2088FF?style=flat-square&logo=github-actions&logoColor=white)](https://github.com/yincongcyincong/MuseBot/actions)

</div>

# MuseBot

This repository provides a **Chat bot** (Telegram, Discord, Slack, Lark（飞书），钉钉, 企业微信, QQ, 微信) that integrates
with **LLM API** to provide
AI-powered responses. The bot supports **openai** **deepseek** **gemini** LLMs, making interactions feel
more natural and dynamic.       
[中文文档](https://github.com/yincongcyincong/MuseBot/blob/main/README_ZH.md)       
[Китайская документация](https://github.com/yincongcyincong/MuseBot/blob/main/README_RU.md)

# Thanks

感谢阮一峰老师的weekly，很荣幸能登榜：https://github.com/ruanyf/weekly      
感谢linux.do 社区，佬友们很给力，给一个我的介绍贴链接：https://linux.do/t/topic/1128110        
Thanks to the Reddit community as well, even though a few of my subreddits got banned 😅: https://www.reddit.com/

## 🚀 Features

- 🤖 **AI Responses**: Uses LLM API for chatbot replies.
- ⏳ **Streaming Output**: Sends responses in real-time to improve user experience.
- 🏗 **Easy Deployment**: Run locally or deploy to a cloud server.
- 👀 **Identify Image**: use image to communicate with LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/imageconf.md).
- 📄 **Identify Document**: send txt/pdf/csv/html documents and ask questions; chat media is stored
  locally per conversation, see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/document.md).
- 🎺 **Support Voice**: use voice to communicate with LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md).
- 🐂 **Function Call**: transform mcp protocol to function call,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall.md).
- 🧩 **Skills**: load local `SKILL.md` instructions into the unified `/task` workflow,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/skills.md).
- 🛠️ **Conversational System Operations**: inside the `/task` workflow, manage runtime settings, MCP, skills, and cron jobs, and run commands and file operations inside safety boundaries,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/conversation_config.md).
- 🌊 **RAG**: Support Rag to fill context,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag.md).
- 🌞 **AdminPlatform**: Use platform to manage MuseBot,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin.md).
- 🌛 **Register**: With the service registration module, robot instances can be automatically registered to the
  registration center [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register.md)
- 🌈 **Metrics**: Support Metrics for monitoring,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/metrics.md).
- 🐶 **Cron**: Support Cron to trigger LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/cron.md).

## Usage Video

easiest way to use: https://www.youtube.com/watch?v=4UHoKRMfNZg     
deepseek: https://www.youtube.com/watch?v=kPtNdLjKVn0   
gemini: https://www.youtube.com/watch?v=7mV9RYvdE6I    
chatgpt: https://www.youtube.com/watch?v=G_DZYMvd5Ug

## 📸 Support Platform

| Platform             | Supported | Description                                                                                                           | Docs / Links                                                                          |
|----------------------|:---------:|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| 🟦 **Telegram**      |     ✅     | Supports Telegram bot (go-telegram-bot-api based, handles commands, inline buttons, ForceReply, etc.)                 | [Docs](https://github.com/yincongcyincong/MuseBot)                                    |
| 🌈 **Discord**       |     ✅     | Supports Discord bot                                                                                                  | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/discord.md)    |
| 🌛 **Web API**       |     ✅     | Provides HTTP/Web API for interacting with LLM (great for custom frontends/backends)                                  | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api.md)    |
| 🔷 **Slack**         |     ✅     | Supports Slack (Socket Mode / Events API / Block Kit interactions)                                                    | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/slack.md)      |
| 🟣 **Lark (Feishu)** |     ✅     | Supports Lark long connection & message handling (based on larksuite SDK, with image/audio download & message update) | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/lark.md)       |
| 🆙 **DingDing**      |     ✅     | Supports Dingding long connection                                                                                     | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/dingding.md)   |
| ⚡️ **Work WeChat**   |     ✅     | Support Work WeChat http callback to trigger LLM                                                                      | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/com_wechat.md) |
| 🌞 **QQ**            |     ✅     | Support QQ http callback to trigger LLM                                                                               | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq.md)         |
| 🚇 **Wechat**        |     ✅     | Support Wechat http callback to trigger LLM                                                                           | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat.md)     |

## Supported Large Language Models

> **NOTE**: All providers are accessed through the **OpenAI-compatible API format** (chat completions, images, audio).
> **Aliyun (Qwen)** and **Vol (Doubao)** are the recommended/fully-supported providers. Third-party relay
> services (OpenRouter / OrcaRouter / 302.AI) have been removed.

| Model               | Provider     | Text Generation | Image Generation | Video Generation | Recognize Photo | TTS | Link                                                                                                          |
|---------------------|--------------|-----------------|:----------------:|:----------------:|----------------:|----:|---------------------------------------------------------------------------------------------------------------|
| 🐦 **Doubao**       | ByteDance (vol) | ✅            |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://www.volcengine.com/)                                                                            |
| 🐦 **Qwen**         | Aliyun       | ✅               |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://bailian.console.aliyun.com/?spm=5176.12818093_47.overview_recent.1.663b2cc9wXXcVC&tab=api#/api) |
| 🌟 **Gemini**       | Google       | ✅               |        ✅         |        ✅         |               ✅ |   ✅ | [doc](https://gemini.google.com/app)                                                                          |
| 💬 **ChatGPT**      | OpenAI       | ✅               |        ✅         |        ❌         |               ✅ |   ✅ | [doc](https://chat.openai.com)                                                                                |
| 🧠 **DeepSeek**     | DeepSeek     | ✅               |        ❌         |        ❌         |               ❌ |   ❌ | [doc](https://www.deepseek.com/)                                                                              |
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
     chmod 777 /home/user/data
     docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" -p 36060:36060 -p 18080:18080 --name my-bot  jackyin0822/MuseBot:latest
   ```

   ```sh
    ALIYUN:
    docker pull crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
    chmod 777 /home/user/data
     docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" -p 36060:36060 -p 18080:18080 --name my-bot  crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
   ```

command: (doc)[https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/param_conf.md]

## ⚙️ Configuration

You can configure the bot via environment variables:

Here’s the **English version** of your environment variable table:
If you use parameter. Please use lower letter and underscore. for example: ./MuseBot -telegram_bot_token=xxx

| Variable Name                   | Description                                                                                  | Default Value                                          |
|---------------------------------|----------------------------------------------------------------------------------------------|--------------------------------------------------------|
| **TELEGRAM_BOT_TOKEN**          | Telegram bot token                                                                           | -                                                      |
| **DISCORD_BOT_TOKEN**           | Discord bot token                                                                            | -                                                      |
| **SLACK_BOT_TOKEN**             | Slack bot token                                                                              | -                                                      |
| **SLACK_APP_TOKEN**             | Slack app-level token                                                                        | -                                                      |
| **LARK_APP_ID**                 | Lark (Feishu) App ID                                                                         | -                                                      |
| **LARK_APP_SECRET**             | Lark (Feishu) App Secret                                                                     | -                                                      |
| **DING_CLIENT_ID**              | DingTalk App Key / Client ID                                                                 | -                                                      |
| **DING_CLIENT_SECRET**          | DingTalk App Secret                                                                          | -                                                      |
| **DING_TEMPLATE_ID**            | DingTalk template message ID                                                                 | -                                                      |
| **COM_WECHAT_TOKEN**            | WeCom (Enterprise WeChat) token                                                              | -                                                      |
| **COM_WECHAT_ENCODING_AES_KEY** | WeCom EncodingAESKey                                                                         | -                                                      |
| **COM_WECHAT_CORP_ID**          | WeCom CorpID                                                                                 | -                                                      |
| **COM_WECHAT_SECRET**           | WeCom App Secret                                                                             | -                                                      |
| **COM_WECHAT_AGENT_ID**         | WeCom Agent ID                                                                               | -                                                      |
| **WECHAT_APP_ID**               | WeChat Official Account AppID                                                                | -                                                      |
| **WECHAT_APP_SECRET**           | WeChat Official Account AppSecret                                                            | -                                                      |
| **WECHAT_ENCODING_AES_KEY**     | WeChat Official Account EncodingAESKey                                                       | -                                                      |
| **WECHAT_TOKEN**                | WeChat Official Account Token                                                                | -                                                      |
| **WECHAT_ACTIVE**               | Whether to enable WeChat message listening (true/false)                                      | false                                                  |
| **QQ_APP_ID**                   | QQ Open Platform AppID                                                                       | -                                                      |
| **QQ_APP_SECRET**               | QQ Open Platform AppSecret                                                                   | -                                                      |
| **QQ_ONEBOT_RECEIVE_TOKEN**     | Token for ONEBOT → MuseBot event messages                                                    | MuseBot                                                |
| **QQ_ONEBOT_SEND_TOKEN**        | Token for MuseBot → ONEBOT message sending                                                   | MuseBot                                                |
| **QQ_ONEBOT_HTTP_SERVER**       | ONEBOT HTTP server address                                                                   | [http://127.0.0.1:3000](http://127.0.0.1:3000)         |
| **DEEPSEEK_TOKEN**              | DeepSeek API key                                                                             | -                                                      |
| **OPENAI_TOKEN**                | OpenAI API key                                                                               | -                                                      |
| **GEMINI_TOKEN**                | Google Gemini API token                                                                      | -                                                      |
| **ALIYUN_TOKEN**                | Aliyun Bailian token [doc](https://bailian.console.aliyun.com/#/doc/?type=model&url=2840915) | -                                                      |
| **VOL_TOKEN**                   | Volcano Engine general token [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)   | -                                                      |
| **VOLC_AK**                     | Volcano Engine multimedia access key [doc](https://www.volcengine.com/docs/6444/1340578)     | -                                                      |
| **VOLC_SK**                     | Volcano Engine multimedia secret key [doc](https://www.volcengine.com/docs/6444/1340578)     | -                                                      |
| **ERNIE_AK**                    | Baidu ERNIE large model AK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)     | -                                                      |
| **ERNIE_SK**                    | Baidu ERNIE large model SK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)     | -                                                      |
| **CUSTOM_URL**                  | Custom LLM API endpoint                                                                      | [https://api.deepseek.com/](https://api.deepseek.com/) |
| **TYPE**                        | LLM type (deepseek/openai/gemini/vol/chatanywhere/aliyun/ollama), all via OpenAI-compatible format | deepseek                                         |
| **MEDIA_TYPE**                  | Media generation source (openai/gemini/vol/aliyun), all via OpenAI-compatible format          | vol                                                    |
| **DB_TYPE**                     | Database type (sqlite3/mysql)                                                                | sqlite3                                                |
| **DB_CONF**                     | Database config path or connection string                                                    | ./data/muse_bot.db                                     |
| **LLM_PROXY**                   | LLM network proxy (e.g. [http://127.0.0.1:7890](http://127.0.0.1:7890))                      | -                                                      |
| **ROBOT_PROXY**                 | Bot network proxy (e.g. [http://127.0.0.1:7890](http://127.0.0.1:7890))                      | -                                                      |
| **LANG**                        | Language (en/zh)                                                                             | en                                                     |
| **TOKEN_PER_USER**              | Max tokens allowed per user, 0 means no limit                                                | 10000                                                  |
| **MAX_USER_CHAT**               | Maximum concurrent chats per user                                                            | 2                                                      |
| **HTTP_HOST**                   | MuseBot HTTP server port                                                                     | :36060                                                 |
| **USE_TOOLS**                   | Enable function-calling tools (true/false)                                                   | false                                                  |
| **MAX_QA_PAIR**                 | Max number of question-answer pairs to keep as context                                       | 100                                                    |
| **CHARACTER**                   | AI personality description                                                                   | -                                                      |
| **CRT_FILE**                    | HTTPS certificate file path                                                                  | -                                                      |
| **KEY_FILE**                    | HTTPS private key file path                                                                  | -                                                      |
| **CA_FILE**                     | HTTPS CA certificate file path                                                               | -                                                      |
| **ADMIN_USER_IDS**              | Comma-separated list of admin user IDs                                                       | -                                                      |
| **ALLOWED_USER_IDS**            | Comma-separated user IDs allowed to use the bot; empty = all allowed; 0 = all banned         | -                                                      |
| **ALLOWED_GROUP_IDS**           | Comma-separated group IDs allowed to use the bot; empty = all allowed; 0 = all banned        | -                                                      |
| **BOT_NAME**                    | Bot name                                                                                     | MuseBot                                                |
| **CHAT_ANY_WHERE_TOKEN**        | ChatAnyWhere platform token                                                                  | -                                                      |
| **SMART_MODE**                  | Automatically check what you want to generate (txt/photo/video) and route multi-step tasks to `/task` | true |
| **SEND_MCP_RES**                | send mcp result to user                                                                      | false                                                  |
| **TASK_AUTO_APPROVE**           | auto approve task tool operations (commands/config changes) without /y confirmation          | false                                                  |
| **DEFAULT_MODEL**               | default txt model                                                                            | -                                                      |
| **CONFIG**                      | path to a json config file; file values beat command-line flags, env vars beat the file       | -                                                      |

### config file

Instead of long flag/env lists, put your configuration in a json file and point the bot at it with `-conf` or the `CONFIG` env var:

```sh
go run main.go -conf=conf/examples/telegram_deepseek.json

# docker: mount the config file
docker run -d -v /home/user/data:/app/data \
  -v /home/user/mybot.json:/app/mybot.json \
  -e CONFIG="/app/mybot.json" \
  -p 36060:36060 -p 18080:18080 --name my-bot jackyin0822/musebot:latest
```

Two formats are supported:

**Flat** (recommended — top-level keys use the lowercased-underscore form of the env names):

```json
{
  "telegram_bot_token": "your-telegram-bot-token",
  "type": "deepseek",
  "deepseek_token": "your-deepseek-api-key",
  "smart_mode": true,
  "admin_user_ids": {"123456789": true}
}
```

**Sectioned** (same layout as the config persisted by the [admin panel](static/doc/admin.md): `base`/`llm`/`audio`/`photo`/`video`/`rag`/`tools`/`register` sections for fine-grained control):

```json
{
  "base": {
    "telegram_bot_token": "your-telegram-bot-token",
    "type": "deepseek",
    "deepseek_token": "your-deepseek-api-key"
  },
  "llm": {
    "temperature": 0.7
  }
}
```

Precedence: **env vars > config file > command-line flags > defaults**. Keys missing from the file keep their defaults; restart to apply changes.

[conf/examples/](conf/examples/) ships a few classic configs ready to copy and edit:

| File                            | Scenario                                                        |
|---------------------------------|------------------------------------------------------------------|
| `telegram_deepseek.json`        | Telegram + DeepSeek, simplest setup                              |
| `telegram_glm_thirdparty.json`  | Telegram + Zhipu GLM (third-party OpenAI-compatible), smart_mode + task auto-approval on |
| `discord_openai.json`           | Discord + OpenAI, function call + streaming on                   |
| `lark_aliyun.json`              | Lark + Aliyun Bailian                                            |
| `dingtalk_deepseek.json`        | DingTalk + DeepSeek, smart_mode on                               |
| `full_nested.json`              | Full sectioned example (base + llm + tools)                      |

### CUSTOM_URL

If you are using a self-deployed llm, you can set CUSTOM_URL to route requests to your self-deployed llm.

### DB_TYPE

support sqlite3 or mysql

### DB_CONF

if DB_TYPE is sqlite3, give a file path, such as `./data/telegram_bot.db`
if DB_TYPE is mysql, give a mysql link, such as
`root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`, database must be created.

### LANG

choose a language for bot, English (`en`), Chinese (`zh`), Russian (`ru`).

### other config

[llm_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/deepseekconf.md)        
[photo_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/photoconf.md)      
[video_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/videoconf.md)      
[audio_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md)
[advanced_conf (third-party models / smart_mode / task auto-approval)](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/advanced.md)

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

## Deployment

### Deploy with Docker

The image runs **two processes** under supervisord: the main MuseBot service and
the MuseBotAdmin management platform. Both restart automatically if they crash.

1. **Build the Docker image**
   ```sh
    docker build -t MuseBot .
   ```

2. **Run the container**
   ```sh
     docker run -d -v /home/user/xxx/data:/app/data \
       -e TELEGRAM_BOT_TOKEN="telegram-bot-token" \
       -e DEEPSEEK_TOKEN="deepseek-auth-token" \
       -p 36060:36060 -p 18080:18080 \
       --name my-bot MuseBot
   ```

3. **Ports and volumes**
   - `/app/data`: persist this volume — sqlite database, admin database
     (`muse_bot_admin.db`), media files and the auto-generated admin session key.
   - `36060`: main service HTTP (webhook callbacks, config API, SSE chat).
   - `18080`: admin web UI.
   - `ffmpeg`/`ffprobe`, `nodejs`/`npm` and `libopus` are pre-installed for
     voice, video and skill features.

4. **First admin login**
   The default account `admin` / `admin` is created on first start. After
   login the admin UI shows a change-password prompt (you can click "Later"
   to skip, but changing it soon is recommended).
   Log in at `http://<host>:18080` (see the
   [admin doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin.md)).

## Task web search

The built-in `web_search` task agent uses DuckDuckGo with Bing fallback by default.
To use [Parallel Search MCP](https://docs.parallel.ai/integrations/mcp/search-mcp)
instead, set `WEB_SEARCH_PROVIDER=parallel` before starting MuseBot, or pass
`-web_search_provider=parallel` on a fresh installation. The setting is also saved
as `tools.web_search_provider` in the bot configuration. Environment values
override the saved setting.

Parallel searches use Fast mode at `https://search.parallel.ai/mcp`, without an
account or API key. The task agent receives titles, source URLs and relevant
excerpts in the existing search result format. This provider supports search;
it does not add a page-fetch tool. Anonymous usage is subject to service limits,
and errors are returned to the task without switching providers.

Use `default` to retain DuckDuckGo/Bing, or `disabled` to block built-in web
search. Custom MCP servers and their tool settings are separate. Each `/task`
or `/mcp` research run has its own opaque Parallel session identifier, reused by
its agents and subsequent searches within that run.

## Contributing

Feel free to submit issues and pull requests to improve this bot. 🚀

## group

telegram-group: https://t.me/+WtaMcDpaMOlhZTE1 , or you can have a try robot `Guanwushan_bot`.
every body have **10000** token to try this bot, please give me a star!

QQ群：1031411708

## License

MIT License © 2025 jack yin
