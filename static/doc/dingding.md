---

# ✨ DingTalk Bot

This project is a cross-platform chatbot that supports **DingTalk** in both private and group chats.
It is powered by the **LLM**, with built-in commands for image and video generation, balance checking, conversation clearing, and more.

---

## 🚀 Starting in DingTalk Mode

You can launch the bot in DingTalk mode with the following command:

```bash
./MuseBot-darwin-amd64 \
  -ding_client_id=dingdzbvgn8i0pnznhyk \
  -ding_client_secret=-YXnsH4ETINs6tD0_lPlRt1bvilWaHECBkWPRnB548F924_1Ij2givm-WS4C5_Ye \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

---

## 💬 How to Use

### 1. Private Chat with the Bot

Send commands directly to the bot in a DingTalk private chat.

Supported commands:

* `/photo` — Generate an image
* `/video` — Generate a video
* `/balance` — Check your DeepSeek Token balance
* `/state` — View the current conversation state (including model info and system prompts)
* `/clear` — Clear the current conversation context
* `/help` — Display help information
* `/mode` — Choose LLM mode

