# ✨ QQ Bot

This project is a cross-platform chatbot powered by **LLM**, supporting **QQ**.
It comes with a variety of built-in commands, including image and video generation, conversation clearing, and more.

## 🚀 Starting in QQ Mode

You can launch the bot in QQ mode using the following command:

```bash
./MuseBot-darwin-amd64 \
  -qq_app_id=xxx \
  -qq_app_secret=xxx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### Parameter Descriptions:

* `qq_bot_token`: Your QQ Bot APP ID (required)
* `qq_app_secret`: Your QQ Bot APP Secret (required)


---

## 💬 How to Use

### set QQ
go to website: https://q.qq.com/qqbot/

sandbox config:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/ab2ab481-3a41-41f7-b279-0873175ec6c0" />

callback config:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/dd88981f-eca8-4728-a021-c5ebdc0767ca" />

### Private Chat with the Bot
<img width="400" alt="image" src="https://github.com/user-attachments/assets/44394437-ed93-4e89-bb15-a0bbe55ea0e6" />

You can directly chat with the bot via QQ private message. 

Supported commands:

* `/photo`: Generate an image.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/0e15c7cc-cd24-4418-821a-2675d0e2ed9a" />


* `/video`: Generate a video.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/8e895453-6e2d-49b0-a3f8-a625404d136e" />

* `/state`: View the current chat state (including model info and system prompts).    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/c6bf87b0-706f-40dc-9aa1-20790af94923" />
  
* `/clear`: Clear the current conversation context.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/d49ff765-6a62-4a8c-aefd-77fe5bc834e7" />


---

### Group Chat Mode

In a QQ group chat, you can talk to the bot by mentioning it using `@YourBotName`, or directly using the command.

All the above commands are also available in group chats without needing to switch to private chat.


