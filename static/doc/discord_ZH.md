# ✨ Telegram & Discord DeepSeek Bot

本项目是一个支持 **DeepSeek 大模型** 的跨平台聊天机器人，支持 **Telegram** 和 **Discord**，同时集成了多种指令功能，如生成图片、视频、查看余额、清除对话状态等。

## 🚀 启动 Discord 模式

你可以使用以下命令启动 Discord 模式：

```bash
./MuseBot-darwin-amd64 \
-discord_bot_token=xxx \
-deepseek_token=sk-xxx \
-volc_ak=xxx \
-volc_sk=xxx \
-vol_token=xxx
```

参数说明：

* `discord_bot_token`：你的 Discord Bot Token（必填）
* `deepseek_token`：你的 DeepSeek 模型访问 Token（必填）
* `volc_ak` / `volc_sk`：用于生成图片和视频的火山引擎 Access Key 和 Secret Key（使用 `/photo` 和 `/video` 命令时必填）
* `vol_token`：火山引擎视频功能使用的 Token

其他参数请使用首页[readme](https://github.com/yincongcyincong/MuseBot)

## 💬 使用方式

### 私聊机器人

直接私聊你的机器人，即可对话。    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/6d8ded05-8454-4946-9025-bdd4bb7f8dbb" />

支持以下命令：

* `/photo`：生成图片。⚠️ 需要配置 `volc_ak` 和 `volc_sk`
<img width="400" alt="image" src="https://github.com/user-attachments/assets/ba0eb926-7924-4c58-bc61-7cff522bd71c" />
<img width="400" alt="image" src="https://github.com/user-attachments/assets/bef94980-4498-4eba-b4b5-bd5531816009" />

* `/video`：生成视频。⚠️ 需要配置 `volc_ak` 和 `volc_sk`
<img width="400" alt="image" src="https://github.com/user-attachments/assets/24bdde29-685c-4af7-8834-873dbc14b84f" />
<img width="400" alt="image" src="https://github.com/user-attachments/assets/b9e85a58-58fe-4e45-ab44-52b73bcaea59" />
  
* `/balance`：查看当前 DeepSeek Token 剩余额度
<img width="400" alt="image" src="https://github.com/user-attachments/assets/bb20e8fd-470f-4c70-b584-abc1fb5855d2" />
  
* `/state`：查看当前会话状态（包括模型、角色设定等）
<img width="400" alt="image" src="https://github.com/user-attachments/assets/bf57f0fa-add1-4cb2-8e82-7bd484a880b8" />
  
* `/clear`：清除当前聊天上下文（重置对话）
<img width="400" alt="image" src="https://github.com/user-attachments/assets/ebba556f-267a-4052-a3a3-3eab019eb4f4" />

### 群聊模式

在群聊中，通过 `@你的机器人` 的方式与机器人对话，或者直接使用命令，例如：
<img width="400" alt="image" src="https://github.com/user-attachments/assets/c93196d9-8506-474b-8b09-1930b8bb42f1" />


群聊中也可以使用上述所有命令，无需切换至私聊。
