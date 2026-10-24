# 进阶配置

本文介绍几个常用但容易漏配的能力：接入第三方 OpenAI 兼容模型、smart_mode 智能路由、任务工具审批与自动同意。

---

## 一、接入第三方模型（CUSTOM_URL）

MuseBot 所有文本请求都走 OpenAI 兼容格式。`TYPE` 决定使用哪个内置 API 地址和 Token，而 **`CUSTOM_URL` 一旦设置会覆盖所有内置地址**，所以任何 OpenAI 兼容的第三方服务都能直接接入。

### 配置方法

| 环境变量      | 说明                                                         |
|-----------|------------------------------------------------------------|
| `TYPE`    | 固定填 `openai`（使用 `OPENAI_TOKEN`，且不带内置地址）                    |
| `OPENAI_TOKEN` | 第三方平台的 API Key                                        |
| `CUSTOM_URL` | 第三方 OpenAI 兼容 BaseURL（一般以 `/v1` 结尾）                     |
| `DEFAULT_MODEL` | 默认文本模型名（按第三方平台的模型命名填写）                              |

本地运行：

```sh
go run main.go -telegram_bot_token=your-telegram-token \
  -type=openai \
  -openai_token=sk-xxx \
  -custom_url=https://open.bigmodel.cn/api/paas/v4 \
  -default_model=glm-4.6
```

Docker 运行：

```sh
docker run -d -v /home/user/data:/app/data \
  -e TELEGRAM_BOT_TOKEN="your-telegram-token" \
  -e TYPE="openai" \
  -e OPENAI_TOKEN="sk-xxx" \
  -e CUSTOM_URL="https://open.bigmodel.cn/api/paas/v4" \
  -e DEFAULT_MODEL="glm-4.6" \
  -p 36060:36060 -p 18080:18080 --name my-bot jackyin0822/musebot:latest
```

### 常见第三方服务参考

| 服务              | CUSTOM_URL                                        |
|-----------------|---------------------------------------------------|
| 智谱 GLM          | `https://open.bigmodel.cn/api/paas/v4`            |
| Kimi (Moonshot) | `https://api.moonshot.cn/v1`                      |
| 通义千问（DashScope 兼容模式） | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| DeepSeek        | `https://api.deepseek.com/v1`（也可直接 `TYPE=deepseek`） |
| Ollama 本地模型     | `http://127.0.0.1:11434/v1`                       |
| OneAPI / NewAPI 网关 | `http://your-host:3000/v1`                     |

### 注意事项

- **图片识别需要视觉模型**：`DEFAULT_MODEL` 必须支持图片输入（如 `glm-4.6`、`qwen-vl-*`、`gpt-4o`），默认的 `gpt-3.5-turbo` 不支持。
- 运行中可用 `/txt_type`、`/txt_model` 命令临时切换模型类型和名称。
- 图片/视频**生成**走的是另一套配置（`MEDIA_TYPE`：openai/gemini/vol/aliyun），与 `CUSTOM_URL` 无关，详见[图片参数](./imageconf_ZH.md)与[视频参数](./videoconf_ZH.md)。

---

## 二、smart_mode 智能模式

`SMART_MODE`（默认 `true`）开启后，普通聊天消息会先经过一次意图分类，自动路由到合适的能力：

| 用户输入特征          | 自动路由                                  |
|-----------------|---------------------------------------|
| 简单问答、闲聊        | 直接按 `/chat` 回答                        |
| 复杂多步骤任务（写代码、调研、操作配置） | 转入 `/task` 多智能体流程               |
| "每天 9 点提醒我…"    | 自动创建 `/cron` 定时任务                     |
| 发送图片要求识别        | 转入 `/task` 图片识别                       |

分类耗时期间机器人会发送"正在分析你的需求..."提示。

```sh
# 启动参数 / 环境变量
go run main.go -smart_mode=true
SMART_MODE=true ./MuseBot
```

关闭后所有消息都按普通聊天处理，复杂任务需手动使用 `/task` 命令。也可以在管理后台（`http://your-host:18080`）里通过配置项 `smart_mode` 运行时开关。

---

## 三、任务工具审批与自动同意（TASK_AUTO_APPROVE）

`/task` 流程中涉及敏感操作时（如修改机器人配置、执行命令、扩展文件沙箱范围），默认会先向你发送授权请求：

```
⚠️ 任务操作需要授权。
工具：bot_config
目标：config.set smart_mode
详情：...
请在 5 分钟内回复 `/y` 授权，或回复 `/n` 拒绝。
```

- 回复 **`/y`**（或 `/approve`）授权，**`/n`**（或 `/deny`）拒绝，可附带消息中的 8 位授权码；
- 5 分钟未回复自动过期，任务该步骤失败；
- 同一用户同一会话同时只有一个待授权操作。

### 自动同意配置

如果部署环境受信任、不想每次手动确认，可以开启自动同意：

```sh
# 启动参数
go run main.go -task_auto_approve=true

# 环境变量
TASK_AUTO_APPROVE=true
```

开启后所有任务工具操作直接放行，不再发送 `/y` 确认。也可在管理后台通过配置项 `task_auto_approve` 运行时开关。

### 安全提醒

- 自动同意 = 任务可以不经确认修改机器人配置、执行允许列表内的命令，建议仅在**私有部署或受信任环境**开启；
- 即使开启自动同意，shell 命令仍受 `ALLOWED_COMMANDS` 白名单约束；运行中新批准的命令会记录在 `data/runtime_boundary.json`，重启后依然生效。

---

## 相关文档

- [命令与参数总览](./param_conf_ZH.md)
- [Function Call / MCP 工具](./functioncall_ZH.md)
- [管理后台](./admin_ZH.md)
- [Cron 定时任务](./cron_ZH.md)
