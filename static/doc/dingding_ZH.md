# ✨ 钉钉机器人

本项目是一个跨平台聊天机器人，支持 **钉钉** 私聊和群聊。
它由 **LLM** 提供智能对话能力，内置图片、视频生成，余额查询，会话清空等多种功能。

---

## 🚀 在钉钉模式启动

你可以使用以下命令启动钉钉模式的机器人：

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

## 💬 使用方法

### 1. 私聊机器人

在钉钉中直接私聊机器人应用，输入指令即可。

支持的指令：

* `/photo` — 生成图片
* `/video` — 生成视频
* `/balance` — 查询 DeepSeek Token 余额
* `/state` — 查看当前会话状态（包括模型信息和系统提示）
* `/clear` — 清空当前会话上下文
* `/help` — 显示帮助信息
* `/mode` — 切换 LLM 模式


