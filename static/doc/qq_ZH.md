# ✨ QQ Bot

本项目是一个基于 **LLM** 的跨平台聊天机器人，支持 **QQ**。
它内置多种功能，包括图片与视频生成、余额查询、对话清除等。

## 🚀 启动 QQ 模式

你可以使用以下命令启动 QQ 模式：

```bash
./MuseBot-darwin-amd64 \
  -qq_app_id=xxx \
  -qq_app_secret=xxx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

### 参数说明：

* `qq_app_id`：你的 QQ Bot APP ID（必填）
* `qq_app_secret`：你的 QQ Bot APP Secret（必填）


---

## 💬 使用方法

去网页创建机器人: https://q.qq.com/qqbot/

沙盒设置:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/ab2ab481-3a41-41f7-b279-0873175ec6c0" />

回调设置:    
<img width="400" alt="image" src="https://github.com/user-attachments/assets/dd88981f-eca8-4728-a021-c5ebdc0767ca" />


### 私聊机器人

<img width="400" alt="image" src="https://github.com/user-attachments/assets/44394437-ed93-4e89-bb15-a0bbe55ea0e6" />

你可以直接通过 QQ 私聊与机器人对话。

支持的指令：

* `/photo`：生成图片

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/0e15c7cc-cd24-4418-821a-2675d0e2ed9a" />

* `/video`：生成视频

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/8e895453-6e2d-49b0-a3f8-a625404d136e" />

* `/balance`：查询 DeepSeek Token 剩余额度

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/4bff7083-05c4-4645-8c0f-83d53c601eec" />

* `/state`：查看当前会话状态（包括模型信息和系统提示）

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/c6bf87b0-706f-40dc-9aa1-20790af94923" />

* `/clear`：清除当前对话上下文

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/d49ff765-6a62-4a8c-aefd-77fe5bc834e7" />

---

### 群聊模式

在 QQ 群聊中，你可以通过 `@机器人名称` 与它对话，或直接输入命令。

上述所有指令在群聊中也可以直接使用，无需切换到私聊。

