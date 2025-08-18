# ✨ Slack DeepSeek Bot

This project is a cross-platform chatbot powered by the **LLM**, supporting **Slack**. It comes with a variety of built-in commands, including image and video generation, balance checking, conversation clearing, and more.

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

### Create a bot
go to web: https://api.slack.com/apps/    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/15b97d1f-7a50-4c9e-953b-7899a1ecd935" />

Oauth:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/e15ef32e-e4c4-4560-b076-a44e27a8c65e" />
<img width="400" alt="image" src="https://github.com/user-attachments/assets/cfc51f3b-2e57-4575-ae2d-b6d9568921e7" />

Set command:  
<img width="400" alt="image" src="https://github.com/user-attachments/assets/3bd4ba71-c383-42f2-8b6b-47ca4d1c1f32" />

Event subscription:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/8f67c815-d755-4688-983d-647bc64122f1" />


### Private Chat with the Bot

You can directly chat with the bot in Slack via **Direct Message**.    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/08ee0e28-e08c-47f6-825e-45afe2621bba" />

Supported commands:    

* `/photo`: Generate an image.
<img width="400" alt="image" src="https://github.com/user-attachments/assets/078d1899-57eb-4a00-a240-d44cb7dd1a51" />

* `/video`: Generate a video.
<img width="400" alt="image" src="https://github.com/user-attachments/assets/6db607d1-3a73-4a50-a0b9-81899e19f4f6" />

* `/balance`: Check the remaining quota of your DeepSeek Token
<img width="400" alt="image" src="https://github.com/user-attachments/assets/845d4cff-4180-4200-aaf6-dca376d2259c" />

* `/state`: View the current chat state (including model info and system prompts)    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/9c336a10-9250-41a3-9406-4e385fe8d9db" />

* `/clear`: Clear the current conversation context
<img width="400" alt="image" src="https://github.com/user-attachments/assets/11defe94-5642-490a-bc20-d22e5e430f81" />
  
* `/help`: Show command help info    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/1e7da122-7199-4792-94a2-0835d647b9b5" />

* `/mode` Show model info
<img width="400" alt="image" src="https://github.com/user-attachments/assets/225cc31b-8461-4a9a-a036-abd55b151924" />

