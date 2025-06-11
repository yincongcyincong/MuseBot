package param

import (
	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
)

const (
	DeepSeek      = "deepseek"
	DeepSeekLlava = "deepseek-ollama"

	Gemini                        = "gemini"
	ModelGemini25Pro       string = "gemini-2.5-pro"
	ModelGemini25Flash     string = "gemini-2.5-flash"
	ModelGemini20Flash     string = "gemini-2.0-flash"
	ModelGemini20FlashLite string = "gemini-2.0-flash-lite"
	ModelGemini15Pro       string = "gemini-1.5-pro"
	ModelGemini15Flash     string = "gemini-1.5-flash"
	ModelGemini10Ultra     string = "gemini-1.0-ultra"
	ModelGemini10Pro       string = "gemini-1.0-pro"
	ModelGemini10Nano      string = "gemini-1.0-nano"

	// 特定功能模型
	ModelGeminiFlashPreviewTTS string = "gemini-flash-preview-tts"
	ModelGeminiEmbedding       string = "gemini-embedding"
	ModelImagen3               string = "imagen-3"
	ModelVeo2                  string = "veo-2"

	OpenAi = "openai"

	OpenRouter = "openrouter"

	LLAVA = "llava:latest"

	ImageTokenUsage = 10000
	VideoTokenUsage = 20000
)

