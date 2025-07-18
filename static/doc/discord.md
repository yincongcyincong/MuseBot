# ✨ Telegram & Discord DeepSeek Bot

This project is a cross-platform chatbot powered by the **DeepSeek LLM**, supporting both **Telegram** and **Discord**. It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

## 🚀 Starting in Discord Mode

You can launch the bot in Discord mode using the following command:

```bash
./telegram-deepseek-bot-darwin-amd64 \
  -discord_bot_token=xxx \
  -deepseek_token=sk-xxx \
  -volc_ak=xxx \
  -volc_sk=xxx \
  -vol_token=xxx
```

### Parameter Descriptions:

* `discord_bot_token`: Your Discord Bot Token (required)
* `deepseek_token`: Your DeepSeek API Token (required)
* `volc_ak` / `volc_sk`: Volcano Engine Access Key and Secret Key (required for `/photo` and `/video` commands)
* `vol_token`: Token for using Volcano Engine video capabilities

other usage see this [doc](https://github.com/yincongcyincong/telegram-deepseek-bot))

## 💬 How to Use

### Private Chat with the Bot

You can directly chat with the bot via private message.

Supported commands:

* `/photo`: Generate an image. ⚠️ Requires `volc_ak` and `volc_sk`
* `/video`: Generate a video. ⚠️ Requires `volc_ak` and `volc_sk`
* `/balance`: Check the remaining quota of your DeepSeek Token
* `/state`: View the current chat state (including model info and system prompts)
* `/clear`: Clear the current conversation context

### Group Chat Mode

In a group chat, you can talk to the bot by mentioning it using `@YourBotName`, or directly using the command. For example:

```
@MyDeepSeekBot Help me write an English email.
```

All the above commands are also available in group chats without needing to switch to private chat.

---

Let me know if you'd like this formatted as a full `README.md` file or want a bilingual version!
