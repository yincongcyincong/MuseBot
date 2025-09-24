## group

telegram群: https://t.me/+WtaMcDpaMOlhZTE1, 或者尝试一下Guanwushan_bot。
每个人有 **10000** token 去试用robot, 点个star吧!

QQ群：1031411708

# MuseBot

本仓库提供了一个基于 **Golang** 构建的 **机器人**，集成了 **LLM API**，实现 AI 驱动的回复。
该机器人支持 **openai** **deepseek** **gemini** **Doubao** **openrouter**等大模型，让对话体验更加自然和流畅。
[English Doc](https://github.com/yincongcyincong/MuseBot/blob/main/Readme.md)

---

## 🌞 视频

deepseek： https://www.bilibili.com/video/BV1CB8kzHEJi/    
gemini： https://www.bilibili.com/video/BV1D4htz4Ekv/    
chatgpt: https://www.bilibili.com/video/BV1RutxzJEGY/    
豆包：https://www.bilibili.com/video/BV1QDbEzwE43/    
怎么使用mcp: https://www.bilibili.com/video/BV1JbtJzVEJd/

## 🚀 功能特性

- 🤖 **AI 回复**：使用 大模型 API 提供聊天机器人回复。
- ⏳ **流式输出**：实时发送回复，提升用户体验。
- 🏗 **轻松部署**：可本地运行或部署到云服务器。
- 👀 **图像识别**：使用图片与 大模型
  进行交流，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/imageconf.md)。
- 🎺 **支持语音**：使用语音与 大模型
  进行交流，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md)。
- 🐂 **函数调用**：将 MCP
  协议转换为函数调用，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall.md)。
- 🌊 **RAG（检索增强生成）**：支持 RAG
  以填充上下文，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag.md)。
- 🌞 **管理平台（AdminPlatform）**：使用管理平台来管理
  MuseBot，详见[文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin_ZH.md)。
- 🌛 **注册中心**：支持服务注册，机器人实例可自动注册，详见 [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register_ZH.md)
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
| ⚡️ **QQ**          |  ✅   | 支持QQ触发大模型                                                       | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq_ZH.md)         |
| ⚡️ **WeChat**      |  ✅   | 支持微信触发大模型                                                       | [文档](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat_ZH.md)     |

## 支持的大型语言模型

