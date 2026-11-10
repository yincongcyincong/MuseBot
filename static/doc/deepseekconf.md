### LLM Sampling Parameters

These sampling parameters apply to **all OpenAI-compatible chat providers** (deepseek, openai, gemini, vol, aliyun, ollama, chatanywhere), not only DeepSeek.

> They take effect only when `LLM_OPTION_PARAM` is enabled (`-llm_option_param=true` or env `LLM_OPTION_PARAM=true`, default off). With the switch off, requests are sent without these fields and the provider defaults apply.

| Parameter Name      | Type     | Required/Optional | Description                                      |

| Parameter Name      | Type     | Required/Optional | Description                                      |
|---------------------|----------|-------------------|--------------------------------------------------|
| `FREQUENCY_PENALTY` | `float`  | Optional          | Reduces repetition (higher values decrease repetition). |
| `MAX_TOKENS`        | `int`    | Optional          | Maximum number of tokens per generation (limited by model context window). |
| `PRESENCE_PENALTY`  | `float`  | Optional          | Discourages already-used words (similar to but different from frequency_penalty). |
| `TEMPERATURE`       | `float`  | Optional          | Controls randomness (0 = deterministic, 2 = highly random). |
| `TOP_P`             | `int`    | Optional          | Nucleus Sampling, controls the range of candidate words. |
| `STOP`              | `string` | Optional          | Stops generation upon encountering specified strings (e.g., ["\n", "."]). |
| `LOG_PROBS`         | `bool`   | Optional          | Determines whether to return log probabilities of generated tokens. |
| `TOP_LOG_PROBS`     | `int`    | Optional          | Displays the top N most likely words and their log probabilities at each step. |

