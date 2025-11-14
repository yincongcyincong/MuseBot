# 感谢

感谢阮一峰老师的weekly，很荣幸能登榜：https://github.com/ruanyf/weekly      
感谢linux.do 社区，佬友们很给力，给一个我的介绍贴链接：https://linux.do/t/topic/1128110        
感谢reddit社区，虽然封了我几个subreddit：https://www.reddit.com/

# MuseBot

本仓库提供了一个是基于 **Golang** 构建的 **智能机器人**，集成了 **LLM API**，实现 AI 驱动的自然对话与智能回复。
它支持 **OpenAI**、**DeepSeek**、**Gemini**、**Doubao**、**Qwen** 等多种大模型，    
并可无缝接入 **Telegram**、**Discord**、**Slack**、**Lark（飞书）**、**钉钉**、**企业微信**、**QQ**、**微信**
等聊天平台，为用户带来更加流畅、多平台联通的 AI 对话体验。
[English Doc](https://github.com/yincongcyincong/MuseBot)

---

## 🌞 视频

最简单教程：https://www.bilibili.com/video/BV1f9nCzoERb/      
deepseek： https://www.bilibili.com/video/BV1CB8kzHEJi/    
gemini： https://www.bilibili.com/video/BV1D4htz4Ekv/    
chatgpt: https://www.bilibili.com/video/BV1RutxzJEGY/    
豆包：https://www.bilibili.com/video/BV1QDbEzwE43/    
怎么使用mcp: https://www.bilibili.com/video/BV1JbtJzVEJd/

## 🚀 功能特性

- 🤖 **AI 回复**：使用 大模型 API 提供聊天机器人回复。
- ⏳ **流式输出**：实时发送回复，提升用户体验。
- 🏗 **轻松部署**：可本地运行或部署到云服务器。
- 👀 **图像识别**
  ：使用图片与大模型进行交流，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/photoconf_ZH.md)。
- 🎺 **支持语音**
  ：使用语音与大模型进行交流，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf_ZH.md)。
- 🐂 **函数调用**：将
  MCP协议转换为函数调用，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall_ZH.md)。
- 🌊 **RAG（检索增强生成）**：支持
  RAG以填充上下文，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag_ZH.md)。
- 🌞 **管理平台（AdminPlatform）**
  ：使用管理平台来管理MuseBot，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin_ZH.md)。
- 🌛 **注册中心**
  ：支持服务注册，机器人实例可自动注册，详见 [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register_ZH.md)
- 🌈 **监控数据**：支持监控数据，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/metrics_ZH.md)。
  🐶 **Cron**: 定时触发LLM,
  see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/cron_ZH.md).

---

## 支持平台

| 平台                 | 支持情况 | 简要说明                                                            | 文档 / 链接                                                                                |
|--------------------|:----:|-----------------------------------------------------------------|----------------------------------------------------------------------------------------|
| 🟦 **Telegram**    |  ✅   | 支持 Telegram 机器人（基于 go-telegram-bot-api，可处理命令、内联按钮、ForceReply 等） | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/README_ZH.md)                |
| 🌈 **Discord**     |  ✅   | 支持 Discord 机器人                                                  | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/discord_ZH.md)    |
| 🌛 **Web API**     |  ✅   | 提供 HTTP/Web API 与 LLM 交互（适合构建自己的前端或后端集成）                        | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api_ZH.md)    |
| 🔷 **Slack**       |  ✅   | 支持 Slack（Socket Mode / Events API / Block Kit 交互）               | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/slack_ZH.md)      |
| 🟣 **Lark（飞书）**    |  ✅   | 支持 Lark 长连接与消息处理（基于 larksuite SDK，支持图片/音频下载与消息更新）               | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/lark_ZH.md)       |
| 🆙 **钉钉**          |  ✅   | 支持钉钉长链接服务                                                       | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/dingding_ZH.md)   |
| ⚡️ **Work WeChat** |  ✅   | 支持企业微信触发大模型                                                     | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/com_wechat_ZH.md) |
| ⚡️ **QQ**          |  ✅   | 支持QQ机器人触发大模型                                                    | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq_ZH.md)         |
| ⚡️ **WeChat**      |  ✅   | 支持微信触发大模型                                                       | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat_ZH.md)     |