// open router model
const (
	// Google models
	GoogleGemini2_5ProPreview                = "google/gemini-2.5-pro-preview"
	GoogleGemma3nE4bItFree                   = "google/gemma-3n-e4b-it:free"
	GoogleGemini2_5FlashPreview05_20         = "google/gemini-2.5-flash-preview-05-20"
	GoogleGemini2_5FlashPreview05_20Thinking = "google/gemini-2.5-flash-preview-05-20:thinking"
	GoogleGemini2_5ProPreview05_06           = "google/gemini-2.5-pro-preview-05-06"
	GoogleGemini2_5FlashPreview              = "google/gemini-2.5-flash-preview"
	GoogleGemini2_5FlashPreviewThinking      = "google/gemini-2.5-flash-preview:thinking"
	GoogleGemma3_1bItFree                    = "google/gemma-3-1b-it:free"
	GoogleGemma3_4bItFree                    = "google/gemma-3-4b-it:free"
	GoogleGemma3_4bIt                        = "google/gemma-3-4b-it"
	GoogleGemma3_12bItFree                   = "google/gemma-3-12b-it:free"
	GoogleGemma3_12bIt                       = "google/gemma-3-12b-it"
	GoogleGemma3_27bItFree                   = "google/gemma-3-27b-it:free"
	GoogleGemma3_27bIt                       = "google/gemma-3-27b-it"
	GoogleGemini2_0FlashLite001              = "google/gemini-2.0-flash-lite-001"
	GoogleGemini2_0Flash001                  = "google/gemini-2.0-flash-001"
	GoogleGemini2_0FlashExpFree              = "google/gemini-2.0-flash-exp:free"
	GoogleGeminiFlash1_5_8b                  = "google/gemini-flash-1.5-8b"
	GoogleGemini2_5ProExp03_25               = "google/gemini-2.5-pro-exp-03-25"
	GoogleGemma2_27bIt                       = "google/gemma-2-27b-it"
	GoogleGemma2_9bItFree                    = "google/gemma-2-9b-it:free"
	GoogleGemma2_9bIt                        = "google/gemma-2-9b-it"
	GoogleGeminiFlash1_5                     = "google/gemini-flash-1.5"
	GoogleGeminiPro1_5                       = "google/gemini-pro-1.5"
	GoogleGeminiExp1121                      = "google/gemini-exp-1121"
	GoogleGeminiExp1114                      = "google/gemini-exp-1114"
	GoogleGeminiFlash1_5Exp                  = "google/gemini-flash-1.5-exp"
	GoogleGeminiPro1_5Exp                    = "google/gemini-pro-1.5-exp"
	GooglePalm2ChatBison32k                  = "google/palm-2-chat-bison-32k"
	GooglePalm2CodechatBison32k              = "google/palm-2-codechat-bison-32k"
	GooglePalm2ChatBison                     = "google/palm-2-chat-bison"
	GooglePalm2CodechatBison                 = "google/palm-2-codechat-bison"
	GoogleGemma2bIt                          = "google/gemma-2b-it"
	GoogleGemma7bIt                          = "google/gemma-7b-it"

	// Sentientagi models
	SentientagiDobbyMiniUnhingedPlusLlama3_1_8b = "sentientagi/dobby-mini-unhinged-plus-llama-3.1-8b"

	// Deepseek models
	DeepseekDeepseekR1DistillQwen7b       = "deepseek/deepseek-r1-distill-qwen-7b"
	DeepseekDeepseekR1_0528Qwen3_8bFree   = "deepseek/deepseek-r1-0528-qwen3-8b:free"
	DeepseekDeepseekR1_0528Qwen3_8b       = "deepseek/deepseek-r1-0528-qwen3-8b"
	DeepseekDeepseekR1_0528Free           = "deepseek/deepseek-r1-0528:free"
	DeepseekDeepseekR1_0528               = "deepseek/deepseek-r1-0528"
	DeepseekDeepseekProverV2Free          = "deepseek/deepseek-prover-v2:free"
	DeepseekDeepseekProverV2              = "deepseek/deepseek-prover-v2"
	DeepseekDeepseekV3BaseFree            = "deepseek/deepseek-v3-base:free"
	DeepseekDeepseekChatV3_0324Free       = "deepseek/deepseek-chat-v3-0324:free"
	DeepseekDeepseekChatV3_0324           = "deepseek/deepseek-chat-v3-0324"
	DeepseekDeepseekR1ZeroFree            = "deepseek/deepseek-r1-zero:free"
	DeepseekDeepseekR1DistillLlama8b      = "deepseek/deepseek-r1-distill-llama-8b"
	DeepseekDeepseekR1DistillQwen1_5b     = "deepseek/deepseek-r1-distill-qwen-1.5b"
	DeepseekDeepseekR1DistillQwen32bFree  = "deepseek/deepseek-r1-distill-qwen-32b:free"
	DeepseekDeepseekR1DistillQwen32b      = "deepseek/deepseek-r1-distill-qwen-32b"
	DeepseekDeepseekR1DistillQwen14bFree  = "deepseek/deepseek-r1-distill-qwen-14b:free"
	DeepseekDeepseekR1DistillQwen14b      = "deepseek/deepseek-r1-distill-qwen-14b"
	DeepseekDeepseekR1DistillLlama70bFree = "deepseek/deepseek-r1-distill-llama-70b:free"
	DeepseekDeepseekR1DistillLlama70b     = "deepseek/deepseek-r1-distill-llama-70b"
	DeepseekDeepseekR1Free                = "deepseek/deepseek-r1:free"
	DeepseekDeepseekR1                    = "deepseek/deepseek-r1"
	DeepseekDeepseekChatFree              = "deepseek/deepseek-chat:free"
	DeepseekDeepseekChat                  = "deepseek/deepseek-chat"
	DeepseekDeepseekChatV2_5              = "deepseek/deepseek-chat-v2.5"
	DeepseekDeepseekCoder                 = "deepseek/deepseek-coder"

	// Sarvamai models
	SarvamaiSarvamMFree = "sarvamai/sarvam-m:free"

	// Thedrummer models
	ThedrummerValkyrie49bV1   = "thedrummer/valkyrie-49b-v1"
	ThedrummerAnubisPro105bV1 = "thedrummer/anubis-pro-105b-v1"
	ThedrummerSkyfall36bV2    = "thedrummer/skyfall-36b-v2"
	ThedrummerUnslopnemo12b   = "thedrummer/unslopnemo-12b"
	ThedrummerRocinante12b    = "thedrummer/rocinante-12b"

	// Anthropic models
	AnthropicClaudeOpus4                 = "anthropic/claude-opus-4"
	AnthropicClaudeSonnet4               = "anthropic/claude-sonnet-4"
	AnthropicClaude3_7Sonnet             = "anthropic/claude-3.7-sonnet"
	AnthropicClaude3_7SonnetBeta         = "anthropic/claude-3.7-sonnet:beta"
	AnthropicClaude3_7SonnetThinking     = "anthropic/claude-3.7-sonnet:thinking"
	AnthropicClaude3_5HaikuBeta          = "anthropic/claude-3.5-haiku:beta"
	AnthropicClaude3_5Haiku              = "anthropic/claude-3.5-haiku"
	AnthropicClaude3_5Haiku20241022Beta  = "anthropic/claude-3.5-haiku-20241022:beta"
	AnthropicClaude3_5Haiku20241022      = "anthropic/claude-3.5-haiku-20241022"
	AnthropicClaude3_5SonnetBeta         = "anthropic/claude-3.5-sonnet:beta"
	AnthropicClaude3_5Sonnet             = "anthropic/claude-3.5-sonnet"
	AnthropicClaude3_5Sonnet20240620Beta = "anthropic/claude-3.5-sonnet-20240620:beta"
	AnthropicClaude3_5Sonnet20240620     = "anthropic/claude-3.5-sonnet-20240620"
	AnthropicClaude3HaikuBeta            = "anthropic/claude-3-haiku:beta"
	AnthropicClaude3Haiku                = "anthropic/claude-3-haiku"
	AnthropicClaude3OpusBeta             = "anthropic/claude-3-opus:beta"
	AnthropicClaude3Opus                 = "anthropic/claude-3-opus"
	AnthropicClaude3SonnetBeta           = "anthropic/claude-3-sonnet:beta"
	AnthropicClaude3Sonnet               = "anthropic/claude-3-sonnet"
	AnthropicClaude2_1Beta               = "anthropic/claude-2.1:beta"
	AnthropicClaude2_1                   = "anthropic/claude-2.1"
	AnthropicClaude2Beta                 = "anthropic/claude-2:beta"
	AnthropicClaude2                     = "anthropic/claude-2"
	AnthropicClaude2_0Beta               = "anthropic/claude-2.0:beta"
	AnthropicClaude2_0                   = "anthropic/claude-2.0"
	AnthropicClaudeInstant1_1            = "anthropic/claude-instant-1.1"
	AnthropicClaudeInstant1              = "anthropic/claude-instant-1"
	AnthropicClaude1                     = "anthropic/claude-1"
	AnthropicClaude1_2                   = "anthropic/claude-1.2"
	AnthropicClaudeInstant1_0            = "anthropic/claude-instant-1.0"

	// Mistralai models
	MistralaiDevstralSmallFree               = "mistralai/devstral-small:free"
	MistralaiDevstralSmall                   = "mistralai/devstral-small"
	MistralaiMistralMedium3                  = "mistralai/mistral-medium-3"
	MistralaiCodestral2501                   = "mistralai/codestral-2501"
	MistralaiMistralSmall24bInstruct2501Free = "mistralai/mistral-small-24b-instruct-2501:free"
	MistralaiMistralSmall24bInstruct2501     = "mistralai/mistral-small-24b-instruct-2501"
	MistralaiMistralSaba                     = "mistralai/mistral-saba"
	MistralaiMistralSmall3_1_24bInstructFree = "mistralai/mistral-small-3.1-24b-instruct:free"
	MistralaiMistralSmall3_1_24bInstruct     = "mistralai/mistral-small-3.1-24b-instruct"
	MistralaiMistralLarge2411                = "mistralai/mistral-large-2411"
	MistralaiMistralLarge2407                = "mistralai/mistral-large-2407"
	MistralaiPixtralLarge2411                = "mistralai/pixtral-large-2411"
	MistralaiMinistral8b                     = "mistralai/ministral-8b"
	MistralaiMinistral3b                     = "mistralai/ministral-3b"
	MistralaiPixtral12b                      = "mistralai/pixtral-12b"
	MistralaiMistralNemoFree                 = "mistralai/mistral-nemo:free"
	MistralaiMistralNemo                     = "mistralai/mistral-nemo"
	MistralaiMistral7bInstructFree           = "mistralai/mistral-7b-instruct:free"
	MistralaiMistral7bInstruct               = "mistralai/mistral-7b-instruct"
	MistralaiMistral7bInstructV0_3           = "mistralai/mistral-7b-instruct-v0.3"
	MistralaiMistralLarge                    = "mistralai/mistral-large"
	MistralaiMistralMedium                   = "mistralai/mistral-medium"
	MistralaiMistralSmall                    = "mistralai/mistral-small"
	MistralaiMistralTiny                     = "mistralai/mistral-tiny"
	MistralaiMistral7bInstructV0_2           = "mistralai/mistral-7b-instruct-v0.2"
	MistralaiMixtral8x7bInstruct             = "mistralai/mixtral-8x7b-instruct"
	MistralaiMixtral8x22bInstruct            = "mistralai/mixtral-8x22b-instruct"
	MistralaiCodestralMamba                  = "mistralai/codestral-mamba"
	MistralaiMixtral8x22b                    = "mistralai/mixtral-8x22b"

	// Arcee-ai models
	ArceeAiCallerLarge      = "arcee-ai/caller-large"
	ArceeAiSpotlight        = "arcee-ai/spotlight"
	ArceeAiMaestroReasoning = "arcee-ai/maestro-reasoning"
	ArceeAiVirtuosoLarge    = "arcee-ai/virtuoso-large"
	ArceeAiCoderLarge       = "arcee-ai/coder-large"
	ArceeAiVirtuosoMediumV2 = "arcee-ai/virtuoso-medium-v2"
	ArceeAiArceeBlitz       = "arcee-ai/arcee-blitz"

	// Microsoft models
	MicrosoftPhi4ReasoningPlusFree  = "microsoft/phi-4-reasoning-plus:free"
	MicrosoftPhi4ReasoningPlus      = "microsoft/phi-4-reasoning-plus"
	MicrosoftPhi4ReasoningFree      = "microsoft/phi-4-reasoning:free"
	MicrosoftMaiDsR1Free            = "microsoft/mai-ds-r1:free"
	MicrosoftPhi4MultimodalInstruct = "microsoft/phi-4-multimodal-instruct"
	MicrosoftPhi4                   = "microsoft/phi-4"
	MicrosoftWizardlm2_8x22b        = "microsoft/wizardlm-2-8x22b"
	MicrosoftPhi3_5Mini128kInstruct = "microsoft/phi-3.5-mini-128k-instruct"
	MicrosoftPhi3Mini128kInstruct   = "microsoft/phi-3-mini-128k-instruct"
	MicrosoftPhi3Medium128kInstruct = "microsoft/phi-3-medium-128k-instruct"
	MicrosoftWizardlm2_7b           = "microsoft/wizardlm-2-7b"
	MicrosoftPhi3Medium4kInstruct   = "microsoft/phi-3-medium-4k-instruct"

	// Inception models
	InceptionMercuryCoderSmallBeta = "inception/mercury-coder-small-beta"

	// Opengvlab models
	OpengvlabInternvl3_14bFree = "opengvlab/internvl3-14b:free"
	OpengvlabInternvl3_2bFree  = "opengvlab/internvl3-2b:free"

	// Meta-llama models
	MetaLlamaLlamaGuard4_12b                = "meta-llama/llama-guard-4-12b"
	MetaLlamaLlama3_3_8bInstructFree        = "meta-llama/llama-3.3-8b-instruct:free"
	MetaLlamaLlamaGuard3_8b                 = "meta-llama/llama-guard-3-8b"
	MetaLlamaLlama4MaverickFree             = "meta-llama/llama-4-maverick:free"
	MetaLlamaLlama4Maverick                 = "meta-llama/llama-4-maverick"
	MetaLlamaLlama4ScoutFree                = "meta-llama/llama-4-scout:free"
	MetaLlamaLlama4Scout                    = "meta-llama/llama-4-scout"
	MetaLlamaLlama3_3_70bInstructFree       = "meta-llama/llama-3.3-70b-instruct:free"
	MetaLlamaLlama3_3_70bInstruct           = "meta-llama/llama-3.3-70b-instruct"
	MetaLlamaLlama3_2_3bInstructFree        = "meta-llama/llama-3.2-3b-instruct:free"
	MetaLlamaLlama3_2_3bInstruct            = "meta-llama/llama-3.2-3b-instruct"
	MetaLlamaLlama3_2_1bInstructFree        = "meta-llama/llama-3.2-1b-instruct:free"
	MetaLlamaLlama3_2_1bInstruct            = "meta-llama/llama-3.2-1b-instruct"
	MetaLlamaLlama3_2_90bVisionInstruct     = "meta-llama/llama-3.2-90b-vision-instruct"
	MetaLlamaLlama3_2_11bVisionInstructFree = "meta-llama/llama-3.2-11b-vision-instruct:free"
	MetaLlamaLlama3_2_11bVisionInstruct     = "meta-llama/llama-3.2-11b-vision-instruct"
	MetaLlamaLlamaGuard2_8b                 = "meta-llama/llama-guard-2-8b"
	MetaLlamaLlama3_1_405bFree              = "meta-llama/llama-3.1-405b:free"
	MetaLlamaLlama3_1_405b                  = "meta-llama/llama-3.1-405b"
	MetaLlamaLlama3_1_8bInstructFree        = "meta-llama/llama-3.1-8b-instruct:free"
	MetaLlamaLlama3_1_8bInstruct            = "meta-llama/llama-3.1-8b-instruct"
	MetaLlamaLlama3_1_405bInstruct          = "meta-llama/llama-3.1-405b-instruct"
	MetaLlamaLlama3_1_70bInstruct           = "meta-llama/llama-3.1-70b-instruct"
	MetaLlamaLlama3_8bInstruct              = "meta-llama/llama-3-8b-instruct"
	MetaLlamaLlama3_70bInstruct             = "meta-llama/llama-3-70b-instruct"
	MetaLlamaLlama2_70bChat                 = "meta-llama/llama-2-70b-chat"
	MetaLlamaCodellama70bInstruct           = "meta-llama/codellama-70b-instruct"
	MetaLlamaCodellama34bInstruct           = "meta-llama/codellama-34b-instruct"
	MetaLlamaLlama2_13bChat                 = "meta-llama/llama-2-13b-chat"
	MetaLlamaLlama3_8b                      = "meta-llama/llama-3-8b"
	MetaLlamaLlama3_70b                     = "meta-llama/llama-3-70b"

	// Qwen models
	QwenQwen3_30bA3bFree            = "qwen/qwen3-30b-a3b:free"
	QwenQwen3_30bA3b                = "qwen/qwen3-30b-a3b"
	QwenQwen3_8bFree                = "qwen/qwen3-8b:free"
	QwenQwen3_8b                    = "qwen/qwen3-8b"
	QwenQwen3_14bFree               = "qwen/qwen3-14b:free"
	QwenQwen3_14b                   = "qwen/qwen3-14b"
	QwenQwen3_32bFree               = "qwen/qwen3-32b:free"
	QwenQwen3_32b                   = "qwen/qwen3-32b"
	QwenQwen3_235bA22bFree          = "qwen/qwen3-235b-a22b:free"
	QwenQwen3_235bA22b              = "qwen/qwen3-235b-a22b"
	QwenQwq32bFree                  = "qwen/qwq-32b:free"
	QwenQwq32b                      = "qwen/qwq-32b"
	QwenQwen2_5Vl3bInstructFree     = "qwen/qwen2.5-vl-3b-instruct:free"
	QwenQwen2_5Vl32bInstructFree    = "qwen/qwen2.5-vl-32b-instruct:free"
	QwenQwen2_5Vl32bInstruct        = "qwen/qwen2.5-vl-32b-instruct"
	QwenQwq32bPreview               = "qwen/qwq-32b-preview"
	QwenQwenVlPlus                  = "qwen/qwen-vl-plus"
	QwenQwenVlMax                   = "qwen/qwen-vl-max"
	QwenQwenTurbo                   = "qwen/qwen-turbo"
	QwenQwen2_5Vl72bInstructFree    = "qwen/qwen2.5-vl-72b-instruct:free"
	QwenQwen2_5Vl72bInstruct        = "qwen/qwen2.5-vl-72b-instruct"
	QwenQwenPlus                    = "qwen/qwen-plus"
	QwenQwenMax                     = "qwen/qwen-max"
	QwenQwen2_5Coder32bInstructFree = "qwen/qwen-2.5-coder-32b-instruct:free"
	QwenQwen2_5Coder32bInstruct     = "qwen/qwen-2.5-coder-32b-instruct"
	QwenQwen2_5_7bInstructFree      = "qwen/qwen-2.5-7b-instruct:free"
	QwenQwen2_5_7bInstruct          = "qwen/qwen-2.5-7b-instruct"
	QwenQwen2_5_72bInstructFree     = "qwen/qwen-2.5-72b-instruct:free"
	QwenQwen2_5_72bInstruct         = "qwen/qwen-2.5-72b-instruct"
	QwenQwen2_5Vl7bInstructFree     = "qwen/qwen-2.5-vl-7b-instruct:free"
	QwenQwen2_5Vl7bInstruct         = "qwen/qwen-2.5-vl-7b-instruct"
	QwenQwen2_72bInstruct           = "qwen/qwen-2-72b-instruct"
	QwenQwen2_7bInstruct            = "qwen/qwen-2-7b-instruct"
	QwenQwen3_0_6b04_28             = "qwen/qwen3-0.6b-04-28"
	QwenQwen3_1_7b                  = "qwen/qwen3-1.7b"
	QwenQwen3_4b                    = "qwen/qwen3-4b"
	QwenQwen2_5Coder7bInstruct      = "qwen/qwen2.5-coder-7b-instruct"
	QwenQwen2_5_32bInstruct         = "qwen/qwen2.5-32b-instruct"
	QwenQwen110bChat                = "qwen/qwen-110b-chat"
	QwenQwen72bChat                 = "qwen/qwen-72b-chat"
	QwenQwen32bChat                 = "qwen/qwen-32b-chat"
	QwenQwen14bChat                 = "qwen/qwen-14b-chat"
	QwenQwen7bChat                  = "qwen/qwen-7b-chat"
	QwenQwen4bChat                  = "qwen/qwen-4b-chat"

	// Tngtech models
	TngtechDeepseekR1tChimeraFree = "tngtech/deepseek-r1t-chimera:free"

	// Thudm models
	ThudmGlmZ1Rumination32b = "thudm/glm-z1-rumination-32b"
	ThudmGlmZ1_32bFree      = "thudm/glm-z1-32b:free"
	ThudmGlmZ1_32b          = "thudm/glm-z1-32b"
	ThudmGlm4_32bFree       = "thudm/glm-4-32b:free"
	ThudmGlm4_32b           = "thudm/glm-4-32b"
	ThudmGlmZ1_9b           = "thudm/glm-z1-9b"
	ThudmGlm4_9b            = "thudm/glm-4-9b"

	// OpenAI models
	OpenaiCodexMini              = "openai/codex-mini"
	OpenaiO4MiniHigh             = "openai/o4-mini-high"
	OpenaiO3                     = "openai/o3"
	OpenaiO4Mini                 = "openai/o4-mini"
	OpenaiGpt4_1                 = "openai/gpt-4.1"
	OpenaiGpt4_1Mini             = "openai/gpt-4.1-mini"
	OpenaiGpt4_1Nano             = "openai/gpt-4.1-nano"
	OpenaiO1Pro                  = "openai/o1-pro"
	OpenaiGpt4oMiniSearchPreview = "openai/gpt-4o-mini-search-preview"
	OpenaiGpt4oSearchPreview     = "openai/gpt-4o-search-preview"
	OpenaiGpt4_5Preview          = "openai/gpt-4.5-preview"
	OpenaiO3MiniHigh             = "openai/o3-mini-high"
	OpenaiO3Mini                 = "openai/o3-mini"
	OpenaiO1                     = "openai/o1"
	OpenaiGpt4o20241120          = "openai/gpt-4o-2024-11-20"
	OpenaiGpt4oMini              = "openai/gpt-4o-mini"
	OpenaiGpt4oMini20240718      = "openai/gpt-4o-mini-2024-07-18"
	OpenaiO1Preview              = "openai/o1-preview"
	OpenaiO1Preview20240912      = "openai/o1-preview-2024-09-12"
	OpenaiO1Mini                 = "openai/o1-mini"
	OpenaiO1Mini20240912         = "openai/o1-mini-2024-09-12"
	OpenaiChatgpt4oLatest        = "openai/chatgpt-4o-latest"
	OpenaiGpt4o20240806          = "openai/gpt-4o-2024-08-06"
	OpenaiGpt4o                  = "openai/gpt-4o"
	OpenaiGpt4oExtended          = "openai/gpt-4o:extended"
	OpenaiGpt4o20240513          = "openai/gpt-4o-2024-05-13"
	OpenaiGpt4Turbo              = "openai/gpt-4-turbo"
	OpenaiGpt3_5Turbo0613        = "openai/gpt-3.5-turbo-0613"
	OpenaiGpt4TurboPreview       = "openai/gpt-4-turbo-preview"
	OpenaiGpt3_5Turbo1106        = "openai/gpt-3.5-turbo-1106"
	OpenaiGpt4_1106Preview       = "openai/gpt-4-1106-preview"
	OpenaiGpt3_5TurboInstruct    = "openai/gpt-3.5-turbo-instruct"
	OpenaiGpt3_5Turbo16k         = "openai/gpt-3.5-turbo-16k"
	OpenaiGpt3_5Turbo            = "openai/gpt-3.5-turbo"
	OpenaiGpt3_5Turbo0125        = "openai/gpt-3.5-turbo-0125"
	OpenaiGpt4                   = "openai/gpt-4"
	OpenaiGpt4_0314              = "openai/gpt-4-0314"
	OpenaiGpt4VisionPreview      = "openai/gpt-4-vision-preview"
	OpenaiGpt4_32k               = "openai/gpt-4-32k"
	OpenaiGpt4_32k0314           = "openai/gpt-4-32k-0314"

	// Eleutherai models
	EleutheraiLlemma7b = "eleutherai/llemma_7b"

	// Alfredpros models
	AlfredprosCodellama7bInstructSolidity = "alfredpros/codellama-7b-instruct-solidity"

	// Arliai models
	ArliaiQwq32bArliaiRprV1Free = "arliai/qwq-32b-arliai-rpr-v1:free"

	// Agentica-org models
	AgenticaOrgDeepcoder14bPreviewFree = "agentica-org/deepcoder-14b-preview:free"

	// Moonshotai models
	MoonshotaiKimiVlA3bThinkingFree       = "moonshotai/kimi-vl-a3b-thinking:free"
	MoonshotaiMoonlight16bA3bInstructFree = "moonshotai/moonlight-16b-a3b-instruct:free"

	// X-ai models
	XAiGrok3MiniBeta   = "x-ai/grok-3-mini-beta"
	XAiGrok3Beta       = "x-ai/grok-3-beta"
	XAiGrok2Vision1212 = "x-ai/grok-2-vision-1212"
	XAiGrok2_1212      = "x-ai/grok-2-1212"
	XAiGrokVisionBeta  = "x-ai/grok-vision-beta"
	XAiGrokBeta        = "x-ai/grok-beta"
	XAiGrok2Mini       = "x-ai/grok-2-mini"
	XAiGrok2           = "x-ai/grok-2"

	// Nvidia models
	NvidiaLlama3_3NemotronSuper49bV1Free  = "nvidia/llama-3.3-nemotron-super-49b-v1:free"
	NvidiaLlama3_3NemotronSuper49bV1      = "nvidia/llama-3.3-nemotron-super-49b-v1"
	NvidiaLlama3_1NemotronUltra253bV1Free = "nvidia/llama-3.1-nemotron-ultra-253b-v1:free"
	NvidiaLlama3_1NemotronUltra253bV1     = "nvidia/llama-3.1-nemotron-ultra-253b-v1"
	NvidiaLlama3_1Nemotron70bInstruct     = "nvidia/llama-3.1-nemotron-70b-instruct"
	NvidiaLlama3_1NemotronNano8bV1        = "nvidia/llama-3.1-nemotron-nano-8b-v1"

	// All-hands models
	AllHandsOpenhandsLm32bV0_1 = "all-hands/openhands-lm-32b-v0.1"

	// Scb10x models
	Scb10xLlama3_1Typhoon2_70bInstruct = "scb10x/llama3.1-typhoon2-70b-instruct"
	Scb10xLlama3_1Typhoon2_8bInstruct  = "scb10x/llama3.1-typhoon2-8b-instruct"

	// Featherless models
	FeatherlessQwerky72bFree = "featherless/qwerky-72b:free"

	// Open-r1 models
	OpenR1Olympiccoder32bFree = "open-r1/olympiccoder-32b:free"

	// Ai21 models
	Ai21Jamba1_6Large = "ai21/jamba-1.6-large"
	Ai21Jamba1_6Mini  = "ai21/jamba-1.6-mini"
	Ai21Jamba1_5Mini  = "ai21/jamba-1.5-mini"
	Ai21Jamba1_5Large = "ai21/jamba-1.5-large"
	Ai21JambaInstruct = "ai21/jamba-instruct"

	// Cohere models
	CohereCommandA            = "cohere/command-a"
	CohereCommandR7b12_2024   = "cohere/command-r7b-12-2024"
	CohereCommandRPlus08_2024 = "cohere/command-r-plus-08-2024"
	CohereCommandR08_2024     = "cohere/command-r-08-2024"
	CohereCommandRPlus        = "cohere/command-r-plus"
	CohereCommandRPlus04_2024 = "cohere/command-r-plus-04-2024"
	CohereCommand             = "cohere/command"
	CohereCommandR            = "cohere/command-r"
	CohereCommandR03_2024     = "cohere/command-r-03-2024"

	// Rekaai models
	RekaaiRekaFlash3Free = "rekaai/reka-flash-3:free"

	// Perplexity models
	PerplexitySonarReasoningPro            = "perplexity/sonar-reasoning-pro"
	PerplexitySonarPro                     = "perplexity/sonar-pro"
	PerplexitySonarDeepResearch            = "perplexity/sonar-deep-research"
	PerplexityR1_1776                      = "perplexity/r1-1776"
	PerplexitySonarReasoning               = "perplexity/sonar-reasoning"
	PerplexitySonar                        = "perplexity/sonar"
	PerplexityLlama3_1SonarSmall128kOnline = "perplexity/llama-3.1-sonar-small-128k-online"
	PerplexityLlama3_1SonarLarge128kOnline = "perplexity/llama-3.1-sonar-large-128k-online"
	PerplexityLlama3SonarLarge32kOnline    = "perplexity/llama-3-sonar-large-32k-online"
	PerplexityLlama3SonarSmall32kChat      = "perplexity/llama-3-sonar-small-32k-chat"
	PerplexityLlama3SonarSmall32kOnline    = "perplexity/llama-3-sonar-small-32k-online"
	PerplexityLlama3SonarLarge32kChat      = "perplexity/llama-3-sonar-large-32k-chat"

	// Cognitivecomputations models
	CognitivecomputationsDolphin3_0R1Mistral24bFree = "cognitivecomputations/dolphin3.0-r1-mistral-24b:free"
	CognitivecomputationsDolphin3_0Mistral24bFree   = "cognitivecomputations/dolphin3.0-mistral-24b:free"
	CognitivecomputationsDolphinMixtral8x22b        = "cognitivecomputations/dolphin-mixtral-8x22b"
	CognitivecomputationsDolphinLlama3_70b          = "cognitivecomputations/dolphin-llama-3-70b"
	CognitivecomputationsDolphinMixtral8x7b         = "cognitivecomputations/dolphin-mixtral-8x7b"

	// Aion-labs models
	AionLabsAion1_0           = "aion-labs/aion-1.0"
	AionLabsAion1_0Mini       = "aion-labs/aion-1.0-mini"
	AionLabsAionRpLlama3_1_8b = "aion-labs/aion-rp-llama-3.1-8b"

	// Liquid models
	LiquidLfm7b  = "liquid/lfm-7b"
	LiquidLfm3b  = "liquid/lfm-3b"
	LiquidLfm40b = "liquid/lfm-40b"

	// Minimax models
	MinimaxMinimax01 = "minimax/minimax-01"

	// Sao10k models
	Sao10kL3_3Euryale70b   = "sao10k/l3.3-euryale-70b"
	Sao10kL3_1_70bHanamiX1 = "sao10k/l3.1-70b-hanami-x1"
	Sao10kL3_1Euryale70b   = "sao10k/l3.1-euryale-70b"
	Sao10kL3Lunaris8b      = "sao10k/l3-lunaris-8b"
	Sao10kL3Euryale70b     = "sao10k/l3-euryale-70b"
	Sao10kFimbulvetr11bV2  = "sao10k/fimbulvetr-11b-v2"
	Sao10kL3Stheno8b       = "sao10k/l3-stheno-8b"

	// Eva-unit-01 models
	EvaUnit01EvaLlama3_33_70b = "eva-unit-01/eva-llama-3.33-70b"
	EvaUnit01EvaQwen2_5_72b   = "eva-unit-01/eva-qwen-2.5-72b"
	EvaUnit01EvaQwen2_5_32b   = "eva-unit-01/eva-qwen-2.5-32b"
	EvaUnit01EvaQwen2_5_14b   = "eva-unit-01/eva-qwen-2.5-14b"

	// Infermatic models
	InfermaticMnInferor12b = "infermatic/mn-inferor-12b"

	// Raifle models
	RaifleSorcererlm8x22b = "raifle/sorcererlm-8x22b"

	// Anthracite-org models
	AnthraciteOrgMagnumV4_72b = "anthracite-org/magnum-v4-72b"
	AnthraciteOrgMagnumV2_72b = "anthracite-org/magnum-v2-72b"

	// Neversleep models
	NeversleepLlama3_1Lumimaid70b         = "neversleep/llama-3.1-lumimaid-70b"
	NeversleepLlama3_1Lumimaid8b          = "neversleep/llama-3.1-lumimaid-8b"
	NeversleepLlama3Lumimaid70b           = "neversleep/llama-3-lumimaid-70b"
	NeversleepLlama3Lumimaid8b            = "neversleep/llama-3-lumimaid-8b"
	NeversleepNoromaid20b                 = "neversleep/noromaid-20b"
	NeversleepNoromaidMixtral8x7bInstruct = "neversleep/noromaid-mixtral-8x7b-instruct"

	// Inflection models
	InflectionInflection3Productivity = "inflection/inflection-3-productivity"
	InflectionInflection3Pi           = "inflection/inflection-3-pi"

	// Alpindale models
	AlpindaleMagnum72b   = "alpindale/magnum-72b"
	AlpindaleGoliath120b = "alpindale/goliath-120b"

	// Nousresearch models
	NousresearchDeephermes3Mistral24bPreviewFree = "nousresearch/deephermes-3-mistral-24b-preview:free"
	NousresearchDeephermes3Llama3_8bPreviewFree  = "nousresearch/deephermes-3-llama-3-8b-preview:free"
	NousresearchHermes3Llama3_1_70b              = "nousresearch/hermes-3-llama-3.1-70b"
	NousresearchHermes3Llama3_1_405b             = "nousresearch/hermes-3-llama-3.1-405b"
	NousresearchHermes2ProLlama3_8b              = "nousresearch/hermes-2-pro-llama-3-8b"
	NousresearchNousHermes2Mixtral8x7bDpo        = "nousresearch/nous-hermes-2-mixtral-8x7b-dpo"
	NousresearchNousHermes2Mistral7bDpo          = "nousresearch/nous-hermes-2-mistral-7b-dpo"
	NousresearchNousHermes2Mixtral8x7bSft        = "nousresearch/nous-hermes-2-mixtral-8x7b-sft"
	NousresearchNousHermesYi34b                  = "nousresearch/nous-hermes-yi-34b"
	NousresearchNousHermes2Vision7b              = "nousresearch/nous-hermes-2-vision-7b"
	NousresearchNousCapybara7b                   = "nousresearch/nous-capybara-7b"
	NousresearchNousCapybara34b                  = "nousresearch/nous-capybara-34b"
	NousresearchNousHermesLlama2_70b             = "nousresearch/nous-hermes-llama2-70b"
	NousresearchNousHermesLlama2_13b             = "nousresearch/nous-hermes-llama2-13b"
	NousresearchHermes2ThetaLlama3_8b            = "nousresearch/hermes-2-theta-llama-3-8b"

	// Amazon models
	AmazonNovaLiteV1  = "amazon/nova-lite-v1"
	AmazonNovaMicroV1 = "amazon/nova-micro-v1"
	AmazonNovaProV1   = "amazon/nova-pro-v1"

	// 01-ai models
	O1AiYiLarge       = "01-ai/yi-large"
	O1AiYi1_5_34bChat = "01-ai/yi-1.5-34b-chat"
	O1AiYiLargeTurbo  = "01-ai/yi-large-turbo"
	O1AiYiLargeFc     = "01-ai/yi-large-fc"
	O1AiYiVision      = "01-ai/yi-vision"
	O1AiYi34b200k     = "01-ai/yi-34b-200k"
	O1AiYi34b         = "01-ai/yi-34b"
	O1AiYi34bChat     = "01-ai/yi-34b-chat"
	O1AiYi6b          = "01-ai/yi-6b"

	// TokyoTech-LLM models
	TokyotechLlmLlama3_1Swallow8bInstructV0_3 = "tokyotech-llm/llama-3.1-swallow-8b-instruct-v0.3"

	// Openrouter models
	OpenrouterOptimusAlpha = "openrouter/optimus-alpha"
	OpenrouterQuasarAlpha  = "openrouter/quasar-alpha"
	OpenrouterAuto         = "openrouter/auto"
	OpenrouterCinematika7b = "openrouter/cinematika-7b"

	// Allenai models
	AllenaiMolmo7bD               = "allenai/molmo-7b-d"
	AllenaiOlmo2_0325_32bInstruct = "allenai/olmo-2-0325-32b-instruct"
	AllenaiLlama3_1Tulu3_405b     = "allenai/llama-3.1-tulu-3-405b"
	AllenaiOlmo7bInstruct         = "allenai/olmo-7b-instruct"

	// Bytedance-research models
	BytedanceResearchUiTars72b = "bytedance-research/ui-tars-72b"

	// Steelskull models
	SteelskullL3_3ElectraR1_70b = "steelskull/l3.3-electra-r1-70b"

	// Latitudegames models
	LatitudegamesWayfarerLarge70bLlama3_3 = "latitudegames/wayfarer-large-70b-llama-3.3"

	// Inflatebot models
	InflatebotMnMagMellR1 = "inflatebot/mn-mag-mell-r1"

	// Mattshumer models
	MattshumerReflection70b = "mattshumer/reflection-70b"

	// Lynn models
	LynnSoliloquyV3 = "lynn/soliloquy-v3"
	LynnSoliloquyL3 = "lynn/soliloquy-l3"

	// Nvidia models (already defined, adding new ones)
	NvidiaNemotron4_340bInstruct = "nvidia/nemotron-4-340b-instruct"

	// Bigcode models
	BigcodeStarcoder2_15bInstruct = "bigcode/starcoder2-15b-instruct"

	// Openchat models
	OpenchatOpenchat8b = "openchat/openchat-8b"
	OpenchatOpenchat7b = "openchat/openchat-7b"

	// Snowflake models
	SnowflakeSnowflakeArcticInstruct = "snowflake/snowflake-arctic-instruct"

	// Fireworks models
	FireworksFirellava13b = "fireworks/firellava-13b"

	// Huggingfaceh4 models
	Huggingfaceh4ZephyrOrpo141bA35b = "huggingfaceh4/zephyr-orpo-141b-a35b"
	Huggingfaceh4Zephyr7bBeta       = "huggingfaceh4/zephyr-7b-beta"

	// Databricks models
	DatabricksDbrxInstruct = "databricks/dbrx-instruct"

	// Recursal models
	RecursalEagle7b        = "recursal/eagle-7b"
	RecursalRwkv5_3bAiTown = "recursal/rwkv-5-3b-ai-town"

	// Rwkv models
	RwkvRwkv5World3b = "rwkv/rwkv-5-world-3b"

	// Togethercomputer models
	TogethercomputerStripedhyenaNous7b    = "togethercomputer/stripedhyena-nous-7b"
	TogethercomputerStripedhyenaHessian7b = "togethercomputer/stripedhyena-hessian-7b"

	// Koboldai models
	KoboldaiPsyfighter13b2 = "koboldai/psyfighter-13b-2"

	// Gryphe models
	GrypheMythomist7b    = "gryphe/mythomist-7b"
	GrypheMythomaxL2_13b = "gryphe/mythomax-l2-13b"

	// Jebcarter models
	JebcarterPsyfighter13b = "jebcarter/psyfighter-13b"

	// Intel models
	IntelNeuralChat7b = "intel/neural-chat-7b"

	// Teknium models
	TekniumOpenhermes2_5Mistral7b = "teknium/openhermes-2.5-mistral-7b"
	TekniumOpenhermes2Mistral7b   = "teknium/openhermes-2-mistral-7b"

	// Liuhaotian models
	LiuhaotianLlavaYi34b = "liuhaotian/llava-yi-34b"
	LiuhaotianLlava13b   = "liuhaotian/llava-13b"

	// Lizpreciatior models
	LizpreciatiorLzlv70bFp16Hf = "lizpreciatior/lzlv-70b-fp16-hf"

	// Jondurbin models
	JondurbinAiroborosL2_70b = "jondurbin/airoboros-l2-70b"
	JondurbinBagel34b        = "jondurbin/bagel-34b"

	// Xwin-lm models
	XwinLmXwinLm70b = "xwin-lm/xwin-lm-70b"

	// Migtissera models
	MigtisseraSynthia70b = "migtissera/synthia-70b"

	// Phind models
	PhindPhindCodellama34b = "phind/phind-codellama-34b"

	// Mancer models
	MancerWeaver = "mancer/weaver"

	// Undi95 models
	Undi95ToppyM7b        = "undi95/toppy-m-7b"
	Undi95RemmSlerpL2_13b = "undi95/remm-slerp-l2-13b"

	// Austism models
	AustismChronosHermes13b = "austism/chronos-hermes-13b"

	// Nothingiisreal models
	NothingiisrealMnCeleste12b = "nothingiisreal/mn-celeste-12b"

	// Aetherwiing models
	AetherwiingMnStarcannon12b = "aetherwiing/mn-starcannon-12b"

	// Sophosympatheia models
	SophosympatheiaMidnightRose70b = "sophosympatheia/midnight-rose-70b"

	LLMGoogle    = "google"
	LLMDeepseek  = "deepseek"
	LLMOpenai    = "openai"
	LLMQwen      = "qwen"
	LLMMetaLlama = "meta-llama"
	LLMMicrosoft = "microsoft"
	LLMAnthropic = "anthropic"
	LLMMistralai = "mistralai"
)

