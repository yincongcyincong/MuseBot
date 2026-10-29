# Model-Layer Resilience Guide (Failover and Utility Model)

MuseBot's txt chat pipeline ships two resilience/cost features:

- **Provider Failover**: after the primary model request fails completely, automatically degrade and retry on other configured OpenAI-compatible providers
- **Utility Model**: route internal calls (history compression, long-term memory, smart-mode parsing) to a cheap small model, keeping the big model for the main conversation

> Scope: both features apply to the **txt chat protocol** only (OpenAI-compatible requests). Image/video/speech-recognition APIs differ per vendor and do not participate in failover.

---

## 1. Provider Failover

### How it works

1. The primary provider (the user's `txt_type`) exhausts its original `LLMRetryTimes` budget, identical to previous behavior
2. If **401/403 (credentials or quota) or 429 (rate limit)** appears during those retries, the remaining retries are skipped and failover starts immediately — retrying the same provider cannot succeed
3. Once the primary provider is exhausted, the ladder tries **once on each backup provider** (with that provider's default model); the first success is returned, and only a fully failed ladder reports an error to the user
4. If a streaming response already emitted content before breaking, **no failover** happens (resending would duplicate content); the error stands

### Failover chain

- Default: the other token-configured providers returned by `GetAvailTxtType()` (fixed order deepseek → gemini → openai → aliyun → vol → chatanywhere, skipping the primary provider and ollama)
- ollama never joins the ladder: its client message format differs, so a request cannot migrate mid-flight
- `-failover_providers` overrides the order explicitly; each entry may pin a model for custom relays with `provider:model` (e.g. `openai:gpt-4o-mini`), falling back to the provider default

### Configuration

| Flag | Env var | Default | Description |
| --- | --- | --- | --- |
| `-enable_failover` | `ENABLE_FAILOVER=false` disables | `true` | master switch |
| `-failover_providers` | `FAILOVER_PROVIDERS` | `""` | explicit failover chain, comma separated, e.g. `gemini,openai`; entries may pin a model: `openai:gpt-4o-mini` |

Example:

```bash
./MuseBot-darwin-arm64 -enable_failover=true -failover_providers="gemini,openai"
```

### Observability

- Logs: `failover to backup provider from=deepseek to=gemini model=gemini-2.5-flash` / `failover success ...`
- Prometheus: `api_failover_total{from,to}` counter

---

## 2. Utility Model

### Which calls it takes over

| Call site | Frequency | Notes |
| --- | --- | --- |
| smart-mode command parsing | every ordinary message | parses natural language into a command; quality bar is low |
| history compression (`CompressHistory`) | when the context exceeds the threshold | produces the history summary |
| long-term memory merge (`consolidateMemories`) | once per hour per user by default | merges/dedups memory entries |

### Configuration

| Flag | Env var | Default | Description |
| --- | --- | --- | --- |
| `-utility_type` | `UTILITY_TYPE` | `""` | provider for utility calls, e.g. `deepseek` |
| `-utility_model` | `UTILITY_MODEL` | `""` | model for utility calls; **empty = disabled**, everything uses the current model |

Both flags must be set. Example (all internal calls on DeepSeek's cheap tier):

```bash
./MuseBot-darwin-arm64 -utility_type=deepseek -utility_model=deepseek-chat
```

Note: utility requests go straight to the configured provider+model; they **do not participate in failover** and never read per-user LLM config.

---

## 3. Behavior details

| Behavior | Description |
| --- | --- |
| ctx cancel/timeout | never triggers failover; the error is returned directly |
| replies after failover | content is returned to the user as-is (persona unchanged); the switch is only recorded in logs and metrics |
| attempts per backup | one per backup provider (still bounded by the inner `LLMRetryTimes` retry) |
| backup provider's model | that provider's default from `GetTxtModel()` (e.g. gemini → `gemini-2.5-flash`) |
| primary provider is ollama | no failover (different client protocol); behavior identical to older versions |
| with failover disabled | byte-for-byte identical to before this feature |

---

## 4. Related files

| File | Contents |
| --- | --- |
| `llm/failover.go` | failover ladder, error classification, `sendWithFailover` |
| `llm/llm.go` | `WithUtilityModel()` option, `providerOverride/modelOverride` |
| `llm/openai.go` | `GetOpenAIClientForType` (client per type), override priority in `GetModel` |
| `metrics/prom.go` | `api_failover_total` metric |
