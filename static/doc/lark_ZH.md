# ✨ Lark（飞书 Bot

本项目是一个由 **DeepSeek LLM** 驱动的跨平台聊天机器人，支持 **Telegram**、**Slack**、**Discord** 和 **Lark（飞书）**。
它内置了多种命令，包括图片生成、视频生成、额度查询、清空对话等功能。

## 🚀 在飞书模式下启动

你可以使用以下命令以 **飞书模式** 启动机器人：

```bash
./MuseBot-darwin-amd64 \
  -lark_app_id=xx \
  -lark_app_secret=xx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### 参数说明

* `lark_app_id`：你的飞书 App ID（必填）
* `lark_app_secret`：你的飞书 App Secret（必填）
* `deepseek_token`：你的 DeepSeek API Token（必填）

更多用法请参考 [文档](https://github.com/yincongcyincong/MuseBot)

---

## 💬 使用方法

### 创建机器人

访问飞书开放平台：[https://open.feishu.cn/app/](https://open.feishu.cn/app/) <img width="400" alt="image" src="https://github.com/user-attachments/assets/4c96862e-3d90-48ad-a491-6d459ebebcc2" />

配置权限： <img width="400" alt="image" src="https://github.com/user-attachments/assets/27f6747c-bd44-4ad2-ae4c-c600078d93e5" /> <img width="400" alt="image" src="https://github.com/user-attachments/assets/bded8047-1994-4018-b885-4f68dae3eb99" />

选择订阅方式和事件： <img width="400" alt="image" src="https://github.com/user-attachments/assets/302d5aa8-863c-4a6c-92fc-2f9348b0e147" />

---

### 与机器人私聊

你可以在飞书中通过 **私聊** 直接与机器人对话。 <img width="400" alt="image" src="https://github.com/user-attachments/assets/462f1b06-8d75-427c-afe0-0f77cc85bb2f" />

支持的命令：

* `/photo`：生成一张图片

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/b32a54e9-fb17-4baf-a284-42d44156e776" />

* `/video`：生成一个视频

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/2d903781-2ad8-4e1a-9b34-dd99dc398688" />

* `/balance`：查询 DeepSeek Token 剩余额度

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/b0c193ba-9005-4fec-893e-c78c3a77947a" />

* `/state`：查看当前对话状态（包括模型信息和系统提示词）

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/50cb16f8-f94e-459b-85a0-c25dec10afaa" />

* `/clear`：清空当前会话上下文

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/deb65625-3e51-4581-a6d1-736de4ad7c5e" />
