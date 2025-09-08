### 参数列表

| 参数名                     | 类型       | 必填/可选 | 描述                                                                                                                                                                     |
| ----------------------- | -------- | ----- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VOL_AUDIO_APP_ID`      | `string` | 可选    | 火山引擎音频应用 ID                                                                                                                                                            |
| `VOL_AUDIO_TOKEN`       | `string` | 可选    | 火山引擎音频访问 Token                                                                                                                                                         |
| `VOL_AUDIO_REC_CLUSTER` | `string` | 可选    | 火山引擎语音识别集群（默认值：`volcengine_input_common`） [语音模型说明](https://www.volcengine.com/docs/6561/80816)                                                                         |
| `VOL_AUDIO_VOICE_TYPE`  | `string` | 可选    | 火山引擎语音类型                                                                                                                                                               |
| `VOL_AUDIO_TTS_CLUSTER` | `string` | 可选    | 火山引擎 TTS（语音合成）集群（默认值：`volcano_tts`），可选项包括： [volcano\_tts](https://www.volcengine.com/docs/6561/1257584) / [volcano\_icl](https://www.volcengine.com/docs/6561/1305191) |
| `GEMINI_AUDIO_MODEL`    | `string` | 可选    | Gemini 音频模型（默认值：`gemini-2.5-flash-preview-tts`）                                                                                                                        |
| `GEMINI_VOICE_NAME`     | `string` | 可选    | Gemini 声音名称（默认值：`Kore`）                                                                                                                                                |
| `TTS_TYPE`              | `string` | 可选    | TTS 类型：`vol`（火山引擎）或 `gemini`（Gemini）                                                                                                                                   |

进入火山引擎控制台：
![image](https://github.com/user-attachments/assets/6261ee3c-2632-427d-a95e-85e55d85d971)

