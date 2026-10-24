# Advanced Configuration

This guide covers a few commonly used but easily missed settings: connecting a third-party OpenAI-compatible model, smart_mode intent routing, and task tool approvals with auto-approval.

---

## 1. Third-Party Models (CUSTOM_URL)

All text requests in MuseBot use the OpenAI-compatible format. `TYPE` selects the built-in API endpoint and token, while **`CUSTOM_URL`, once set, overrides every built-in endpoint** — so any OpenAI-compatible third-party service works out of the box.

### Configuration

| Environment Variable | Description                                                  |
|----------------------|--------------------------------------------------------------|
| `TYPE`               | Set to `openai` (uses `OPENAI_TOKEN`, no built-in endpoint)  |
| `OPENAI_TOKEN`       | API key of the third-party platform                          |
| `CUSTOM_URL`         | OpenAI-compatible base URL (usually ends with `/v1`)         |
| `DEFAULT_MODEL`      | Default text model name (use the platform's model naming)    |

Run locally:

```sh
go run main.go -telegram_bot_token=your-telegram-token \
  -type=openai \
  -openai_token=sk-xxx \
  -custom_url=https://open.bigmodel.cn/api/paas/v4 \
  -default_model=glm-4.6
```

Run with Docker:

```sh
docker run -d -v /home/user/data:/app/data \
  -e TELEGRAM_BOT_TOKEN="your-telegram-token" \
  -e TYPE="openai" \
  -e OPENAI_TOKEN="sk-xxx" \
  -e CUSTOM_URL="https://open.bigmodel.cn/api/paas/v4" \
  -e DEFAULT_MODEL="glm-4.6" \
  -p 36060:36060 -p 18080:18080 --name my-bot jackyin0822/musebot:latest
```

### Common Third-Party Services

| Service                    | CUSTOM_URL                                           |
|----------------------------|------------------------------------------------------|
| Zhipu GLM                  | `https://open.bigmodel.cn/api/paas/v4`               |
| Kimi (Moonshot)            | `https://api.moonshot.cn/v1`                         |
| Qwen (DashScope compatible mode) | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| DeepSeek                   | `https://api.deepseek.com/v1` (or simply `TYPE=deepseek`) |
| Local Ollama               | `http://127.0.0.1:11434/v1`                          |
| OneAPI / NewAPI gateway    | `http://your-host:3000/v1`                           |

### Notes

- **Image recognition needs a vision model**: `DEFAULT_MODEL` must accept image input (e.g. `glm-4.6`, `qwen-vl-*`, `gpt-4o`); the default `gpt-3.5-turbo` does not.
- Use the `/txt_type` and `/txt_model` commands to switch model type and name at runtime.
- Image/video **generation** is configured separately (`MEDIA_TYPE`: openai/gemini/vol/aliyun) and is unrelated to `CUSTOM_URL`. See [image config](./imageconf.md) and [video config](./videoconf.md).

---

## 2. smart_mode

With `SMART_MODE` (default `true`) enabled, ordinary chat messages first go through intent classification and are routed automatically:

| Message pattern                          | Routing                                   |
|------------------------------------------|-------------------------------------------|
| Simple Q&A, small talk                   | Answered directly like `/chat`            |
| Complex multi-step tasks (coding, research, config changes) | Handed to the `/task` multi-agent flow |
| "Remind me every day at 9..."            | A `/cron` job is created automatically    |
| An image asking to be analyzed           | Handed to `/task` image recognition       |

While classifying, the bot sends an "Analyzing your request..." notice.

```sh
# flag / environment variable
go run main.go -smart_mode=true
SMART_MODE=true ./MuseBot
```

When disabled, every message is treated as plain chat and complex tasks require the explicit `/task` command. You can also toggle the `smart_mode` config key at runtime from the admin panel (`http://your-host:18080`).

---

## 3. Task Tool Approvals & Auto-Approval (TASK_AUTO_APPROVE)

When the `/task` workflow performs a sensitive operation (changing bot config, running commands, expanding the file sandbox), it asks for authorization first by default:

```
⚠️ Task operation requires approval.
Tool: bot_config
Target: config.set smart_mode
Detail: ...
Reply `/y` within 5 minutes to approve, or `/n` to deny.
```

- Reply **`/y`** (or `/approve`) to approve, **`/n`** (or `/deny`) to deny; you may include the 8-digit approval token from the message;
- Unanswered requests expire after 5 minutes and the step fails;
- Only one pending approval exists per user per chat.

### Auto-Approval

In trusted environments you can skip manual confirmation entirely:

```sh
# flag
go run main.go -task_auto_approve=true

# environment variable
TASK_AUTO_APPROVE=true
```

With auto-approval on, every task tool operation is executed immediately without an `/y` prompt. The `task_auto_approve` config key can also be toggled at runtime from the admin panel.

### Security Notes

- Auto-approval means tasks can modify bot config and run allowlisted commands without confirmation — enable it only for **private or trusted deployments**;
- Even with auto-approval, shell commands remain restricted by the `ALLOWED_COMMANDS` allowlist; newly approved commands are persisted in `data/runtime_boundary.json` and survive restarts.

---

## Related Docs

- [Commands & parameters overview](./param_conf.md)
- [Function Call / MCP tools](./functioncall.md)
- [Admin panel](./admin.md)
- [Cron jobs](./cron.md)
