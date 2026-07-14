# MuseBot 架构（基于代码梳理）

## 1. 程序入口与启动顺序

### 主服务 `main.go`
```go
func main() {
    logger.InitLogger()      // 日志
    conf.InitConf()          // 加载配置（文件 + 环境变量）
    i18n.InitI18n()          // 国际化
    db.InitTable()           // 建表（SQLite / MySQL）
    conf.InitTools()         // 加载 MCP 工具 → Function Call 定义
    rag.InitRag()            // RAG 向量库初始化
    http.InitHTTP()          // 内部/回调 HTTP 服务
    metrics.RegisterMetrics()// Prometheus 指标
    robot.StartRobot()       // 按配置拉起各 IM 平台
    register.InitRegister()  // etcd 服务注册
    robot.InitCron()         // 定时任务调度
    // 阻塞等待信号退出
}
```
初始化顺序即模块依赖顺序：配置→存储→工具/RAG→HTTP→机器人→注册/定时。

### 管理后台 `admin/main.go`
```go
func main() {
    logger.InitLogger()
    conf.InitConfig()
    controller.InitSessionStore()   // 登录会话
    db.InitTable()                  // admin 自己的库（bots/users）
    checkpoint.InitStatusCheck()    // 实例健康检查/etcd 发现
    // net/http mux 注册 /user/* /bot/* 等路由
}
```

## 2. 模块划分（顶层包）

| 包 | 职责 |
|----|------|
| `conf` | 配置结构体 + 加载/热更新；含 MCP、Cron、Grafana、systemd 子资源 |
| `db` | 数据访问层：`users`、`records`、`rag_files`、`cron` 四张表 |
| `robot` | 各 IM 平台适配 + 统一消息处理主流程 |
| `llm` | 大模型抽象层 + 具体厂商实现 + Function Call/任务编排 |
| `http` | 主服务对外 HTTP：配置管理、webhook 回调、指标、SSE |
| `rag` | RAG 检索链（langchaingo 封装向量库） |
| `register` | etcd 服务注册 |
| `metrics` | Prometheus 指标定义 |
| `param` | 常量、枚举（平台类型、模型名）、错误码 |
| `i18n` / `logger` / `utils` | 国际化 / 日志 / 工具函数 |
| `admin` | 独立后台：`conf`/`db`/`controller`/`checkpoint` |

## 3. robot 层：统一接口 + 多平台实现

### 核心抽象
- **`Robot` 接口**（`robot/robot.go:77`）：定义每个平台必须实现的行为——
  `checkValid()`、`getMsgContent()`、`requestLLM()`、`sendChatMessage()`、`sendImg()`、`sendVideo()`、`getPrompt()/setPrompt()`、`sendVoiceContent()`、`executeLLM()`、`sendMedia()` 等。
- **`RobotInfo` 结构体**：所有平台共享的「宿主」，持有 `Ctx`、`Cancel`、`Robot`（具体实现）、`cs *ContextState`。共用主流程方法都挂在它上面。
- **平台实现**：每个文件一个平台，都内嵌 `RobotInfo` 并实现 `Robot` 接口：
  `telegram.go`、`discord.go`、`slack.go`、`lark.go`、`ding.go`、`wechat.go`、`comwechat.go`、`qq.go`、`personalqq.go`、`web.go`。

### 平台启动分发（`StartRobot`，`robot.go:546`）
按配置项是否非空，逐个 `go StartXxxRobot(ctx)` 拉起：Telegram / Discord / Lark / Slack / Ding / ComWechat / QQ / Wechat。同一进程可同时驱动多个平台。

### 消息处理主流程（`RobotInfo.Exec`，`robot.go:145`）
```
Exec()
 ├─ checkUserAllow / checkGroupAllow      // 用户/群白名单
 ├─ AddUserInfo()                          // 用户不存在则建档（默认 LLM 配置）
 ├─ Robot.checkValid()                     // 平台侧校验
 ├─ smartMode()                            // 可选：先让 LLM 判断意图/是否设 cron
 └─ Robot.requestLLM(getMsgContent())      // 进入 LLM 请求
```

