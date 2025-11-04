# ✨ WeChat DeepSeek Bot

This project is a cross-platform chatbot powered by the **LLM**, supporting **WeChat**.
It comes with a variety of built-in commands, including image and video generation, conversation clearing, and more.

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

1. Go to the [WeChat Official Account Platform](https://mp.weixin.qq.com/). set domain, token and EncodingAESKey    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/ee252dfd-3a93-41d6-b7af-dcaba530f4fd" />


---

### Chat with the Bot

Once connected, you can chat with the bot directly via **WeChat Official Account**.

Supported commands:

* **Normal chat**: Input text and get AI responses.    
* `/photo`: Generate an image.
<img width="400" alt="image" src="https://github.com/user-attachments/assets/1d3ee270-98f1-437d-900f-8dba6b8c9bf0" />

* `/video`: Generate a video.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/8332c9f0-08aa-4f72-a037-6c94c4a97f60" />

* `/state`: View the current chat state (including model info and system prompts).    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/e7e2260e-d279-4660-962a-99dbc0e7d1f9" />

* `/clear`: Clear the current conversation context.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/6c53c15c-7f2a-41ea-8e53-103c1e8c1e24" />

* `/help`: Show command help info.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/d8cfe98c-b424-4e65-8a29-e95320d51e49" />

* `/mode`: Show model info.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/85477f22-2592-41d0-971b-a41e1d80e54a" />
