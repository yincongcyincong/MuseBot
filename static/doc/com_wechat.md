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
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/9ba4ea00-b4f5-441c-b6ac-1d4dd6fbccde" />
   
3. Set the callback URL to receive events and messages from Enterprise WeChat    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/cd2ef979-de8c-46e4-952e-f0f30304e0aa" />

4. Enterprise Trusted IP is necessary
 <img width="400" alt="image" src="https://github.com/user-attachments/assets/9a1560d7-3cde-43cd-bc10-e95cb471e975" />  


### 1. Private Chat with the Bot

Send messages directly to the bot in Enterprise WeChat private chat. The bot will automatically respond based on message content.
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/e8dcce5e-bc93-4448-8a0a-0a6ed47d3348" />


## Supported Commands Example

* `/photo` — Generate an image    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/64146ff8-f296-49ee-9393-4908c818d5b8" />

* `/video` — Generate a video    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/f5d7185b-060e-4894-917d-91d6766e46b1" />

  
* `/balance` — Check your DeepSeek Token balance    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/f94e420a-1259-4565-b2aa-757b758f6553" />

* `/state` — View the current conversation state (including model info and system prompts)    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/adfc9015-6d2b-4663-80fa-34491a6f9a8a" />
  
* `/clear` — Clear the current conversation context    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/70b79d5e-c429-40c5-bf90-6a56fbcd99d8" />

* `/help` — Display help information    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/07e787f1-67ba-441a-ba03-1e80fb3d2929" />

* `/mode` — Choose LLM mode    
  <img width="400" alt="image" src="https://github.com/user-attachments/assets/4e4dfa98-16b7-46a2-a1a9-b91ac9272b4b" />



  
