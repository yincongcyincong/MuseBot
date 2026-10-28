# 模型层韧性配置指南（Failover 与 Utility Model）

MuseBot 的 txt 聊天链路提供两项韧性/成本能力：

- **Provider Failover**：主模型请求彻底失败后，自动降级到其它已配置的 OpenAI 兼容厂商重试
- **Utility Model（辅助模型）**：内部调用（历史压缩、长期记忆、智能模式解析）走一个便宜的小模型，主对话保持大模型

> 范围说明：两项都只作用于 **txt 聊天协议**（OpenAI 兼容请求）。图像/视频/语音识别各厂商接口形状不同，不参与 failover。

---

## 1. Provider Failover

### 工作方式

1. 主 provider（用户配置的 `txt_type`）按原有 `LLMRetryTimes` 预算完整重试，行为与历史版本一致
2. 重试期间出现 **401/403（凭证或配额）、429（限流）** 时，立即跳过剩余重试、直接进入降级——同级重试不可能成功
3. 主 provider 耗尽后，按梯子依次在**每个备选 provider 上尝试一次**（带该 provider 的默认模型），成功即返回，全部失败才报错给用户
4. 流式输出已经吐过内容后中断的，**不再降级**（重发会造成重复内容），维持报错

### 降级链的构成

- 默认：`GetAvailTxtType()` 返回的其它已配置 token 的厂商（deepseek → gemini → openai → aliyun → vol → chatanywhere 的固定顺序，跳过主 provider 和 ollama）
- ollama 永不参与降级：它的协议客户端消息格式不同，无法中途迁移
- 可用 `-failover_providers` 显式指定顺序

### 配置

| 参数 | 环境变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `-enable_failover` | `ENABLE_FAILOVER=false` 关闭 | `true` | 总开关 |
| `-failover_providers` | `FAILOVER_PROVIDERS` | `""` | 显式降级链，逗号分隔，如 `gemini,openai` |

示例：

```bash
./MuseBot-darwin-arm64 -enable_failover=true -failover_providers="gemini,openai"
```

### 观测

- 日志：`failover to backup provider from=deepseek to=gemini model=gemini-2.5-flash` / `failover success ...`
- Prometheus：`api_failover_total{from,to}` 计数器

---

## 2. Utility Model（辅助模型）

### 接管哪些调用

| 调用点 | 频率 | 说明 |
| --- | --- | --- |
| smart mode 命令解析 | 每条普通消息 | 把自然语言解析成命令，对质量要求低 |
| 会话压缩（`CompressHistory`） | 上下文超阈值时 | 生成历史摘要 |
| 长期记忆合并（`consolidateMemories`） | 每用户默认每小时一次 | 合并/去重记忆条目 |

### 配置

| 参数 | 环境变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `-utility_type` | `UTILITY_TYPE` | `""` | 辅助调用的厂商，如 `deepseek` |
| `-utility_model` | `UTILITY_MODEL` | `""` | 辅助调用的模型；**留空 = 不启用**，全部走当前模型 |

两项都配置才生效。示例（内部调用全部走 DeepSeek 便宜档）：

```bash
./MuseBot-darwin-arm64 -utility_type=deepseek -utility_model=deepseek-chat
```

注意：utility 请求固定走指定的厂商+模型，**不参与 failover**，也不读用户级 LLM 配置。

---

## 3. 行为细节

| 行为 | 说明 |
| --- | --- |
| ctx 取消/超时 | 永不触发降级，直接返回错误 |
| 降级后的回复 | 内容直接返回给用户（人设不变），仅在日志与指标中记录切换 |
| 降级尝试次数 | 每个备选 provider 一次（内部仍受 `LLMRetryTimes` 的内层重试约束） |
| 备选 provider 的模型 | `GetTxtModel()` 的该厂商默认模型（如 gemini → `gemini-2.5-flash`） |
| 主 provider 为 ollama | 不降级（客户端协议不同），行为与旧版一致 |
| 关闭 failover 后 | 与引入本功能前逐字节一致 |

---

## 4. 相关文件

| 文件 | 内容 |
| --- | --- |
| `llm/failover.go` | 降级梯子、错误分类、`sendWithFailover` |
| `llm/llm.go` | `WithUtilityModel()` 选项、`providerOverride/modelOverride` |
| `llm/openai.go` | `GetOpenAIClientForType`（按类型建客户端）、`GetModel` 的 override 优先 |
| `metrics/prom.go` | `api_failover_total` 指标 |
