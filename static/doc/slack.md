好的，我帮你基于 Discord 版本改成 Slack 版本，并且加上 `-slack_bot_token=xoxb-xx` 和 `-slack_app_token=xapp-xx` 参数说明。
我会尽量保持文档风格一致，只是替换成 Slack 场景。

---

# ✨ Slack DeepSeek Bot

This project is a cross-platform chatbot powered by the **DeepSeek LLM**, supporting both **Telegram** and **Slack**. It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

## 🚀 Starting in Slack Mode

You can launch the bot in Slack mode using the following command:

```bash
./MuseBot-darwin-amd64 \
  -slack_bot_token=xoxb-xxx \
  -slack_app_token=xapp-xxx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### Parameter Descriptions:

* `slack_bot_token`: Your Slack Bot User OAuth Token (required, format: `xoxb-xxx`)
* `slack_app_token`: Your Slack App-Level Token (required, format: `xapp-xxx`)
* `deepseek_token`: Your DeepSeek API Token (required)

Other usage see this [doc](https://github.com/yincongcyincong/MuseBot)

---

## 💬 How to Use

### Private Chat with the Bot

You can directly chat with the bot in Slack via **Direct Message**. <img width="400" alt="image" src="https://github.com/user-attachments/assets/6d8ded05-8454-4946-9025-bdd4bb7f8dbb" />

Supported commands:

* `/photo`: Generate an image.


* `/video`: Generate a video.

* `/balance`: Check the remaining quota of your DeepSeek Token


* `/state`: View the current chat state (including model info and system prompts)


* `/clear`: Clear the current conversation context
* 
* `/help`: Show command help info


---

### Channel Mode

In a Slack channel, you can talk to the bot by mentioning it using `@YourBotName`, or directly using the slash commands. For example: 


All the above commands are also available in channels without needing to switch to private chat.