var (
	GeminiModels = map[string]bool{
		ModelGemini25Pro:       true,
		ModelGemini25Flash:     true,
		ModelGemini20Flash:     true,
		ModelGemini20FlashLite: true,
		ModelGemini15Pro:       true,
		ModelGemini15Flash:     true,
		ModelGemini10Ultra:     true,
		ModelGemini10Pro:       true,
		ModelGemini10Nano:      true,
	}

	DeepseekModels = map[string]bool{
		deepseek.DeepSeekChat:     true,
		deepseek.DeepSeekReasoner: true,
		deepseek.DeepSeekCoder:    true,
	}

	OpenRouterModelTypes = map[string]bool{
		LLMGoogle:    true,
		LLMDeepseek:  true,
		LLMOpenai:    true,
		LLMQwen:      true,
		LLMMetaLlama: true,
		LLMMicrosoft: true,
		LLMAnthropic: true,
		LLMMistralai: true,
	}

	OpenRouterModels = map[string]bool{
		// Google models
		GoogleGemini2_5ProPreview:                true,
		GoogleGemma3nE4bItFree:                   true,
		GoogleGemini2_5FlashPreview05_20:         true,
		GoogleGemini2_5FlashPreview05_20Thinking: true,
		GoogleGemini2_5ProPreview05_06:           true,
		GoogleGemini2_5FlashPreview:              true,
		GoogleGemini2_5FlashPreviewThinking:      true,
		GoogleGemma3_1bItFree:                    true,
		GoogleGemma3_4bItFree:                    true,
		GoogleGemma3_4bIt:                        true,
		GoogleGemma3_12bItFree:                   true,
		GoogleGemma3_12bIt:                       true,
		GoogleGemma3_27bItFree:                   true,
		GoogleGemma3_27bIt:                       true,
		GoogleGemini2_0FlashLite001:              true,
		GoogleGemini2_0Flash001:                  true,
		GoogleGemini2_0FlashExpFree:              true,
		GoogleGeminiFlash1_5_8b:                  true,
		GoogleGemini2_5ProExp03_25:               true,
		GoogleGemma2_27bIt:                       true,
		GoogleGemma2_9bItFree:                    true,
		GoogleGemma2_9bIt:                        true,
		GoogleGeminiFlash1_5:                     true,
		GoogleGeminiPro1_5:                       true,
		GoogleGeminiExp1121:                      true,
		GoogleGeminiExp1114:                      true,
		GoogleGeminiFlash1_5Exp:                  true,
		GoogleGeminiPro1_5Exp:                    true,
		GooglePalm2ChatBison32k:                  true,
		GooglePalm2CodechatBison32k:              true,
		GooglePalm2ChatBison:                     true,
		GooglePalm2CodechatBison:                 true,
		GoogleGemma2bIt:                          true,
		GoogleGemma7bIt:                          true,

		// Sentientagi models
		SentientagiDobbyMiniUnhingedPlusLlama3_1_8b: true,

		// Deepseek models
		DeepseekDeepseekR1DistillQwen7b:       true,
		DeepseekDeepseekR1_0528Qwen3_8bFree:   true,
		DeepseekDeepseekR1_0528Qwen3_8b:       true,
		DeepseekDeepseekR1_0528Free:           true,
		DeepseekDeepseekR1_0528:               true,
		DeepseekDeepseekProverV2Free:          true,
		DeepseekDeepseekProverV2:              true,
		DeepseekDeepseekV3BaseFree:            true,
		DeepseekDeepseekChatV3_0324Free:       true,
		DeepseekDeepseekChatV3_0324:           true,
		DeepseekDeepseekR1ZeroFree:            true,
		DeepseekDeepseekR1DistillLlama8b:      true,
		DeepseekDeepseekR1DistillQwen1_5b:     true,
		DeepseekDeepseekR1DistillQwen32bFree:  true,
		DeepseekDeepseekR1DistillQwen32b:      true,
		DeepseekDeepseekR1DistillQwen14bFree:  true,
		DeepseekDeepseekR1DistillQwen14b:      true,
		DeepseekDeepseekR1DistillLlama70bFree: true,
		DeepseekDeepseekR1DistillLlama70b:     true,
		DeepseekDeepseekR1Free:                true,
		DeepseekDeepseekR1:                    true,
		DeepseekDeepseekChatFree:              true,
		DeepseekDeepseekChat:                  true,
		DeepseekDeepseekChatV2_5:              true,
		DeepseekDeepseekCoder:                 true,

		// Sarvamai models
		SarvamaiSarvamMFree: true,

		// Thedrummer models
		ThedrummerValkyrie49bV1:   true,
		ThedrummerAnubisPro105bV1: true,
		ThedrummerSkyfall36bV2:    true,
		ThedrummerUnslopnemo12b:   true,
		ThedrummerRocinante12b:    true,

		// Anthropic models
		AnthropicClaudeOpus4:                 true,
		AnthropicClaudeSonnet4:               true,
		AnthropicClaude3_7Sonnet:             true,
		AnthropicClaude3_7SonnetBeta:         true,
		AnthropicClaude3_7SonnetThinking:     true,
		AnthropicClaude3_5HaikuBeta:          true,
		AnthropicClaude3_5Haiku:              true,
		AnthropicClaude3_5Haiku20241022Beta:  true,
		AnthropicClaude3_5Haiku20241022:      true,
		AnthropicClaude3_5SonnetBeta:         true,
		AnthropicClaude3_5Sonnet:             true,
		AnthropicClaude3_5Sonnet20240620Beta: true,
		AnthropicClaude3_5Sonnet20240620:     true,
		AnthropicClaude3HaikuBeta:            true,
		AnthropicClaude3Haiku:                true,
		AnthropicClaude3OpusBeta:             true,
		AnthropicClaude3Opus:                 true,
		AnthropicClaude3SonnetBeta:           true,
		AnthropicClaude3Sonnet:               true,
		AnthropicClaude2_1Beta:               true,
		AnthropicClaude2_1:                   true,
		AnthropicClaude2Beta:                 true,
		AnthropicClaude2:                     true,
		AnthropicClaude2_0Beta:               true,
		AnthropicClaude2_0:                   true,
		AnthropicClaudeInstant1_1:            true,
		AnthropicClaudeInstant1:              true,
		AnthropicClaude1:                     true,
		AnthropicClaude1_2:                   true,
		AnthropicClaudeInstant1_0:            true,

		// Mistralai models
		MistralaiDevstralSmallFree:               true,
		MistralaiDevstralSmall:                   true,
		MistralaiMistralMedium3:                  true,
		MistralaiCodestral2501:                   true,
		MistralaiMistralSmall24bInstruct2501Free: true,
		MistralaiMistralSmall24bInstruct2501:     true,
		MistralaiMistralSaba:                     true,
		MistralaiMistralSmall3_1_24bInstructFree: true,
		MistralaiMistralSmall3_1_24bInstruct:     true,
		MistralaiMistralLarge2411:                true,
		MistralaiMistralLarge2407:                true,
		MistralaiPixtralLarge2411:                true,
		MistralaiMinistral8b:                     true,
		MistralaiMinistral3b:                     true,
		MistralaiPixtral12b:                      true,
		MistralaiMistralNemoFree:                 true,
		MistralaiMistralNemo:                     true,
		MistralaiMistral7bInstructFree:           true,
		MistralaiMistral7bInstruct:               true,
		MistralaiMistral7bInstructV0_3:           true,
		MistralaiMistralLarge:                    true,
		MistralaiMistralMedium:                   true,
		MistralaiMistralSmall:                    true,
		MistralaiMistralTiny:                     true,
		MistralaiMistral7bInstructV0_2:           true,
		MistralaiMixtral8x7bInstruct:             true,
		MistralaiMixtral8x22bInstruct:            true,
		MistralaiCodestralMamba:                  true,
		MistralaiMixtral8x22b:                    true,

		// Arcee-ai models
		ArceeAiCallerLarge:      true,
		ArceeAiSpotlight:        true,
		ArceeAiMaestroReasoning: true,
		ArceeAiVirtuosoLarge:    true,
		ArceeAiCoderLarge:       true,
		ArceeAiVirtuosoMediumV2: true,
		ArceeAiArceeBlitz:       true,

		// Microsoft models
		MicrosoftPhi4ReasoningPlusFree:  true,
		MicrosoftPhi4ReasoningPlus:      true,
		MicrosoftPhi4ReasoningFree:      true,
		MicrosoftMaiDsR1Free:            true,
		MicrosoftPhi4MultimodalInstruct: true,
		MicrosoftPhi4:                   true,
		MicrosoftWizardlm2_8x22b:        true,
		MicrosoftPhi3_5Mini128kInstruct: true,
		MicrosoftPhi3Mini128kInstruct:   true,
		MicrosoftPhi3Medium128kInstruct: true,
		MicrosoftWizardlm2_7b:           true,
		MicrosoftPhi3Medium4kInstruct:   true,

		// Inception models
		InceptionMercuryCoderSmallBeta: true,

		// Opengvlab models
		OpengvlabInternvl3_14bFree: true,
		OpengvlabInternvl3_2bFree:  true,

		// Meta-llama models
		MetaLlamaLlamaGuard4_12b:                true,
		MetaLlamaLlama3_3_8bInstructFree:        true,
		MetaLlamaLlamaGuard3_8b:                 true,
		MetaLlamaLlama4MaverickFree:             true,
		MetaLlamaLlama4Maverick:                 true,
		MetaLlamaLlama4ScoutFree:                true,
		MetaLlamaLlama4Scout:                    true,
		MetaLlamaLlama3_3_70bInstructFree:       true,
		MetaLlamaLlama3_3_70bInstruct:           true,
		MetaLlamaLlama3_2_3bInstructFree:        true,
		MetaLlamaLlama3_2_3bInstruct:            true,
		MetaLlamaLlama3_2_1bInstructFree:        true,
		MetaLlamaLlama3_2_1bInstruct:            true,
		MetaLlamaLlama3_2_90bVisionInstruct:     true,
		MetaLlamaLlama3_2_11bVisionInstructFree: true,
		MetaLlamaLlama3_2_11bVisionInstruct:     true,
		MetaLlamaLlamaGuard2_8b:                 true,
		MetaLlamaLlama3_1_405bFree:              true,
		MetaLlamaLlama3_1_405b:                  true,
		MetaLlamaLlama3_1_8bInstructFree:        true,
		MetaLlamaLlama3_1_8bInstruct:            true,
		MetaLlamaLlama3_1_405bInstruct:          true,
		MetaLlamaLlama3_1_70bInstruct:           true,
		MetaLlamaLlama3_8bInstruct:              true,
		MetaLlamaLlama3_70bInstruct:             true,
		MetaLlamaLlama2_70bChat:                 true,
		MetaLlamaCodellama70bInstruct:           true,
		MetaLlamaCodellama34bInstruct:           true,
		MetaLlamaLlama2_13bChat:                 true,
		MetaLlamaLlama3_8b:                      true,
		MetaLlamaLlama3_70b:                     true,

		// Qwen models
		QwenQwen3_30bA3bFree:            true,
		QwenQwen3_30bA3b:                true,
		QwenQwen3_8bFree:                true,
		QwenQwen3_8b:                    true,
		QwenQwen3_14bFree:               true,
		QwenQwen3_14b:                   true,
		QwenQwen3_32bFree:               true,
		QwenQwen3_32b:                   true,
		QwenQwen3_235bA22bFree:          true,
		QwenQwen3_235bA22b:              true,
		QwenQwq32bFree:                  true,
		QwenQwq32b:                      true,
		QwenQwen2_5Vl3bInstructFree:     true,
		QwenQwen2_5Vl32bInstructFree:    true,
		QwenQwen2_5Vl32bInstruct:        true,
		QwenQwq32bPreview:               true,
		QwenQwenVlPlus:                  true,
		QwenQwenVlMax:                   true,
		QwenQwenTurbo:                   true,
		QwenQwen2_5Vl72bInstructFree:    true,
		QwenQwen2_5Vl72bInstruct:        true,
		QwenQwenPlus:                    true,
		QwenQwenMax:                     true,
		QwenQwen2_5Coder32bInstructFree: true,
		QwenQwen2_5Coder32bInstruct:     true,
		QwenQwen2_5_7bInstructFree:      true,
		QwenQwen2_5_7bInstruct:          true,
		QwenQwen2_5_72bInstructFree:     true,
		QwenQwen2_5_72bInstruct:         true,
		QwenQwen2_5Vl7bInstructFree:     true,
		QwenQwen2_5Vl7bInstruct:         true,
		QwenQwen2_72bInstruct:           true,
		QwenQwen2_7bInstruct:            true,
		QwenQwen3_0_6b04_28:             true,
		QwenQwen3_1_7b:                  true,
		QwenQwen3_4b:                    true,
		QwenQwen2_5Coder7bInstruct:      true,
		QwenQwen2_5_32bInstruct:         true,
		QwenQwen110bChat:                true,
		QwenQwen72bChat:                 true,
		QwenQwen32bChat:                 true,
		QwenQwen14bChat:                 true,
		QwenQwen7bChat:                  true,
		QwenQwen4bChat:                  true,

		// Tngtech models
		TngtechDeepseekR1tChimeraFree: true,

		// Thudm models
		ThudmGlmZ1Rumination32b: true,
		ThudmGlmZ1_32bFree:      true,
		ThudmGlmZ1_32b:          true,
		ThudmGlm4_32bFree:       true,
		ThudmGlm4_32b:           true,
		ThudmGlmZ1_9b:           true,
		ThudmGlm4_9b:            true,

		// OpenAI models
		OpenaiCodexMini:              true,
		OpenaiO4MiniHigh:             true,
		OpenaiO3:                     true,
		OpenaiO4Mini:                 true,
		OpenaiGpt4_1:                 true,
		OpenaiGpt4_1Mini:             true,
		OpenaiGpt4_1Nano:             true,
		OpenaiO1Pro:                  true,
		OpenaiGpt4oMiniSearchPreview: true,
		OpenaiGpt4oSearchPreview:     true,
		OpenaiGpt4_5Preview:          true,
		OpenaiO3MiniHigh:             true,
		OpenaiO3Mini:                 true,
		OpenaiO1:                     true,
		OpenaiGpt4o20241120:          true,
		OpenaiGpt4oMini:              true,
		OpenaiGpt4oMini20240718:      true,
		OpenaiO1Preview:              true,
		OpenaiO1Preview20240912:      true,
		OpenaiO1Mini:                 true,
		OpenaiO1Mini20240912:         true,
		OpenaiChatgpt4oLatest:        true,
		OpenaiGpt4o20240806:          true,
		OpenaiGpt4o:                  true,
		OpenaiGpt4oExtended:          true,
		OpenaiGpt4o20240513:          true,
		OpenaiGpt4Turbo:              true,
		OpenaiGpt3_5Turbo0613:        true,
		OpenaiGpt4TurboPreview:       true,
		OpenaiGpt3_5Turbo1106:        true,
		OpenaiGpt4_1106Preview:       true,
		OpenaiGpt3_5TurboInstruct:    true,
		OpenaiGpt3_5Turbo16k:         true,
		OpenaiGpt3_5Turbo:            true,
		OpenaiGpt3_5Turbo0125:        true,
		OpenaiGpt4:                   true,
		OpenaiGpt4_0314:              true,
		OpenaiGpt4VisionPreview:      true,
		OpenaiGpt4_32k:               true,
		OpenaiGpt4_32k0314:           true,

		// Eleutherai models
		EleutheraiLlemma7b: true,

		// Alfredpros models
		AlfredprosCodellama7bInstructSolidity: true,

		// Arliai models
		ArliaiQwq32bArliaiRprV1Free: true,

		// Agentica-org models
		AgenticaOrgDeepcoder14bPreviewFree: true,

		// Moonshotai models
		MoonshotaiKimiVlA3bThinkingFree:       true,
		MoonshotaiMoonlight16bA3bInstructFree: true,

		// X-ai models
		XAiGrok3MiniBeta:   true,
		XAiGrok3Beta:       true,
		XAiGrok2Vision1212: true,
		XAiGrok2_1212:      true,
		XAiGrokVisionBeta:  true,
		XAiGrokBeta:        true,
		XAiGrok2Mini:       true,
		XAiGrok2:           true,

		// Nvidia models
		NvidiaLlama3_3NemotronSuper49bV1Free:  true,
		NvidiaLlama3_3NemotronSuper49bV1:      true,
		NvidiaLlama3_1NemotronUltra253bV1Free: true,
		NvidiaLlama3_1NemotronUltra253bV1:     true,
		NvidiaLlama3_1Nemotron70bInstruct:     true,
		NvidiaLlama3_1NemotronNano8bV1:        true,

		// All-hands models
		AllHandsOpenhandsLm32bV0_1: true,

		// Scb10x models
		Scb10xLlama3_1Typhoon2_70bInstruct: true,
		Scb10xLlama3_1Typhoon2_8bInstruct:  true,

		// Featherless models
		FeatherlessQwerky72bFree: true,

		// Open-r1 models
		OpenR1Olympiccoder32bFree: true,

		// Ai21 models
		Ai21Jamba1_6Large: true,
		Ai21Jamba1_6Mini:  true,
		Ai21Jamba1_5Mini:  true,
		Ai21Jamba1_5Large: true,
		Ai21JambaInstruct: true,

		// Cohere models
		CohereCommandA:            true,
		CohereCommandR7b12_2024:   true,
		CohereCommandRPlus08_2024: true,
		CohereCommandR08_2024:     true,
		CohereCommandRPlus:        true,
		CohereCommandRPlus04_2024: true,
		CohereCommand:             true,
		CohereCommandR:            true,
		CohereCommandR03_2024:     true,

		// Rekaai models
		RekaaiRekaFlash3Free: true,

		// Perplexity models
		PerplexitySonarReasoningPro:            true,
		PerplexitySonarPro:                     true,
		PerplexitySonarDeepResearch:            true,
		PerplexityR1_1776:                      true,
		PerplexitySonarReasoning:               true,
		PerplexitySonar:                        true,
		PerplexityLlama3_1SonarSmall128kOnline: true,
		PerplexityLlama3_1SonarLarge128kOnline: true,
		PerplexityLlama3SonarLarge32kOnline:    true,
		PerplexityLlama3SonarSmall32kChat:      true,
		PerplexityLlama3SonarSmall32kOnline:    true,
		PerplexityLlama3SonarLarge32kChat:      true,

		// Cognitivecomputations models
		CognitivecomputationsDolphin3_0R1Mistral24bFree: true,
		CognitivecomputationsDolphin3_0Mistral24bFree:   true,
		CognitivecomputationsDolphinMixtral8x22b:        true,
		CognitivecomputationsDolphinLlama3_70b:          true,
		CognitivecomputationsDolphinMixtral8x7b:         true,

		// Aion-labs models
		AionLabsAion1_0:           true,
		AionLabsAion1_0Mini:       true,
		AionLabsAionRpLlama3_1_8b: true,

		// Liquid models
		LiquidLfm7b:  true,
		LiquidLfm3b:  true,
		LiquidLfm40b: true,

		// Minimax models
		MinimaxMinimax01: true,

		// Sao10k models
		Sao10kL3_3Euryale70b:   true,
		Sao10kL3_1_70bHanamiX1: true,
		Sao10kL3_1Euryale70b:   true,
		Sao10kL3Lunaris8b:      true,
		Sao10kL3Euryale70b:     true,
		Sao10kFimbulvetr11bV2:  true,
		Sao10kL3Stheno8b:       true,

		// Eva-unit-01 models
		EvaUnit01EvaLlama3_33_70b: true,
		EvaUnit01EvaQwen2_5_72b:   true,
		EvaUnit01EvaQwen2_5_32b:   true,
		EvaUnit01EvaQwen2_5_14b:   true,

		// Infermatic models
		InfermaticMnInferor12b: true,

		// Raifle models
		RaifleSorcererlm8x22b: true,

		// Anthracite-org models
		AnthraciteOrgMagnumV4_72b: true,
		AnthraciteOrgMagnumV2_72b: true,

		// Neversleep models
		NeversleepLlama3_1Lumimaid70b:         true,
		NeversleepLlama3_1Lumimaid8b:          true,
		NeversleepLlama3Lumimaid70b:           true,
		NeversleepLlama3Lumimaid8b:            true,
		NeversleepNoromaid20b:                 true,
		NeversleepNoromaidMixtral8x7bInstruct: true,

		// Inflection models
		InflectionInflection3Productivity: true,
		InflectionInflection3Pi:           true,

		// Alpindale models
		AlpindaleMagnum72b:   true,
		AlpindaleGoliath120b: true,

		// Nousresearch models
		NousresearchDeephermes3Mistral24bPreviewFree: true,
		NousresearchDeephermes3Llama3_8bPreviewFree:  true,
		NousresearchHermes3Llama3_1_70b:              true,
		NousresearchHermes3Llama3_1_405b:             true,
		NousresearchHermes2ProLlama3_8b:              true,
		NousresearchNousHermes2Mixtral8x7bDpo:        true,
		NousresearchNousHermes2Mistral7bDpo:          true,
		NousresearchNousHermes2Mixtral8x7bSft:        true,
		NousresearchNousHermesYi34b:                  true,
		NousresearchNousHermes2Vision7b:              true,
		NousresearchNousCapybara7b:                   true,
		NousresearchNousCapybara34b:                  true,
		NousresearchNousHermesLlama2_70b:             true,
		NousresearchNousHermesLlama2_13b:             true,
		NousresearchHermes2ThetaLlama3_8b:            true,

		// Amazon models
		AmazonNovaLiteV1:  true,
		AmazonNovaMicroV1: true,
		AmazonNovaProV1:   true,

		// 01-ai models
		O1AiYiLarge:       true,
		O1AiYi1_5_34bChat: true,
		O1AiYiLargeTurbo:  true,
		O1AiYiLargeFc:     true,
		O1AiYiVision:      true,
		O1AiYi34b200k:     true,
		O1AiYi34b:         true,
		O1AiYi34bChat:     true,
		O1AiYi6b:          true,

		// TokyoTech-LLM models
		TokyotechLlmLlama3_1Swallow8bInstructV0_3: true,

		// Openrouter models
		OpenrouterOptimusAlpha: true,
		OpenrouterQuasarAlpha:  true,
		OpenrouterAuto:         true,
		OpenrouterCinematika7b: true,

		// Allenai models
		AllenaiMolmo7bD:               true,
		AllenaiOlmo2_0325_32bInstruct: true,
		AllenaiLlama3_1Tulu3_405b:     true,
		AllenaiOlmo7bInstruct:         true,

		// Bytedance-research models
		BytedanceResearchUiTars72b: true,

		// Steelskull models
		SteelskullL3_3ElectraR1_70b: true,

		// Latitudegames models
		LatitudegamesWayfarerLarge70bLlama3_3: true,

		// Inflatebot models
		InflatebotMnMagMellR1: true,

		// Mattshumer models
		MattshumerReflection70b: true,

		// Lynn models
		LynnSoliloquyV3: true,
		LynnSoliloquyL3: true,

		// Nvidia models
		NvidiaNemotron4_340bInstruct: true,

		// Bigcode models
		BigcodeStarcoder2_15bInstruct: true,

		// Openchat models
		OpenchatOpenchat8b: true,
		OpenchatOpenchat7b: true,

		// Snowflake models
		SnowflakeSnowflakeArcticInstruct: true,

		// Fireworks models
		FireworksFirellava13b: true,

		// Huggingfaceh4 models
		Huggingfaceh4ZephyrOrpo141bA35b: true,
		Huggingfaceh4Zephyr7bBeta:       true,

		// Databricks models
		DatabricksDbrxInstruct: true,

		// Recursal models
		RecursalEagle7b:        true,
		RecursalRwkv5_3bAiTown: true,

		// Rwkv models
		RwkvRwkv5World3b: true,

		// Togethercomputer models
		TogethercomputerStripedhyenaNous7b:    true,
		TogethercomputerStripedhyenaHessian7b: true,

		// Koboldai models
		KoboldaiPsyfighter13b2: true,

		// Gryphe models
		GrypheMythomist7b:    true,
		GrypheMythomaxL2_13b: true,

		// Jebcarter models
		JebcarterPsyfighter13b: true,

		// Intel models
		IntelNeuralChat7b: true,

		// Teknium models
		TekniumOpenhermes2_5Mistral7b: true,
		TekniumOpenhermes2Mistral7b:   true,

		// Liuhaotian models
		LiuhaotianLlavaYi34b: true,
		LiuhaotianLlava13b:   true,

		// Lizpreciatior models
		LizpreciatiorLzlv70bFp16Hf: true,

		// Jondurbin models
		JondurbinAiroborosL2_70b: true,
		JondurbinBagel34b:        true,

		// Xwin-lm models
		XwinLmXwinLm70b: true,

		// Migtissera models
		MigtisseraSynthia70b: true,

		// Phind models
		PhindPhindCodellama34b: true,

		// Mancer models
		MancerWeaver: true,

		// Undi95 models
		Undi95ToppyM7b:        true,
		Undi95RemmSlerpL2_13b: true,

		// Austism models
		AustismChronosHermes13b: true,

		// Nothingiisreal models
		NothingiisrealMnCeleste12b: true,

		// Aetherwiing models
		AetherwiingMnStarcannon12b: true,

		// Sophosympatheia models
		SophosympatheiaMidnightRose70b: true,
	}

	OpenAIModels = map[string]bool{
		openai.GPT3Dot5Turbo0125:       true,
		openai.O1Mini:                  true,
		openai.O1Mini20240912:          true,
		openai.O1Preview:               true,
		openai.O1Preview20240912:       true,
		openai.O1:                      true,
		openai.O120241217:              true,
		openai.O3Mini:                  true,
		openai.O3Mini20250131:          true,
		openai.GPT432K0613:             true,
		openai.GPT432K0314:             true,
		openai.GPT432K:                 true,
		openai.GPT40613:                true,
		openai.GPT40314:                true,
		openai.GPT4o:                   true,
		openai.GPT4o20240513:           true,
		openai.GPT4o20240806:           true,
		openai.GPT4o20241120:           true,
		openai.GPT4oLatest:             true,
		openai.GPT4oMini:               true,
		openai.GPT4oMini20240718:       true,
		openai.GPT4Turbo:               true,
		openai.GPT4Turbo20240409:       true,
		openai.GPT4Turbo0125:           true,
		openai.GPT4Turbo1106:           true,
		openai.GPT4TurboPreview:        true,
		openai.GPT4VisionPreview:       true,
		openai.GPT4:                    true,
		openai.GPT4Dot5Preview:         true,
		openai.GPT4Dot5Preview20250227: true,
		openai.GPT3Dot5Turbo1106:       true,
		openai.GPT3Dot5Turbo0613:       true,
		openai.GPT3Dot5Turbo0301:       true,
		openai.GPT3Dot5Turbo16K:        true,
		openai.GPT3Dot5Turbo16K0613:    true,
		openai.GPT3Dot5Turbo:           true,
		openai.GPT3Dot5TurboInstruct:   true,
	}

	DeepseekLocalModels = map[string]bool{
		LLAVA:                         true,
		deepseek.AzureDeepSeekR1:      true,
		deepseek.OpenRouterDeepSeekR1: true,
		deepseek.OpenRouterDeepSeekR1DistillLlama70B: true,
		deepseek.OpenRouterDeepSeekR1DistillLlama8B:  true,
		deepseek.OpenRouterDeepSeekR1DistillQwen14B:  true,
		deepseek.OpenRouterDeepSeekR1DistillQwen1_5B: true,
		deepseek.OpenRouterDeepSeekR1DistillQwen32B:  true,
	}
)

type MsgInfo struct {
	MsgId   int
	Content string
	SendLen int
}

type ImgResponse struct {
	Code    int              `json:"code"`
	Data    *ImgResponseData `json:"data"`
	Message string           `json:"message"`
	Status  string           `json:"status"`
}

type ImgResponseData struct {
	AlgorithmBaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_message"`
	} `json:"algorithm_base_resp"`
	ImageUrls        []string `json:"image_urls"`
	PeResult         string   `json:"pe_result"`
	PredictTagResult string   `json:"predict_tag_result"`
	RephraserResult  string   `json:"rephraser_result"`
}
