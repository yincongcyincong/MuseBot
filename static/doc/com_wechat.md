Sure! Here’s the English version of the WeChat Work (Enterprise WeChat) bot introduction, styled similarly to the DingTalk example you shared:

---

# ✨ Enterprise WeChat Bot

This project is a cross-platform chatbot supporting **Enterprise WeChat** in both private and group chats.
The bot provides intelligent conversational features for automatic replies, information push, and collaboration scenarios.

---

## 🚀 Starting the Enterprise WeChat Bot

You can launch the bot with the following command and parameters:

```bash
./MuseBot-darwin-amd64 \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
  -com_wechat_token=xxx \
  -com_wechat_encoding_aes_key=xxx \
  -com_wechat_corp_id=xxx \
  -com_wechat_secret=xxx \
  -com_wechat_agent_id=xxx
```

---

## 💬 How to Use

### Create an Enterprise WeChat Bot

1. Log in to the Enterprise WeChat admin console:https://work.weixin.qq.com/wework_admin/   


2. Create a new application to obtain the `AgentId`, `Secret`, and configure the `Token` and `EncodingAESKey`


3. Set the callback URL to receive events and messages from Enterprise WeChat

### 1. Private Chat with the Bot

Send messages directly to the bot in Enterprise WeChat private chat. The bot will automatically respond based on message content.

---

## Supported Commands Example

* `/photo` — Generate an image    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/b3634f64-43ef-4884-9212-b8fadba5a474" />

* `/video` — Generate a video    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/ba39278d-83e8-4db0-9a72-b4f0d8c0785b" />

* `/balance` — Check your DeepSeek Token balance    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/fd2ec827-a876-4bad-8e68-a762385b064f" />

* `/state` — View the current conversation state (including model info and system prompts)
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/adfc9015-6d2b-4663-80fa-34491a6f9a8a" />

* `/clear` — Clear the current conversation context
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/47b74d3e-1425-402e-a117-882a8003bbe9" />


* `/help` — Display help information
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/2971ff23-9b68-4dbc-ad06-ad92b3c12bc8" />

* `/mode` — Choose LLM mode    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/c8020a59-2619-4160-beac-5fd628b62e4c" />
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/5327afba-9160-44e8-a0e5-f05bae6cbfd6" />


  
