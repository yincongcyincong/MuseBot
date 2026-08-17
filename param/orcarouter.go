package param

// OrcaRouter gateway models.
// OrcaRouter (https://www.orcarouter.ai) is an OpenAI-compatible routing
// gateway: a single API key and base URL expose 150+ models from OpenAI,
// Anthropic, Google, DeepSeek, Qwen and others under namespaced model IDs.
const (
	// OrcaRouterAuto is OrcaRouter's per-request virtual router model.
	OrcaRouterAuto = "orcarouter/auto"
	// OrcaRouterGPT4oMini is a pinned OpenAI model routed through OrcaRouter.
	OrcaRouterGPT4oMini = "openai/gpt-4o-mini"
)