## 支持的大型语言模型

| 模型                  | 提供方          | 文本生成 | 图片生成 | 视频生成 | 识别照片 | 返回语音 | 链接                                                                                                           |
|---------------------|--------------|------|:----:|:----:|:----:|-----:|--------------------------------------------------------------------------------------------------------------|
| 🌟 **Gemini**       | 谷歌           | ✅    |  ✅   |  ✅   |  ✅   |    ✅ | [文档](https://gemini.google.com/app)                                                                          |
| 💬 **ChatGPT**      | OpenAI       | ✅    |  ✅   |  ❌   |  ✅   |    ✅ | [文档](https://chat.openai.com)                                                                                |
| 🐦 **Doubao**       | 字节跳动         | ✅    |  ✅   |  ✅   |  ✅   |    ✅ | [文档](https://www.volcengine.com/)                                                                            |
| 🐦 **Qwen**         | 阿里云          | ✅    |  ✅   |  ✅   |  ✅   |    ✅ | [文档](https://bailian.console.aliyun.com/?spm=5176.12818093_47.overview_recent.1.663b2cc9wXXcVC&tab=api#/api) |
| ⚙️ **302.AI**       | 302.AI       | ✅    |  ✅   |  ✅   |  ✅   |    ❌ | [文档](https://302.ai/)                                                                                        |
| 🧠 **DeepSeek**     | DeepSeek     | ✅    |  ❌   |  ❌   |  ❌   |    ❌ | [文档](https://www.deepseek.com/)                                                                              |
| 🌐 **OpenRouter**   | OpenRouter   | ✅    |  ✅   |  ❌   |  ✅   |    ❌ | [文档](https://openrouter.ai/)                                                                                 |
| 🌐 **ChatAnywhere** | ChatAnywhere | ✅    |  ✅   |  ❌   |  ✅   |    ❌ | [文档](https://api.chatanywhere.tech/#/)                                                                       |

## 🤖 文本示例

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f6b5cdc7-836f-410f-a784-f7074a672c0e" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/621861a4-88d1-4796-bf35-e64698ab1b7b" />

## 🎺 多模态示例

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b4057dce-9ea9-4fcc-b7fa-bcc297482542" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/67ec67e0-37a4-4998-bee0-b50463b87125" />

---

## 📥 安装

1. **克隆仓库**
   ```sh
   git clone https://github.com/yincongcyincong/MuseBot.git
   cd MuseBot
   ```

2. **安装依赖**
   ```sh
   go mod tidy
   ```

3. **设置环境变量**
   ```sh
   export TELEGRAM_BOT_TOKEN="你的Telegram Bot Token"
   export DEEPSEEK_TOKEN="你的DeepSeek API密钥"
   ```

---

## 🚀 使用方法

在本地运行：

```sh
go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
```

使用 Docker 运行：

```sh
docker pull jackyin0822/musebot:latest
chmod 777 /home/user/data
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="你的Telegram Bot Token" -e DEEPSEEK_TOKEN="你的DeepSeek API密钥" -p 36060:36060  --name my-bot jackyin0822/musebot:latest
```

```sh
阿里云:
docker pull crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
chmod 777 /home/user/data
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" -p 36060:36060 --name my-bot  crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
  
```

命令介绍: (文档)[https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/param_conf_ZH.md]

---

## ⚙️ 配置项

如果用参数形式，使用小写加下划线形式，比如./MuseBot -telegram_bot_token=xxx

| 环境变量名字                          | 描述                                                                                  | 默认值                   |
|---------------------------------|-------------------------------------------------------------------------------------|-----------------------|
| **TELEGRAM_BOT_TOKEN**          | Telegram 机器人 Token                                                                  | -                     |
| **DISCORD_BOT_TOKEN**           | Discord 机器人 Token                                                                   | -                     |
| **SLACK_BOT_TOKEN**             | Slack 机器人 Bot Token                                                                 | -                     |
| **SLACK_APP_TOKEN**             | Slack App-level Token                                                               | -                     |
| **LARK_APP_ID**                 | 飞书 App ID                                                                           | -                     |
| **LARK_APP_SECRET**             | 飞书 App Secret                                                                       | -                     |
| **DING_CLIENT_ID**              | 钉钉 App Key / Client ID                                                              | -                     |
| **DING_CLIENT_SECRET**          | 钉钉 App Secret                                                                       | -                     |
| **DING_TEMPLATE_ID**            | 钉钉 模板消息 ID                                                                          | -                     |
| **COM_WECHAT_TOKEN**            | 企业微信 Token                                                                          | -                     |
| **COM_WECHAT_ENCODING_AES_KEY** | 企业微信 EncodingAESKey                                                                 | -                     |
| **COM_WECHAT_CORP_ID**          | 企业微信 CorpID                                                                         | -                     |
| **COM_WECHAT_SECRET**           | 企业微信 Secret                                                                         | -                     |
| **COM_WECHAT_AGENT_ID**         | 企业微信 AgentID                                                                        | -                     |
| **WECHAT_APP_ID**               | 微信公众号 AppID                                                                         | -                     |
| **WECHAT_APP_SECRET**           | 微信公众号 AppSecret                                                                     | -                     |
| **WECHAT_ENCODING_AES_KEY**     | 微信公众号 EncodingAESKey                                                                | -                     |
| **WECHAT_TOKEN**                | 微信公众号 Token                                                                         | -                     |
| **WECHAT_ACTIVE**               | 是否启用公众号消息监听（true/false）                                                             | false                 |
| **QQ_APP_ID**                   | QQ 开放平台 AppID                                                                       | -                     |
| **QQ_APP_SECRET**               | QQ 开放平台 AppSecret                                                                   | -                     |
| **QQ_ONEBOT_RECEIVE_TOKEN**     | ONEBOT → MuseBot 事件推送 token                                                         | MuseBot               |
| **QQ_ONEBOT_SEND_TOKEN**        | MuseBot → ONEBOT 消息发送 token                                                         | MuseBot               |
| **QQ_ONEBOT_HTTP_SERVER**       | ONEBOT HTTP 服务地址                                                                    | http://127.0.0.1:3000 |
| **DEEPSEEK_TOKEN**              | DeepSeek API Key                                                                    | -                     |
| **OPENAI_TOKEN**                | OpenAI API Key                                                                      | -                     |
| **GEMINI_TOKEN**                | Google Gemini Token                                                                 | -                     |
| **OPEN_ROUTER_TOKEN**           | OpenRouter Token [doc](https://openrouter.ai/docs/quickstart)                       | -                     |
| **ALIYUN_TOKEN**                | 阿里云百炼 Token [doc](https://bailian.console.aliyun.com/#/doc/?type=model&url=2840915) | -                     |
| **AI_302_TOKEN**                | 302.AI Token [doc](https://302.ai/)                                                 | -                     |
| **VOL_TOKEN**                   | 火山引擎通用 Token [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)          | -                     |
| **VOLC_AK**                     | 火山引擎多媒体 AK [doc](https://www.volcengine.com/docs/6444/1340578)                      | -                     |
| **VOLC_SK**                     | 火山引擎多媒体 SK [doc](https://www.volcengine.com/docs/6444/1340578)                      | -                     |
| **ERNIE_AK**                    | 百度文心大模型 AK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)            | -                     |
| **ERNIE_SK**                    | 百度文心大模型 SK [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)            | -                     |
| **ALIYUN_TOKEN**                | 阿里云大模型 Token                                                                        | -                     |
| **AI_302_TOKEN**                | 302.AI 平台 Token                                                                     | -                     |
| **OPEN_ROUTER_TOKEN**           | OpenRouter API Key                                                                  | -                     |
| **CUSTOM_URL**                  | 自定义 LLM API 地址                                                                      |                       |
| **TYPE**                        | LLM 类型（deepseek/openai/gemini/openrouter/vol/302-ai/chatanywhere）                   | deepseek              |
| **MEDIA_TYPE**                  | 图片/视频生成模型来源（openai/gemini/vol/openrouter/aliyun/302-ai）                             | vol                   |
| **DB_TYPE**                     | 数据库类型（sqlite3/mysql）                                                                | sqlite3               |
| **DB_CONF**                     | 数据库配置路径或连接字符串                                                                       | ./data/muse_bot.db    |
| **LLM_PROXY**                   | LLM 网络代理（如 http://127.0.0.1:7890）                                                   | -                     |
| **ROBOT_PROXY**                 | 机器人访问代理（如 http://127.0.0.1:7890）                                                    | -                     |
| **LANG**                        | 语言（en/zh）                                                                           | en                    |
| **TOKEN_PER_USER**              | 每个用户可用的最大 token 数，0为不限制token                                                        | 10000                 |
| **MAX_USER_CHAT**               | 每个用户可同时存在的最大对话数                                                                     | 2                     |
| **HTTP_HOST**                   | MuseBot HTTP 服务监听端口                                                                 | :36060                |
| **USE_TOOLS**                   | 是否启用 Function Call 工具（true/false）                                                   | false                 |
| **MAX_QA_PAIR**                 | 上下文保留问答对数量                                                                          | 100                   |
| **CHARACTER**                   | AI 的人格设定描述                                                                          | -                     |
| **CRT_FILE**                    | HTTPS 公钥文件路径                                                                        | -                     |
| **KEY_FILE**                    | HTTPS 私钥文件路径                                                                        | -                     |
| **CA_FILE**                     | HTTPS CA 证书路径                                                                       | -                     |
| **ADMIN_USER_IDS**              | 管理员用户 ID，逗号分隔                                                                       | -                     |
| **ALLOWED_USER_IDS**            | 允许使用的用户 ID，逗号分隔；空=全部可用；0=全部禁用                                                       | -                     |
| **ALLOWED_GROUP_IDS**           | 允许使用的群组 ID，逗号分隔；空=全部可用；0=全部禁用                                                       | -                     |
| **BOT_NAME**                    | Bot 名称                                                                              | MuseBot               |
| **CHAT_ANY_WHERE_TOKEN**        | ChatAnyWhere 平台 Token                                                               | -                     |
| **SMART_MODE**                  | 自动检测你想生成什么样的内容                                                                      | true                  |

### 其他配置

[deepseek参数](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/deepseekconf_ZH.md)
[图片参数](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/photoconf_ZH.md)
[视频参数](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/videoconf_ZH.md)

---

## 💬 命令

### `/clear`

清除与 DeepSeek 的历史对话记录，用于上下文清理。

### `/retry`

重试上一次问题。

### `/txt_type /photo_type /video_type /rec_type`

选择你想用的 文字/图片/视频的 模型类型.      
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b001e178-4c2a-4e4f-a679-b60be51a776b" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/ad7c3b84-b471-418b-8fe7-05af53893842" />

### `/txt_model /img_model /video_model /rec_model`

选择具体的模型名称.      
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/882f7766-c237-45e7-b0d1-9035fc65ff73" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/276af04a-d602-470e-b2c1-ba22e16225b0" />

### `/mode`

展示正在使用的模型信息
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/47fb4043-7385-4f81-b8f9-83f8352b81f9" />

### `/state`

统计用户的 Token 使用量。

<img width="400" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132"  alt=""/>

### `/photo` `/edit_photo`

<img width="400" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b"  alt=""/>        

/edit_photo 支持编辑图片。     
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b26c123a-8a61-4329-ba31-9b371bd9251c" />

### `/video`

<img width="400" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16"  alt=""/>

### `/chat`

在群组中使用 `/chat` 命令与机器人对话，无需将机器人设置为管理员。

<img width="400" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d"  alt=""/>

### `/help`

显示帮助信息。

<img width="400" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818"  alt=""/>


---

## 🚀 Docker 部署

1. **构建 Docker 镜像**
   ```sh
   docker build -t musebot .
   ```

2. **运行 Docker 容器**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="你的Telegram Bot Token" -e DEEPSEEK_TOKEN="你的DeepSeek API密钥" --name my-telegram-bot musebot
   ```

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request，一起优化和改进本项目！🚀

## 群聊

telegram群: https://t.me/+WtaMcDpaMOlhZTE1, 或者尝试一下Guanwushan_bot。
每个人有 **10000** token 去试用robot, 点个star吧!

QQ群：1031411708

---

## 📜 开源协议

MIT License © 2025 Jack Yin