| 模型                | 提供方        | 文本生成 | 图片生成 | 视频生成 | 识别照片 | 链接                                  |
|-------------------|------------|------|:----:|:----:|:----:|-------------------------------------|
| 🌟 **Gemini**     | 谷歌         | ✅    |  ✅   |  ✅   |  ✅   | [文档](https://gemini.google.com/app) |
| 💬 **ChatGPT**    | OpenAI     | ✅    |  ✅   |  ❌   |  ✅   | [文档](https://chat.openai.com)       |
| 🐦 **Doubao**     | 字节跳动       | ✅    |  ✅   |  ✅   |  ✅   | [文档](https://www.volcengine.com/)   |
| ⚙️ **302.AI**     | 302.AI     | ✅    |  ✅   |  ✅   |  ✅   | [文档](https://302.ai/)               |
| 🧠 **DeepSeek**   | DeepSeek   | ✅    |  ❌   |  ❌   |  ❌   | [文档](https://www.deepseek.com/)     |
| 🌐 **OpenRouter** | OpenRouter | ✅    |  ✅   |  ❌   |  ✅   | [文档](https://openrouter.ai/)        |

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
   git clone git@github.com:yincongcyincong/MuseBot.git
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
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="你的Telegram Bot Token" -e DEEPSEEK_TOKEN="你的DeepSeek API密钥" --name my-telegram-bot jackyin0822/MuseBot:latest
```

```sh
阿里云:
docker pull crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot
```

---

## ⚙️ 配置项

| 变量名                         | 描述                                                                                                            | 默认值                       |
|:----------------------------|:--------------------------------------------------------------------------------------------------------------|:--------------------------|
| **TELEGRAM_BOT_TOKEN** (必需) | 您的 Telegram 机器人令牌                                                                                             | -                         |
| **DEEPSEEK_TOKEN** (必需)     | DeepSeek API 密钥                                                                                               | -                         |
| **OPENAI_TOKEN**            | OpenAI 令牌                                                                                                     | -                         |
| **GEMINI_TOKEN**            | Gemini 令牌                                                                                                     | -                         |
| **OPEN_ROUTER_TOKEN**       | OpenRouter 令牌 [文档](https://openrouter.ai/docs/quickstart)                                                     | -                         |
| **VOL_TOKEN**               | 火山引擎 令牌 [文档](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                          | -                         |
| **CUSTOM_URL**              | 自定义 DeepSeek URL                                                                                              | https://api.deepseek.com/ |
| **TYPE**                    | 模型类型：deepseek/openai/gemini/openrouter/vol/302-ai                                                             | deepseek                  |
| **VOLC_AK**                 | 火山引擎图片模型 AK [文档](https://www.volcengine.com/docs/6444/1340578)                                                | -                         |
| **VOLC_SK**                 | 火山引擎图片模型 SK [文档](https://www.volcengine.com/docs/6444/1340578)                                                | -                         |
| **Ernie_AK**                | 文心一言 AK [文档](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                          | -                         |
| **Ernie_SK**                | 文心一言 SK [文档](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                          | -                         |
| **DB_TYPE**                 | 数据库类型：sqlite3 / mysql                                                                                         | sqlite3                   |
| **DB_CONF**                 | 数据库配置：./data/telegram_bot.db 或 root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local | ./data/telegram_bot.db    |
| **ALLOWED_USER_IDS**        | 允许使用机器人的 Telegram 用户 ID，多个 ID 用逗号分隔。为空表示所有用户可用。为 0 表示禁止所有用户。                                                  | -                         |
| **ALLOWED_GROUP_IDS**       | 允许使用机器人的 Telegram 群组 ID，多个 ID 用逗号分隔。为空表示所有群组可用。为 0 表示禁止所有群组。                                                  | -                         |
| **DEEPSEEK_PROXY**          | DeepSeek 代理                                                                                                   | -                         |
| **TELEGRAM_PROXY**          | Telegram 代理                                                                                                   | -                         |
| **LANG**                    | 语言：en / zh                                                                                                    | en                        |
| **TOKEN_PER_USER**          | 每个用户可使用的令牌数                                                                                                   | 10000                     |
| **ADMIN_USER_IDS**          | 管理员用户 ID，可使用一些管理命令                                                                                            | -                         |
| **NEED_AT_BOT**             | 在群组中是否需要 @机器人才能触发                                                                                             | false                     |
| **MAX_USER_CHAT**           | 每个用户最大同时存在的聊天数                                                                                                | 2                         |
| **VIDEO_TOKEN**             | 火山引擎视频模型 API 密钥 [文档](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                  | -                         |
| **HTTP_PORT**               | HTTP 服务器端口                                                                                                    | 36060                     |
| **USE_TOOLS**               | 普通对话是否使用函数调用工具                                                                                                | false                     |
| **CA_FILE**		               | http 服务的 ca文件                                                                                                 | -                         |
| **CRT_FILE**		              | http 服务的 crt 文件                                                                                               | -                         |
| **KEY_FILE**		              | http 服务的 key 文件                                                                                               | -                         |
| **MEDIA_TYPE**		            | openai/gemini/vol/openrouter/302-ai   图片或视频生成模型                                                               | vol                       |
| **MAX_QA_PAIR**	            | 用多少问题对作为上下文                                                                                                   | 15                        |
| **CHARACTER**	              | 角色背景                                                                                                          | -                         |

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

### `/mode`

<img width="400" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e"  alt=""/>

### `/balance`

查询当前用户的 DeepSeek API 余额。

<img width="400" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410"  alt=""/>

### `/state`

统计用户的 Token 使用量。

<img width="400" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132"  alt=""/>

### `/photo` `/edit_photo`

使用火山引擎图片模型生成图片，DeepSeek 暂不支持图片生成。       
需要配置 `VOLC_AK` 和 `VOLC_SK`。[文档](https://www.volcengine.com/docs/6444/1340578)

<img width="400" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b"  alt=""/>        

/edit_photo 支持编辑图片。     
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b26c123a-8a61-4329-ba31-9b371bd9251c" />

### `/video`

生成视频，需要使用火山引擎 API 密钥（`DEEPSEEK_TOKEN`），DeepSeek 暂不支持视频生成。
[文档](https://www.volcengine.com/docs/82379/1399008#b00dee71)

<img width="400" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16"  alt=""/>

### `/chat`

在群组中使用 `/chat` 命令与机器人对话，无需将机器人设置为管理员。

<img width="400" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d"  alt=""/>

### `/help`

显示帮助信息。

<img width="400" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818"  alt=""/>

### /change_photo

对腾讯系的app起作用：qq，微信 ，企业微信     
输入一段prompt用户修改图片        
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/81e1eb85-ddb6-4a2b-b6bd-73da0d276036" />

### /rec_photo

对腾讯系的app起作用：qq，微信 ，企业微信     
输入一段prompt用户识别图片        
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b67a2be2-cc5e-4985-90f3-d72c7a9bf4c1" />


### /save_voice
仅适用于腾讯应用（微信、QQ、企业微信）
将你的语音保存到电脑。
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/74499d85-4405-43d3-836e-2977de08cb31" />


---

## 🚀 Docker 部署

1. **构建 Docker 镜像**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```

2. **运行 Docker 容器**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="你的Telegram Bot Token" -e DEEPSEEK_TOKEN="你的DeepSeek API密钥" --name my-telegram-bot deepseek-telegram-bot
   ```

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request，一起优化和改进本项目！🚀

---

## 📜 开源协议

MIT License © 2025 Jack Yin
