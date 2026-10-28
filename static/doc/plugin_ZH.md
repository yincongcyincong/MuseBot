# 插件化扩展指南（Provider / Channel 注册表）

MuseBot 内部用**注册表**管理两类扩展点：

- **LLM Provider**（`llm` 包）：语音识别（Recognizer）、聊天协议（LLMClient）、OpenAI 兼容端点
- **聊天渠道**（`robot` 包）：Telegram / Discord / Slack / Lark / 钉钉 / 企业微信 / QQ / 微信

核心约定：**每个 Provider / 渠道在自己文件的 `init()` 里注册自己**。新增一个模型或渠道 = 新增一个文件（外加少量配置字段），不需要修改核心分发代码（`StartRobot`、`GetAudioContent`、`NewLLM`、`GetOpenAIClient`）。

---

## 1. 部署者视角：配置无需任何修改

注册表的 key 就是原有配置里的类型字符串（`openai` / `gemini` / `aliyun` / `vol` / `ollama` / `deepseek` / `chatanywhere`），`-type`、`-media_type`、各家 `*_TOKEN`、用户级 LLM 配置的语义全部不变，升级后直接运行。

升级后唯一的新行为是**启动诊断日志**：每个因配置缺失而未启动的渠道，会打印一条 INFO 日志说明缺哪个环境变量，例如：

```text
INFO channel disabled, missing config channel=com_wechat missing=[COM_WECHAT_SECRET, COM_WECHAT_AGENT_ID, COM_WECHAT_ENCODING_AES_KEY]
```

排查"为什么我的 XX 机器人没起来"时直接看启动日志即可。另外，识别类型（`rec_type` / `MEDIA_TYPE`）配置成了未注册的值时，原来会静默返回空文本，现在会打 Warn 日志并报错 `no recognizer registered for type "xxx"`。

---

## 2. 新增 LLM Provider

### 2.1 OpenAI 兼容模型（最常见）

以接入 `moonshot` 为例：

**第 1 步**：`param/param.go` 加类型常量：

```go
Moonshot = "moonshot"
```

**第 2 步**：`conf/conf.go` 加 token 字段、flag 和 env 读取（照抄任意一家现有厂商）：

```go
flag.StringVar(&BaseConfInfo.MoonshotToken, "moonshot_token", "", "Moonshot token")
// applyEnvConf 中:
if os.Getenv("MOONSHOT_TOKEN") != "" {
    BaseConfInfo.MoonshotToken = os.Getenv("MOONSHOT_TOKEN")
}
```

**第 3 步**：`llm/openai.go` 的 `openaiEndpoints` 表加一行（token + baseURL）：

```go
param.Moonshot: {token: func() string { return conf.BaseConfInfo.MoonshotToken }, baseURL: "https://api.moonshot.cn/v1"},
```

**第 4 步**：`utils/llm.go` 的 `GetAvailTxtType()`（以及需要暴露图像/视频能力的对应函数）加可用性判断——有 token 才算可用：

```go
if conf.BaseConfInfo.MoonshotToken != "" {
    res = append(res, param.Moonshot)
}
```

到此完成：用户 LLM 配置里把 `txt_type` 设为 `moonshot` 即可聊天（走 OpenAI 兼容协议，默认 `OpenAIReq` 客户端）。

**可选**——如果该厂商提供语音识别，在新建的 `llm/moonshot.go` 里注册 Recognizer（见 2.3）；不注册则它不会出现在可用识别类型里。

### 2.2 非自定义协议的聊天客户端

厂商协议不是 OpenAI 兼容格式时，实现 `llm.LLMClient` 接口（`Send` / `SyncSend` / `GetMessage` / `GetImageMessage` / `GetAudioMessage` / `AppendMessages` / `GetModel`，完整定义见 `llm/llm.go`），然后注册：

```go
// llm/moonshot.go
func init() {
    RegisterChatClient(param.Moonshot, func() LLMClient { return &MoonshotReq{} })
}
```

查不到注册项的类型一律回落 OpenAI 兼容客户端（`OpenAIReq`）。完整示例参考 `llm/ollama.go`。

### 2.3 语音识别（Recognizer）

```go
// 签名: 音频字节 → 文本, token 消耗, 错误
type Recognizer func(ctx context.Context, audioContent []byte) (string, int, error)
```

```go
func init() {
    RegisterRecognizer(param.Moonshot, func(ctx context.Context, audioContent []byte) (string, int, error) {
        text, token, err := moonshotTranscribe(ctx, audioContent)
        return text, token, err
    })
}
```

约定：

- 无法返回精确 token 时，返回 `param.AudioTokenUsage`（固定计费值）即可
- 识别请求需要特殊凭证时（不共用厂商主 token），把实现放到 `utils` 包，注册处做适配，参考 `llm/vol.go`（火山音频走 `utils.FileRecognize`）

---

## 3. 新增聊天渠道

**第 1 步**：新建 `robot/xxx.go`，实现 `StartXxxRobot(ctx context.Context)`。共享消息状态直接嵌入 `baseRobot`（见 `robot/robot.go`），再实现 `Robot` 接口的平台差异方法，完整例子参考 `robot/qq.go`。

**第 2 步**：同文件 `init()` 里注册：

```go
func init() {
    RegisterChannel(Channel{
        Name: param.Xxx,
        MissingCfg: func() []string {
            var missing []string
            missing = required(missing, "XXX_APP_ID", conf.BaseConfInfo.XxxAppID)
            return required(missing, "XXX_APP_SECRET", conf.BaseConfInfo.XxxAppSecret)
        },
        Start: StartXxxRobot,
    })
}
```

约定：

- `Name` 使用 `param` 包里的渠道常量
- `MissingCfg` 返回**环境变量名**，启动日志会原样展示给部署者
- `Start` 内**不要**再自行检查配置是否为空——`startChannels` 已经把没配齐的渠道挡在门外
- `Start` 必须是非阻塞的（内部自己起 goroutine / 长轮询循环），与现有平台一致

**第 3 步**：`conf/conf.go` 加凭证字段 + flag + env 读取（同 2.1 第 2 步）。

---

## 4. 行为细节

| 行为 | 说明 |
| --- | --- |
| 注册时机 | 全部发生在 `init()`，早于 `main()`，`StartRobot()` 只读注册表 |
| 渠道启动顺序 | 各文件 `init()` 按文件名字母序执行，渠道按注册顺序依次 `go Start(ctx)`，顺序无功能影响 |
| 重复注册 | Recognizer / ChatClient 后注册覆盖先注册；渠道则追加 |
| 非法注册 | 名称为空或回调为 nil 直接 panic，进程启动即失败（fail-fast），不会带病运行 |

---

## 5. 相关文件索引

| 文件 | 内容 |
| --- | --- |
| `llm/registry.go` | `RegisterRecognizer` / `RecognizeAudio` / `RegisterChatClient` / `NewChatClient` |
| `llm/openai.go` | `openaiEndpoints`：OpenAI 兼容厂商的 token + baseURL 表 |
| `llm/ollama.go` | 自有协议聊天客户端注册示例 |
| `llm/vol.go` | 特殊凭证 Recognizer 注册示例 |
| `robot/registry.go` | `Channel` / `RegisterChannel` / `required` / `startChannels` |
| `utils/llm.go` | `GetAvailTxtType` 等类型可用性判断（token 存在即可用） |

> 另外，**工具扩展**（MCP Server，`conf/mcp`）与**技能扩展**（`conf/skills` 的 `SKILL.md`）是免编译的文件级扩展，放置配置即生效，参见 [functioncall](functioncall_ZH.md) 与 [skills](skills_ZH.md) 文档。
