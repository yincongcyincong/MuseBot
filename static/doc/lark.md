
# ✨ Lark (Feishu) DeepSeek Bot

This project is a cross-platform chatbot powered by the **DeepSeek LLM**, supporting **Telegram**, **Slack**, **Discord**, and **Lark (Feishu)**.
It comes with a variety of built-in commands, including image and video generation, conversation clearing, and more.

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

### create bot
go to web: https://open.feishu.cn/app/    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/4c96862e-3d90-48ad-a491-6d459ebebcc2" />    


define auth:     
<img width="400" alt="image" src="https://github.com/user-attachments/assets/27f6747c-bd44-4ad2-ae4c-c600078d93e5" />
<img width="400" alt="image" src="https://github.com/user-attachments/assets/bded8047-1994-4018-b885-4f68dae3eb99" />

choose subscription method and event:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/302d5aa8-863c-4a6c-92fc-2f9348b0e147" />


### Private Chat with the Bot

You can directly chat with the bot in Lark via **Private Chat**.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/462f1b06-8d75-427c-afe0-0f77cc85bb2f" />    

Supported commands:

* `/photo`: Generate an image.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/b32a54e9-fb17-4baf-a284-42d44156e776" />

* `/video`: Generate a video.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/2d903781-2ad8-4e1a-9b34-dd99dc398688" />

* `/state`: View the current chat state (including model info and system prompts)    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/50cb16f8-f94e-459b-85a0-c25dec10afaa" />


* `/clear`: Clear the current conversation context
<img width="400" alt="image" src="https://github.com/user-attachments/assets/deb65625-3e51-4581-a6d1-736de4ad7c5e" />


