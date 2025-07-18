# ✨ Telegram & Discord DeepSeek Bot

本项目是一个支持 **DeepSeek 大模型** 的跨平台聊天机器人，支持 **Telegram** 和 **Discord**，同时集成了多种指令功能，如生成图片、视频、查看余额、清除对话状态等。

## 🚀 启动 Discord 模式

你可以使用以下命令启动 Discord 模式：

```bash
./telegram-deepseek-bot-darwin-amd64 \
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

其他参数请使用首页[readme](https://github.com/yincongcyincong/telegram-deepseek-bot)

## 💬 使用方式

### 私聊机器人

直接私聊你的机器人，即可对话。

支持以下命令：

* `/photo`：生成图片。⚠️ 需要配置 `volc_ak` 和 `volc_sk`
* `/video`：生成视频。⚠️ 需要配置 `volc_ak` 和 `volc_sk`
* `/balance`：查看当前 DeepSeek Token 剩余额度
* `/state`：查看当前会话状态（包括模型、角色设定等）
* `/clear`：清除当前聊天上下文（重置对话）

### 群聊模式

在群聊中，通过 `@你的机器人` 的方式与机器人对话，或者直接使用命令，例如：

```
@MyDeepSeekBot 帮我写一篇英文邮件
```

群聊中也可以使用上述所有命令，无需切换至私聊。