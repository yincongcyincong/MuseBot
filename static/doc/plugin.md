# Plugin-Style Extension Guide (Provider / Channel Registries)

MuseBot manages two kinds of extension points through **registries**:

- **LLM providers** (`llm` package): speech recognition (Recognizer), chat protocol (LLMClient), OpenAI-compatible endpoints
- **Chat channels** (`robot` package): Telegram / Discord / Slack / Lark / DingTalk / ComWechat / QQ / WeChat

Core convention: **every provider / channel registers itself in its own file's `init()`**. Adding a model or a channel means adding one file (plus a few config fields) — no changes to the core dispatch code (`StartRobot`, `GetAudioContent`, `NewLLM`, `GetOpenAIClient`).

---

## 1. For operators: no configuration changes required

Registry keys are the same type strings the existing configuration already uses (`openai` / `gemini` / `aliyun` / `vol` / `ollama` / `deepseek` / `chatanywhere`). The semantics of `-type`, `-media_type`, the `*_TOKEN` variables and per-user LLM configs are unchanged — upgrade and run as before.

The only new behavior is **startup diagnostics**: every channel that is not started due to missing configuration logs one INFO line naming the missing environment variables, e.g.:

```text
INFO channel disabled, missing config channel=com_wechat missing=[COM_WECHAT_SECRET, COM_WECHAT_AGENT_ID, COM_WECHAT_ENCODING_AES_KEY]
```

When a recognition type (`rec_type` / `MEDIA_TYPE`) is set to an unregistered value, MuseBot used to silently return empty text; it now logs a warning and returns the error `no recognizer registered for type "xxx"`.

---

## 2. Adding an LLM provider

### 2.1 OpenAI-compatible models (most common)

Example: wiring up `moonshot`.

**Step 1** — add the type constant in `param/param.go`:

```go
Moonshot = "moonshot"
```

**Step 2** — add the token field, flag and env read in `conf/conf.go` (copy any existing vendor):

```go
flag.StringVar(&BaseConfInfo.MoonshotToken, "moonshot_token", "", "Moonshot token")
// in applyEnvConf:
if os.Getenv("MOONSHOT_TOKEN") != "" {
    BaseConfInfo.MoonshotToken = os.Getenv("MOONSHOT_TOKEN")
}
```

**Step 3** — add one line to the `openaiEndpoints` table in `llm/openai.go` (token + baseURL):

```go
param.Moonshot: {token: func() string { return conf.BaseConfInfo.MoonshotToken }, baseURL: "https://api.moonshot.cn/v1"},
```

**Step 4** — add the availability check in `utils/llm.go`'s `GetAvailTxtType()` (and the img/video equivalents you want to expose) — a type is available once its token is set:

```go
if conf.BaseConfInfo.MoonshotToken != "" {
    res = append(res, param.Moonshot)
}
```

Done: users set `txt_type` to `moonshot` in their LLM config and chat through the OpenAI-compatible protocol (default `OpenAIReq` client).

**Optional** — if the vendor offers speech recognition, register a Recognizer in a new `llm/moonshot.go` (see 2.3); without one the vendor simply never appears among the available recognition types.

### 2.2 Chat clients with a custom protocol

When the vendor protocol is not OpenAI-compatible, implement the `llm.LLMClient` interface (`Send` / `SyncSend` / `GetMessage` / `GetImageMessage` / `GetAudioMessage` / `AppendMessages` / `GetModel` — see `llm/llm.go`) and register it:

```go
// llm/moonshot.go
func init() {
    RegisterChatClient(param.Moonshot, func() LLMClient { return &MoonshotReq{} })
}
```

Types without a registered chat client fall back to the OpenAI-compatible client (`OpenAIReq`). Full example: `llm/ollama.go`.

### 2.3 Speech recognition (Recognizer)

```go
// signature: audio bytes → text, token cost, error
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

Conventions:

- If exact token accounting is unavailable, return `param.AudioTokenUsage` (the flat default cost)
- If recognition needs dedicated credentials (not the vendor's main token), put the implementation in the `utils` package and adapt at the registration site — see `llm/vol.go` (Volcano Engine audio goes through `utils.FileRecognize`)

---

## 3. Adding a chat channel

**Step 1** — create `robot/xxx.go` with `StartXxxRobot(ctx context.Context)`. Embed `baseRobot` (see `robot/robot.go`) for the shared per-message state, implement the platform-specific methods of the `Robot` interface; a complete example is `robot/qq.go`.

**Step 2** — register in the same file's `init()`:

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

Conventions:

- `Name` uses the channel constants from the `param` package
- `MissingCfg` returns **environment variable names** — they are shown verbatim in the startup log
- `Start` must **not** re-check whether configuration is present — `startChannels` already gates unconfigured channels off
- `Start` must be non-blocking (spawn its own goroutines / long-poll loop), like every existing platform

**Step 3** — add the credential fields + flag + env read in `conf/conf.go` (same as step 2 in 2.1).

---

## 4. Behavior details

| Behavior | Details |
| --- | --- |
| Registration timing | All registrations happen in `init()`, before `main()`; `StartRobot()` only reads the registry |
| Channel start order | File `init()`s run in file-name order; channels start as goroutines in registration order — order has no functional impact |
| Duplicate registration | Recognizer / ChatClient: later registration wins; channels are appended |
| Invalid registration | Empty name or nil callback panics immediately at process start (fail-fast) — no half-initialized runtime |

---

## 5. File index

| File | Contents |
| --- | --- |
| `llm/registry.go` | `RegisterRecognizer` / `RecognizeAudio` / `RegisterChatClient` / `NewChatClient` |
| `llm/openai.go` | `openaiEndpoints`: token + baseURL table for OpenAI-compatible vendors |
| `llm/ollama.go` | Example: custom-protocol chat client registration |
| `llm/vol.go` | Example: Recognizer with dedicated credentials |
| `robot/registry.go` | `Channel` / `RegisterChannel` / `required` / `startChannels` |
| `utils/llm.go` | `GetAvailTxtType` etc. — type availability checks (token present = available) |

> Separately, **tool extensions** (MCP servers, `conf/mcp`) and **skill extensions** (`SKILL.md` under `conf/skills`) are compile-free, file-level extensions that take effect by dropping in configuration — see the [functioncall](functioncall.md) and [skills](skills.md) docs.
