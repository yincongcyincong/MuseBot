# 🔔 Webhook Event-Triggered Automation

MuseBot can be triggered by external events as well as by time. While cron jobs push messages on a schedule, **webhooks** push messages the moment an external system calls the bot: CI failures, GitHub events, monitoring alerts, or a plain `curl` from any script.

The execution path is identical to cron: the rendered prompt (optionally prefixed with a command such as `/img`) is dispatched to the target platform through the same synthetic-robot machinery.

## 🚀 Quick Start

1. Create a webhook on the internal HTTP API (default port `:36060`):

```bash
curl -X POST http://127.0.0.1:36060/webhook/create -H 'Content-Type: application/json' -d '{
  "webhook_name": "ci-notify",
  "type": "telegram",
  "target_id": "123456789",
  "prompt_template": "CI finished: {{.status}} on {{.repo}} ({{.url}})"
}'
```

Response (the token is shown **exactly once**, the database stores only its SHA-256):

```json
{"code": 0, "data": {"id": "1", "url_path": "/webhook/64-hex-chars...", "token": "64-hex-chars..."}}
```

2. Point your external system at the trigger endpoint:

```bash
curl -X POST http://your-bot:36060/webhook/<token> -H 'Content-Type: application/json' \
  -d '{"status": "failed", "repo": "MuseBot", "url": "https://ci.example/run/42"}'
```

The endpoint answers `202 Accepted` immediately and executes asynchronously (GitHub's 10s timeout is safely met). Non-2xx answers use real HTTP status codes (`404` unknown/disabled webhook, `400` non-JSON payload, `413` oversized) so senders retry correctly.

## 📝 Prompt Template

The JSON body of the request is parsed and exposed to Go `text/template`:

| Template | Payload `{"msg":"hi","nested":{"env":"prod"}}` renders |
| :--- | :--- |
| `event: {{.msg}}` | `event: hi` |
| `env={{.nested.env}}` | `env=prod` |
| missing key | `<no value>` (never an error) |

Rendered prompts are truncated to 8 KB so a huge payload cannot flood the LLM context. Templates are validated when the webhook is created or updated.

An optional `command` field prefixes the prompt (`"/img"` + prompt runs the image command), exactly like cron.

## 📡 Supported Channels

Webhooks reuse the same delivery path as cron, so the channel support matrix is identical. Create the webhook with the channel in `type` and fill `target_id` (direct message) or `group_id` (group message):

| Channel | `type` | `target_id` (direct) | `group_id` (group) |
| :--- | :--- | :--- | :--- |
| **Telegram** | `telegram` | user `userId` | group `ChatID` |
| **WeChat (微信)** | `wechat` | user `userId` | not supported |
| **Com WeChat (企业微信)** | `com_wechat` | user `userId` | not supported |
| **Lark (飞书)** | `lark` | `chatId` | `chatId` |
| **Slack** | `slack` | — | Channel ID |
| **Personal QQ (个人QQ)** | `personal_qq` | QQ number | group number |

Not usable for event push:

| Channel | Reason |
| :--- | :--- |
| **DingTalk (钉钉)** (`ding`) | Wired into the delivery switch, but the target is a short-lived per-session webhook that expires quickly — not suitable for durable automation. |
| **QQ** (`qq`) / **Discord** | Their bot protocols only allow replying to user messages; proactive push is not possible. |

Platform-side caveats that apply to any proactive push (cron included): a Telegram bot can only message users who have already started it (`/start`), and a Slack bot must already be a member of the target channel.

## 🛠️ Management API

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/webhook/create` | POST | Create; returns `id`, `url_path` and the one-time `token` |
| `/webhook/update` | POST | Update name/source/targets/command/template (token is immutable) |
| `/webhook/update_status` | POST | `{"id":1,"status":0\|1}` disable/enable |
| `/webhook/delete` | GET | `?id=1` soft-delete; recreate to rotate a leaked token |
| `/webhook/list` | GET | `?page=1&page_size=10&name=` list (never returns tokens) |

Supported `source`: `generic` (JSON payload). `github` / `alertmanager` adapters are planned. Channel `type` values and target identifiers are listed in [Supported Channels](#-supported-channels).

## ⚙️ Configuration

| Flag | Env | Default | Description |
| :--- | :--- | :--- | :--- |
| `enable_webhook` | `ENABLE_WEBHOOK=false` | `true` | Serve the `/webhook/{token}` trigger endpoint |
| `webhook_payload_mb` | `WEBHOOK_PAYLOAD_MB` | `1` | Max request body size in MB |

## 🔒 Security Notes

- Tokens are 32 random bytes; lookups go through a SHA-256 hash, so the plaintext never touches the database or logs.
- The trigger endpoint answers unknown or disabled tokens with a plain `404`, revealing nothing.
- Management endpoints follow the same exposure model as `/cron/*` (internal/ops network). If you expose `:36060` publicly, front it with a reverse proxy that only forwards `/webhook/<token>` paths, or enable mTLS (`crt_file`/`key_file`).
- Multi-instance: each bot instance only answers webhooks registered by itself (`from_bot` scoping), so point the sender at the right instance's port.
