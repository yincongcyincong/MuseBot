# ✨ QQ DeepSeek Bot

This project is a cross-platform chatbot powered by **LLM**, supporting **QQ**.
It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

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

Other usage see this [doc](https://github.com/yincongcyincong/MuseBot)

---

## 💬 How to Use

### Private Chat with the Bot

You can directly chat with the bot via QQ private message. <img width="400" alt="image" src="https://github.com/user-attachments/assets/6d8ded05-8454-4946-9025-bdd4bb7f8dbb" />

Supported commands:

* `/photo`: Generate an image.



* `/video`: Generate a video.


* `/balance`: Check the remaining quota of your DeepSeek Token.

* `/state`: View the current chat state (including model info and system prompts).

* `/clear`: Clear the current conversation context.

---

### Group Chat Mode

In a QQ group chat, you can talk to the bot by mentioning it using `@YourBotName`, or directly using the command. For example: <img width="400" alt="image" src="https://github.com/user-attachments/assets/c93196d9-8506-474b-8b09-1930b8bb42f1" />

All the above commands are also available in group chats without needing to switch to private chat.


