

# ✨ 钉钉机器人

本项目是一个跨平台聊天机器人，支持 **钉钉** 私聊和群聊。
它由 **LLM** 提供智能对话能力，内置图片、视频生成，余额查询，会话清空等多种功能。

---

## 钉钉配置
去平台：https://open-dev.dingtalk.com/      

1. 设置stream模式    
![image](https://github.com/user-attachments/assets/9af2185b-f129-49b7-91e2-30a3608e7c23)

2. 权限开通
![image](https://github.com/user-attachments/assets/47014a94-a696-442f-b7fb-559c70dc4223)


## 🚀 在钉钉模式启动

你可以使用以下命令启动钉钉模式的机器人：

```bash
./MuseBot-darwin-amd64 \
  -ding_client_id=xx \
  -ding_client_secret=xx \
  -deepseek_token=sk-xxx \
  -gemini_token=xxx \
  -openai_token=xxx \
  -vol_token=xxx
```

---

## 💬 使用方法

### 1. 私聊机器人

在钉钉中直接私聊机器人，输入指令即可。 <img width="400" alt="image" src="https://github.com/user-attachments/assets/f73094cb-7ae1-4ea5-a8be-75bac24b4a4c" />

支持的指令：

* `/photo` — 生成图片

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/b3634f64-43ef-4884-9212-b8fadba5a474" />

* `/video` — 生成视频

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/ba39278d-83e8-4db0-9a72-b4f0d8c0785b" />

* `/state` — 查看当前会话状态（包括模型信息和系统提示）

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/adfc9015-6d2b-4663-80fa-34491a6f9a8a" /> 

* `/clear` — 清空当前会话上下文

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/47b74d3e-1425-402e-a117-882a8003bbe9" />

* `/help` — 显示帮助信息

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/2971ff23-9b68-4dbc-ad06-ad92b3c12bc8" />

* `/mode` — 切换 LLM 模式

  <img width="400" alt="image" src="https://github.com/user-attachments/assets/c8020a59-2619-4160-beac-5fd628b62e4c" />  

<img width="400" alt="image" src="https://github.com/user-attachments/assets/5327afba-9160-44e8-a0e5-f05bae6cbfd6" />

---
