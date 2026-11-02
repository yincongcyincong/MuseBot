# 🔔 Webhook 事件触发自动化

MuseBot 除了定时触发（cron），还支持**事件触发**：外部系统一调用，bot 立即推送消息——CI 失败、GitHub 事件、监控告警，或任意脚本里的一行 `curl`。

执行链路与 cron 完全一致：渲染后的 prompt（可选拼接命令前缀，如 `/img`）通过同一套合成 robot 机制投递到目标平台。

## 🚀 快速开始

1. 在内部 HTTP API（默认端口 `:36060`）上创建 webhook：

```bash
curl -X POST http://127.0.0.1:36060/webhook/create -H 'Content-Type: application/json' -d '{
  "webhook_name": "ci-notify",
  "type": "telegram",
  "target_id": "123456789",
  "prompt_template": "CI finished: {{.status}} on {{.repo}} ({{.url}})"
}'
```

响应（token **只展示这一次**，数据库只存其 SHA-256）：

```json
{"code": 0, "data": {"id": "1", "url_path": "/webhook/64位十六进制...", "token": "64位十六进制..."}}
```

2. 把外部系统的回调指向触发端点：

```bash
curl -X POST http://your-bot:36060/webhook/<token> -H 'Content-Type: application/json' \
  -d '{"status": "failed", "repo": "MuseBot", "url": "https://ci.example/run/42"}'
```

端点立即返回 `202 Accepted` 并异步执行（轻松满足 GitHub 的 10 秒超时）。非 2xx 响应使用真实 HTTP 状态码（`404` webhook 不存在或已禁用、`400` 载荷非 JSON、`413` 载荷超限），发送方能正确重试。

## 📝 Prompt 模板

请求的 JSON body 会被解析并暴露给 Go 的 `text/template`：

| 模板 | 载荷 `{"msg":"hi","nested":{"env":"prod"}}` 渲染结果 |
| :--- | :--- |
| `event: {{.msg}}` | `event: hi` |
| `env={{.nested.env}}` | `env=prod` |
| 缺失的键 | `<no value>`（永不报错） |

渲染结果截断到 8 KB，防止超大载荷灌爆 LLM 上下文。模板在创建/更新时即校验合法性。

可选的 `command` 字段会拼接在 prompt 前（`"/img"` + prompt 即触发画图命令），与 cron 一致。

## 📡 渠道支持

Webhook 与 cron 复用同一投递链路，渠道支持矩阵完全一致。创建时把渠道填进 `type`，`target_id`（私聊）或 `group_id`（群聊）二选一填：

| 渠道 | `type` | `target_id`（私聊） | `group_id`（群聊） |
| :--- | :--- | :--- | :--- |
| **Telegram** | `telegram` | 用户 `userId` | 群 `ChatID` |
| **微信（公众号）** | `wechat` | 用户 `userId` | 不支持 |
| **企业微信** | `com_wechat` | 用户 `userId` | 不支持 |
| **飞书** | `lark` | `chatId` | `chatId` |
| **Slack** | `slack` | — | Channel ID |
| **个人QQ** | `personal_qq` | QQ 号 | 群号 |

不适合事件推送的渠道：

| 渠道 | 原因 |
| :--- | :--- |
| **钉钉**（`ding`） | 机制上在投递链路里，但目标是**会话级临时 webhook，很快过期**，无法用于持久自动化。 |
| **QQ**（`qq`）/ **Discord** | 官方 bot 协议只允许回复用户消息，无法主动推送。 |

平台侧注意（对所有主动推送都成立，cron 同样适用）：Telegram bot 只能给已经 `/start` 过它的用户发消息；Slack bot 必须已在目标 channel 中。

## 🛠️ 管理 API

| 端点 | 方法 | 说明 |
| :--- | :--- | :--- |
| `/webhook/create` | POST | 创建；返回 `id`、`url_path` 和一次性 `token` |
| `/webhook/update` | POST | 更新名称/来源/目标/命令/模板（token 不可改） |
| `/webhook/update_status` | POST | `{"id":1,"status":0\|1}` 禁用/启用 |
| `/webhook/delete` | GET | `?id=1` 软删除；token 泄露后删除重建即可轮换 |
| `/webhook/list` | GET | `?page=1&page_size=10&name=` 列表（永不返回 token） |

支持的 `source`：`generic`（JSON 载荷）。`github` / `alertmanager` 适配器在规划中。渠道 `type` 取值与目标标识见[渠道支持](#-渠道支持)一节。

## ⚙️ 配置

| 参数 | 环境变量 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `enable_webhook` | `ENABLE_WEBHOOK=false` | `true` | 是否提供 `/webhook/{token}` 触发端点 |
| `webhook_payload_mb` | `WEBHOOK_PAYLOAD_MB` | `1` | 请求体大小上限（MB） |

## 🔒 安全说明

- Token 为 32 字节随机数；查找走 SHA-256 哈希，明文不落库、不进日志。
- 触发端点对未知或已禁用的 token 统一返回 `404`，不泄露任何信息。
- 管理端点与 `/cron/*` 的暴露模型一致（内网/运维面）。若需将 `:36060` 暴露公网，请用反向代理只转发 `/webhook/<token>` 路径，或启用 mTLS（`crt_file`/`key_file`）。
- 多实例：每个 bot 实例只响应自己名下（`from_bot`）的 webhook，发送方需指向正确实例的端口。
