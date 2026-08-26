package param

// Synthorai gateway models.
// Synthorai (https://synthorai.io) is an OpenAI-compatible gateway: one API key
// and base URL expose models from Anthropic, OpenAI, Google, Z.ai, Moonshot,
// DeepSeek, Qwen and others, billed at the upstream provider's own price.
const (
	// SynthoraiDeepSeekV4Flash is the default model: cheap and general purpose.
	SynthoraiDeepSeekV4Flash = "deepseek-v4-flash"
	// SynthoraiClaudeOpus5 is a pinned Anthropic model routed through Synthorai.
	SynthoraiClaudeOpus5 = "claude-opus-5"
)
