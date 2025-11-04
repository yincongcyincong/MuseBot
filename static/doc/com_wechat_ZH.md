# ✨ 企业微信机器人

本项目是一个跨平台聊天机器人，支持**企业微信**的私聊和群聊。
该机器人提供智能对话功能，支持自动回复、信息推送及协作场景。

---

## 🚀 启动企业微信机器人

你可以使用以下命令和参数启动机器人：

```bash
./MuseBot-darwin-amd64 \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx \
  -com_wechat_token=xxx \
  -com_wechat_encoding_aes_key=xxx \
  -com_wechat_corp_id=xxx \
  -com_wechat_secret=xxx \
  -com_wechat_agent_id=xxx
```

---

## 💬 使用说明

### 创建企业微信机器人

1. 登录企业微信管理后台：[https://work.weixin.qq.com/wework\_admin/](https://work.weixin.qq.com/wework_admin/)

2. 创建新的应用，获取 `AgentId`、`Secret`，并配置 `Token` 和 `EncodingAESKey`

     <img width="400" alt="image" src="https://github.com/user-attachments/assets/9ba4ea00-b4f5-441c-b6ac-1d4dd6fbccde" />

3. 设置回调URL，用于接收企业微信的事件和消息

     <img width="400" alt="image" src="https://github.com/user-attachments/assets/cd2ef979-de8c-46e4-952e-f0f30304e0aa" />

4. 配置企业可信IP

     <img width="400" alt="image" src="https://github.com/user-attachments/assets/9a1560d7-3cde-43cd-bc10-e95cb471e975" />

---

### 1. 与机器人私聊

在企业微信私聊窗口直接发送消息给机器人，机器人会根据消息内容自动回复。 <img width="400" alt="image" src="https://github.com/user-attachments/assets/e8dcce5e-bc93-4448-8a0a-0a6ed47d3348" />

---

## 支持的命令示例

* `/photo` — 生成图片 <img width="400" alt="image" src="https://github.com/user-attachments/assets/64146ff8-f296-49ee-9393-4908c818d5b8" />

* `/video` — 生成视频 <img width="400" alt="image" src="https://github.com/user-attachments/assets/f5d7185b-060e-4894-917d-91d6766e46b1" />

* `/state` — 查看当前对话状态（包括模型信息和系统提示） <img width="400" alt="image" src="https://github.com/user-attachments/assets/adfc9015-6d2b-4663-80fa-34491a6f9a8a" />

* `/clear` — 清理当前对话上下文 <img width="400" alt="image" src="https://github.com/user-attachments/assets/70b79d5e-c429-40c5-bf90-6a56fbcd99d8" />

* `/help` — 显示帮助信息 <img width="400" alt="image" src="https://github.com/user-attachments/assets/07e787f1-67ba-441a-ba03-1e80fb3d2929" />

* `/mode` — 选择 LLM 模式 <img width="400" alt="image" src="https://github.com/user-attachments/assets/4e4dfa98-16b7-46a2-a1a9-b91ac9272b4b" />


