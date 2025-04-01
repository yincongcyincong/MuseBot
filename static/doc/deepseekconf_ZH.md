### 参数介绍

| Parameter Name      | Type     | Required/Optional | Description                                |
|---------------------|----------|-------------------|--------------------------------------------|
| `FREQUENCY_PENALTY` | `float`  | Optional          | 抑制重复词（值越高，重复越少）                            |
| `MAX_TOKENS`        | `int`    | Optional          | 单次生成的最大 token 数（受模型上下文窗口限制）                |
| `PRESENCE_PENALTY`  | `float`  | Optional          | 抑制已出现过的词（类似但不同于 frequency_penalty）         |
| `TEMPERATURE`       | `float`  | Optional          | 控制随机性（0=确定性高，2=随机性强）                       |
| `TOP_P`             | `int`    | Optional          | Nucleus Sampling，控制候选词范围                   |
| `STOP`              | `string` | Optional          | 遇到指定字符串时停止生成（如 ["\n", "。"]）                |
| `LOG_PROBS`         | `bool`   | Optional          | 控制是否返回模型生成 token 的 对数概率（log probabilities） |
| `TOP_LOG_PROBS`     | `int`    | Optional          | 显示模型在每个生成步骤中最可能的前N个候选词及其对数概率               |