`requestLLM` → `executeLLM()`（各平台实现）通常起两条 goroutine：
```go
go Robot.ExecLLM(prompt, msgChan)         // 生产：调用大模型，结果写入 channel
go Robot.HandleUpdate(msgChan, encoding)  // 消费：从 channel 读取并回发用户
```

### ExecLLM（`robot.go:1332`）——衔接 robot 与 llm
组装上下文（用户图片、人设变量、每条消息长度、MCP 工具集），构造 `llm.NewLLM(...)` 并调用 `CallLLM()`，把流式结果推入 `MessageChan`。

### 其它分支
- **命令/回调**：`ExecCmd`、`handleModelUpdate`、`changeType`/`changeModel`、`showXxxModel`、`cronList/Del/Clear` 等处理 `/model`、`/mode`、`/cron` 等。
- **多模态**：`CreatePhoto`（生图）、`CreateVideo`（生视频）、`GetAudioContent`（语音转文字）、`GetVoiceBaseTTS`/`sendVoice`（TTS 回发）、`recPhoto`（识图）。
- **多智能体**：`sendMultiAgent`（`robot.go:1460`）根据类型构造 `llm.LLMTaskReq`，调 `ExecuteMcp()` 或 `ExecuteTask()`。

## 4. llm 层：模型抽象与调用

### 抽象
- **`LLMClient` 接口**（`llm/llm.go`）：`Send`（流式）、`SyncSend`（同步）、`GetMessage`/`GetImageMessage`/`GetAudioMessage`、`AppendMessages`、`GetModel`。
- **`LLM` 结构体**：承载一次请求的全部状态——内容、图片、模型名、各厂商工具集（Deepseek/Vol/OpenAI/Gemini/OpenRouter Tools）、`MessageChan`、上下文 `Cs`。
- **实现选择**（`NewLLM`，`llm.go:151`）：根据用户配置 `utils.GetTxtType(...)`：
  - `Ollama` → `OllamaReq`
  - 其余全部 → `OpenAIReq`（OpenAI 兼容协议，覆盖 OpenAI/DeepSeek/Vol豆包/Aliyun通义/302/OpenRouter/Gemini，具体模型名在 `OpenAIReq.GetModel` 内 `switch` 决定）。

### 调用主路径（`CallLLM`，`llm.go:84`）
```
CallLLM()
 ├─ GetContent / GetMessages   // 取历史+当前内容组装消息
 ├─ InsertCharacter           // 注入人设（Go template，支持 username 等变量）
 ├─ GetModel                  // 按 provider 定模型
 ├─ metrics.APIRequestCount++ // 打点
 ├─ if IsStreaming: Send() else: SyncSend()  // 流式/同步
 │     └─ 期间可触发 Function Call（RequestToolsCall），最多 MostLoop=15 轮
 └─ InsertOrUpdate()          // 落库对话记录 + token 用量
```

### Function Call / MCP（`llm/mcp.go`、`conf/tools_conf.go`）
`conf.InitTools()` 把 MCP server 的工具转换成各厂商的 tool 定义（`DeepseekTools`/`VolTools`/`OpenAITools`/`GeminiTools`），随请求下发。模型返回 tool_call 时，`OpenAIReq.RequestToolsCall` 执行工具、把结果作为新消息回灌，循环直至无工具调用。

### 任务编排（`llm/task.go`，多智能体）
`ExecuteTask()`：先用 `assign_task_prompt` 让 LLM 拆解出 `TaskInfo.Plan` → `loopTask` 逐个子任务 `requestTask` 执行（可再触发工具）→ 最后 `summary_task_prompt` 汇总。无计划则直接回答。

### 多模态生成（独立函数）
`GenerateOpenAIImg` / `GenerateMixImg`（302.AI）/ `Generate302AIVideo` / `GenerateOpenAIText`（ASR）/ `OpenAITTS`，分散在 `openai.go`、`mix.go`。

## 5. 数据流总览

```
IM 平台消息
   │  (各平台 SDK / webhook)
   ▼
robot.RobotInfo.Exec ──► 白名单/建档/校验/smartMode
   │
   ▼ requestLLM → executeLLM
   ├── go ExecLLM ──► llm.NewLLM → CallLLM
   │                     ├─ 组装消息 + 人设 + MCP 工具
   │                     ├─ OpenAIReq/OllamaReq.Send|SyncSend
   │                     │     └─ Function Call 循环(≤15) / task 编排
   │                     ├─ metrics 打点
   │                     └─ db 落库 records + token
   │                        └─ 结果 → MessageChan
   └── go HandleUpdate ◄── MessageChan  ──► 回发（文本/图片/视频/TTS语音）
```

