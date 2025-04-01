### Parameter List

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

