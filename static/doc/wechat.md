# ✨ WeChat DeepSeek Bot

This project is a cross-platform chatbot powered by the **LLM**, supporting **WeChat**.
It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

---

## 🚀 Starting in WeChat Mode

You can launch the bot in **WeChat mode** using the following command:

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

### Parameter Descriptions:

* `wechat_app_secret`: Your WeChat Official Account **AppSecret** (required)
* `wechat_app_id`: Your WeChat Official Account **AppID** (required)
* `wechat_token`: Your WeChat Official Account **Token** (required)
* `wechat_active`: Whether the bot can **actively send messages** (`true/false`)

    * `true`: Support proactive messages (limited by WeChat’s daily quota)
    * `false`: Only passive reply mode (WeChat requires response within 15s, otherwise truncated)
* `gemini_token`: Your **Gemini API Token** (required)
* `type` / `media_type`: The model type, here set to `gemini`

⚠️ Recommendation: Use a **sandbox account**, which allows unlimited proactive messages.

Other usage see this [doc](https://github.com/yincongcyincong/MuseBot)

---

## 💬 How to Use

### Create a WeChat Official Account App

1. Go to the [WeChat Official Account Platform](https://mp.weixin.qq.com/).

2. Register a public account (建议使用测试号 sandbox)。

3. Get the following values:

    * **AppID**
    * **AppSecret**
    * **Token** (自行设置，在“开发 → 基本配置 → 服务器配置”中填写)

4. Start your bot with these parameters.

---

### Chat with the Bot

Once connected, you can chat with the bot directly via **WeChat Official Account**.

Supported commands:

* **Normal chat**: Input text and get AI responses.
* `/photo`: Generate an image.
* `/video`: Generate a video.
* `/balance`: Check the remaining quota of your DeepSeek Token.
* `/state`: View the current chat state (including model info and system prompts).
* `/clear`: Clear the current conversation context.
* `/help`: Show command help info.
* `/mode`: Show model info.