## 6. 存储（db 包）

| 表 | 关键内容 | 主要方法 |
|----|---------|---------|
| `users` | 用户、LLMConfig(JSON)、token 用量 | `InsertUser`/`GetUserByID`/`UpdateUserLLMConfig`/`AddToken` |
| `records` | 对话记录、图片记录、token | `InsertMsgRecord`/`GetMsgRecord`/`GetLastImageRecord`/`GetTokenByUserIdAndTime` |
| `rag_files` | RAG 文件与向量 ID 映射 | `InsertRagFile`/`UpdateVectorIdByFileMd5` |
| `cron` | 定时任务配置 | `InsertCron`/`GetActiveCrons`/`UpdateCronStatus` |

请求上下文里的用户信息通过 `db.GetCtxUserInfo(ctx)` 取用（贯穿 llm 层选择模型/配置）。

## 7. 定时任务（Cron）

`robot.InitCron()`（`robot/cron.go:21`）：延迟 10s 后从 `cron` 表加载启用的任务，用 `robfig/cron`（秒级）注册，回调 `ExecCron`。`ExecCron` 按 `Type` 分发到对应平台的 `ExecXxx`，构造一个 `SkipCheck` 的 `RobotInfo` 走正常 LLM 流程，把结果推送到配置的 target。`smartMode` 也能让用户用自然语言创建 cron（`InsertCron`）。

## 8. HTTP 接口（http 包，`http.go:47`）

主服务同时暴露一组 HTTP 端点，供管理后台调用及作为部分平台的 webhook 入口：
- 配置：`/conf/get|update`、`/command/get`、`/restart`、`/stop`、`/log`
- MCP：`/mcp/get|update|disable|delete|sync`
- 用户/记录：`/user/list`、`/user/token/add`、`/record/list`、`/user/insert/record`
- RAG：`/rag/list|create|delete|get|clear`
- Cron：`/cron/create|update|update_status|delete|list`
- 平台回调：`/com/wechat`、`/wechat`、`/qq`、`/onebot`、`/communicate`（Web API，SSE 流式）
- 运维：`/metrics`、`/pong`（健康检查）、`/dashboard`、`/image`

## 9. 服务注册与后台纳管

- **主服务注册**（`register/etcd.go`）：若配置 `RegisterConfInfo.Type == "etcd"`，起后台 goroutine 持续把自身以 lease 租约写入 `/services/musebot/{md5(host)}/{botName}`，实现自动上线/心跳续约。
- **后台发现**（`admin/checkpoint/bot.go`）：
  - 配了 etcd → `InitEtcdRegister` 从 `/services/musebot/` 前缀拉取实例并入库；
  - 未配 etcd → `ManualCheckPoint` 定时对每个 bot 的 `/pong` 做批量健康检查，写入 `BotMap`（online/offline）。
- **后台控制**（`admin/controller/bot.go`）：通过 HTTP 反向调用各主服务的接口，实现远程 CreateBot/Restart/Stop、配置/MCP 下发、用户与记录查询、日志拉取、`/communicate` 转发（SSE）。

## 10. 配置系统（conf 包）

- `BaseConf`（`conf.go:17`）集中所有平台 token、LLM 开关、`IsStreaming`、`SmartMode`、`Character`(人设模板) 等；另有 `LLMConf`、`AudioConf`、`RagConf`、`RegisterConf` 等分域配置。
- `InitConf` 从配置文件 + 环境变量加载；`SaveConf`/`loadConf` 支持运行期读写（配合 `/conf/update` 热更新）。
- MCP 配置在 `conf/mcp/mcp.json`，Cron 模板在 `conf/cron/`，Grafana/systemd 资源随包提供。

---
**分层小结**：`robot`（平台适配、消息编排）→ `llm`（模型抽象、Function Call、任务编排）→ `db`/`conf`/`rag`（状态与配置）；`http`+`register`+`admin` 构成运维与多实例纳管面；`metrics`+`logger` 为横切关注点。
