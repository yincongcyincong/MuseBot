# ✨ WeChat Bot

本项目是一个基于 **LLM** 的跨平台聊天机器人，支持 **微信（WeChat）**。
它内置了多种命令，包括图片与视频生成、余额查询、对话清理等功能。

---

## 🚀 在微信模式下启动

你可以通过以下命令启动 **微信模式**：

```bash
./MuseBot \
  -wechat_app_secret=xxx \
  -wechat_app_id=xxx \
  -wechat_active=true \
  -wechat_token=xx \
  -gemini_token=xxxxxx \
  -type=gemini \
  -media_type=gemini
```

### 参数说明：

* `wechat_app_secret`：微信公众号的 **AppSecret**（必填）

* `wechat_app_id`：微信公众号的 **AppID**（必填）

* `wechat_token`：微信公众号的 **Token**（必填）

* `wechat_active`：是否支持 **主动发送消息**（`true/false`）

    * `true`：支持主动推送消息（受微信每日额度限制）
    * `false`：仅支持被动回复模式（微信要求在 15 秒内响应，否则消息将被截断）

* `gemini_token`：你的 **Gemini API Token**（必填）

* `type` / `media_type`：模型类型，这里设置为 `gemini`

⚠️ 建议：使用 **测试号（sandbox account）**，该模式下支持无限次主动发送消息。

更多用法详见 [文档](https://github.com/yincongcyincong/MuseBot)

---

## 💬 使用方法

### 创建微信公众号应用

1. 登录 [微信公众平台](https://mp.weixin.qq.com/)，配置 **域名、Token、EncodingAESKey** <img width="400" alt="image" src="https://github.com/user-attachments/assets/ee252dfd-3a93-41d6-b7af-dcaba530f4fd" />

---

### 与机器人对话

连接成功后，你可以通过 **微信公众号** 直接与机器人对话。

支持的指令：

* **普通对话**：输入文本即可获得 AI 回复。

* `/photo`：生成图片 <img width="400" alt="image" src="https://github.com/user-attachments/assets/1d3ee270-98f1-437d-900f-8dba6b8c9bf0" />

* `/video`：生成视频 <img width="400" alt="image" src="https://github.com/user-attachments/assets/8332c9f0-08aa-4f72-a037-6c94c4a97f60" />

* `/balance`：查看 DeepSeek Token 的剩余额度 <img width="400" alt="image" src="https://github.com/user-attachments/assets/1f51e14c-346c-4a6d-a57d-ae06a1b7f4a5" />

* `/state`：查看当前会话token消耗状态 <img width="400" alt="image" src="https://github.com/user-attachments/assets/e7e2260e-d279-4660-962a-99dbc0e7d1f9" />

* `/clear`：清理当前会话上下文 <img width="400" alt="image" src="https://github.com/user-attachments/assets/6c53c15c-7f2a-41ea-8e53-103c1e8c1e24" />

* `/help`：查看命令帮助信息 <img width="400" alt="image" src="https://github.com/user-attachments/assets/d8cfe98c-b424-4e65-8a29-e95320d51e49" />

* `/mode`：显示模型信息 <img width="400" alt="image" src="https://github.com/user-attachments/assets/85477f22-2592-41d0-971b-a41e1d80e54a" />


