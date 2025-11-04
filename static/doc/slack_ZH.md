# ✨ Slack Bot

本项目是一个由 **DeepSeek LLM** 驱动的跨平台聊天机器人，支持 **Slack**。
它内置了多种命令，包括图片生成、视频生成、额度查询、清空对话等功能。

## 🚀 在 Slack 模式下启动

你可以使用以下命令以 **Slack 模式** 启动机器人：

```bash
./MuseBot-darwin-amd64 \
  -slack_bot_token=xoxb-xxx \
  -slack_app_token=xapp-xxx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### 参数说明

* `slack_bot_token`：你的 Slack Bot 用户 OAuth Token（必填，格式：`xoxb-xxx`）
* `slack_app_token`：你的 Slack 应用级 Token（必填，格式：`xapp-xxx`）
* `deepseek_token`：你的 DeepSeek API Token（必填）

更多用法请参考 [文档](https://github.com/yincongcyincong/MuseBot)

---

## 💬 使用方法

### 创建机器人

访问：[https://api.slack.com/apps/](https://api.slack.com/apps/) <img width="400" alt="image" src="https://github.com/user-attachments/assets/15b97d1f-7a50-4c9e-953b-7899a1ecd935" />

配置 OAuth： <img width="400" alt="image" src="https://github.com/user-attachments/assets/e15ef32e-e4c4-4560-b076-a44e27a8c65e" /> <img width="400" alt="image" src="https://github.com/user-attachments/assets/cfc51f3b-2e57-4575-ae2d-b6d9568921e7" />

设置命令： <img width="400" alt="image" src="https://github.com/user-attachments/assets/3bd4ba71-c383-42f2-8b6b-47ca4d1c1f32" />

订阅事件： <img width="400" alt="image" src="https://github.com/user-attachments/assets/8f67c815-d755-4688-983d-647bc64122f1" />

---

### 与机器人私聊

你可以在 Slack 中通过 **私聊** 直接与机器人对话。 <img width="400" alt="image" src="https://github.com/user-attachments/assets/08ee0e28-e08c-47f6-825e-45afe2621bba" />

支持的命令：

* `/photo`：生成图片

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/078d1899-57eb-4a00-a240-d44cb7dd1a51" />

* `/video`：生成视频

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/6db607d1-3a73-4a50-a0b9-81899e19f4f6" />

* `/state`：查看当前会话状态（包括模型信息和系统提示词）

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/9c336a10-9250-41a3-9406-4e385fe8d9db" />

* `/clear`：清空当前会话上下文

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/11defe94-5642-490a-bc20-d22e5e430f81" />

* `/help`：显示命令帮助信息

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/1e7da122-7199-4792-94a2-0835d647b9b5" />

* `/mode`：显示模型信息

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/225cc31b-8461-4a9a-a036-abd55b151924" />

