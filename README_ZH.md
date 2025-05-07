## group
telegram群: https://t.me/+WtaMcDpaMOlhZTE1, 或者尝试一下GWSbot。
每个人有 **3000** token 去试用robot, 点个star吧!

# DeepSeek Telegram Bot

本仓库提供了一个基于 **Golang** 构建的 **Telegram 机器人**，集成了 **DeepSeek API**，实现 AI 驱动的回复。  
该机器人支持 **流式输出**，让对话体验更加自然和流畅。  
[English Doc](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/Readme.md)

---

## 🚀 功能特性

- 🤖 **AI 回复**：通过 DeepSeek API 实现智能聊天。
- ⏳ **流式输出**：流式返回回复，减少等待时间。
- 🎯 **自定义命令**：支持自定义命令扩展功能。
- 🏗 **简易部署**：支持本地或云服务器部署。
- 👀 **识别图像**: 支持图像输入与 DeepSeek 沟通, 详见 [图像文档](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/imageconf_ZH.md).
- 🎺 **支持语音**: 支持语音输入与 DeepSeek 沟通, 详见 [语音文档](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/audioconf_ZH.md).
- 🐂 **函数调用**: 支持将 mcp 协议转为函数调用, 详见 [函数调用文档](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/functioncall_ZH.md).


---

## 🤖 文本示例
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f6b5cdc7-836f-410f-a784-f7074a672c0e" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/621861a4-88d1-4796-bf35-e64698ab1b7b" />

## 🎺 多模态示例
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b4057dce-9ea9-4fcc-b7fa-bcc297482542" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/67ec67e0-37a4-4998-bee0-b50463b87125" />


## 📌 环境要求

- [Go 1.24+](https://go.dev/dl/)
- [Telegram Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [DeepSeek Auth Token](https://api-docs.deepseek.com/zh-cn/)

---

## 📥 安装

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
docker pull jackyin0822/telegram-deepseek-bot:latest
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="你的Telegram Bot Token" -e DEEPSEEK_TOKEN="你的DeepSeek API密钥" --name my-telegram-bot jackyin0822/telegram-deepseek-bot:latest
```

---

## ⚙️ 配置项

机器人支持通过环境变量进行配置：

| 变量名称                      | 描述                                                                                  | 默认值                       |
|------------------------------|-------------------------------------------------------------------------------------|-----------------------------|
| TELEGRAM_BOT_TOKEN（必填）   | Telegram 机器人 Token                                                                  | -                           |
| DEEPSEEK_TOKEN（必填）       | DeepSeek API 密钥或火山引擎密钥 [文档](https://www.volcengine.com/docs/82379/1399008#b00dee71) | -                           |
| CUSTOM_URL                    | 自定义 DeepSeek 请求地址（自建服务）                                                             | https://api.deepseek.com/   |
| DEEPSEEK_TYPE                  | deepseek / 其他模型（如 deepseek-r1-250120, doubao-1.5-pro-32k-250115）                    | deepseek                    |
| VOLC_AK                       | 火山引擎图片模型 AK [文档](https://www.volcengine.com/docs/6444/1340578)                      | -                           |
| VOLC_SK                       | 火山引擎图片模型 SK [文档](https://www.volcengine.com/docs/6444/1340578)                      | -                           |
| DB_TYPE                       | 数据库类型：sqlite3 或 mysql                                                               | sqlite3                     |
| DB_CONF                       | 数据库配置：sqlite3 文件路径或 MySQL 连接串                                                       | ./data/telegram_bot.db      |
| ALLOWED_TELEGRAM_USER_IDS     | 指定允许使用机器人的用户 ID，逗号分隔，空为所有用户可用，0 表示禁止所有用户                                            | -                           |
| ALLOWED_TELEGRAM_GROUP_IDS    | 指定允许使用机器人的群组 ID，逗号分隔，空为所有群可用，0 表示禁止所有群                                              | -                           |
| DEEPSEEK_PROXY                 | DeepSeek 请求代理地址                                                                     | -                           |
| TELEGRAM_PROXY                 | Telegram 请求代理地址                                                                     | -                           |
| LANG	                          | en / zh                                                                             | en                        |
| TOKEN_PER_USER	                | 每个用户可用token数量                                                                       | 10000                     |
| ADMIN_USER_IDS	                | 管理员，逗号分隔，可使用管理员命令                                                                   | -                         |
| NEED_AT_BOT	                   | 在群组中是否需要at机器人触发                                                                     | false                     |
| MAX_USER_CHAT	                 | 每个用户的最大现有聊天次数                                                             | 2                         |
| VIDEO_TOKEN	                   | 火山引擎 Api key[文档](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                                | -                         |


### 其他配置
[deepseek参数](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/deepseekconf_ZH.md)
[图片参数](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/photoconf_ZH.md)    
[视频参数](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/videoconf_ZH.md)

---

## 💬 命令

### `/clear`

清除与 DeepSeek 的历史对话记录，用于上下文清理。

### `/retry`

重试上一次问题。

### `/mode`

选择 DeepSeek 模式，包括：
- `chat`: 对话模式（DeepSeek-V3）
- `coder`: 编程模式（DeepSeek-V3）
- `reasoner`: 推理模式（DeepSeek-R1）

<img width="400" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e" />

### `/balance`

查询当前用户的 DeepSeek API 余额。

<img width="400" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### `/state`

统计用户的 Token 使用量。

<img width="400" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132" />

### `/photo`

使用火山引擎图片模型生成图片，DeepSeek 暂不支持图片生成。  
需要配置 `VOLC_AK` 和 `VOLC_SK`。[文档](https://www.volcengine.com/docs/6444/1340578)

<img width="400" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b" />

### `/video`

生成视频，需要使用火山引擎 API 密钥（`DEEPSEEK_TOKEN`），DeepSeek 暂不支持视频生成。  
[文档](https://www.volcengine.com/docs/82379/1399008#b00dee71)

<img width="400" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16" />

### `/chat`

在群组中使用 `/chat` 命令与机器人对话，无需将机器人设置为管理员。

<img width="400" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### `/help`

显示帮助信息。

<img width="400" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />


## 管理员命令
### /addtoken
给用户增加token.  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/12d98272-0718-4c9b-bc5c-e0a92e6c8664" />

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
