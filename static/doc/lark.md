
# ✨ Lark (Feishu) DeepSeek Bot

This project is a cross-platform chatbot powered by the **DeepSeek LLM**, supporting **Telegram**, **Slack**, **Discord**, and **Lark (Feishu)**.
It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

## 🚀 Starting in Lark Mode

You can launch the bot in **Lark** mode using the following command:

```bash
./MuseBot-darwin-amd64 \
  -lark_app_id=xx \
  -lark_app_secret=xx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### Parameter Descriptions:

* `lark_app_id`: Your Lark (Feishu) App ID (required)
* `lark_app_secret`: Your Lark (Feishu) App Secret (required)
* `deepseek_token`: Your DeepSeek API Token (required)

Other usage see this [doc](https://github.com/yincongcyincong/MuseBot)

---

## 💬 How to Use

### Private Chat with the Bot

You can directly chat with the bot in Lark via **Private Chat**. <img width="400" alt="image" src="https://github.com/user-attachments/assets/6d8ded05-8454-4946-9025-bdd4bb7f8dbb" />

Supported commands:

* `/photo`: Generate an image.


* `/video`: Generate a video.


* `/balance`: Check the remaining quota of your DeepSeek Token


* `/state`: View the current chat state (including model info and system prompts)


* `/clear`: Clear the current conversation context


---

### Group Chat Mode

In a **Lark group chat**, you can talk to the bot by mentioning it using `@YourBotName`, or directly using the slash commands. For example: 


All the above commands are also available in group chats without needing to switch to private chat.

---
