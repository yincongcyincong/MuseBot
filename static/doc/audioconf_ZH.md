### 参数列表

| 参数名                      | 类型       | 必填 | 默认值                            | 说明                                                                                                                                                       |
|--------------------------|----------|----|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `VOL_AUDIO_APP_ID`       | `string` | 否  | —                              | 火山引擎音频应用的 **App ID**，可在 [Volcengine 控制台](https://console.volcengine.com/) 获取。                                                                            |
| `VOL_AUDIO_TOKEN`        | `string` | 否  | —                              | 火山引擎音频的 **访问令牌（Access Token）**。                                                                                                                          |
| `VOL_AUDIO_REC_CLUSTER`  | `string` | 否  | `volcengine_input_common`      | 火山引擎语音识别集群。可参考 [语音识别模型文档](https://www.volcengine.com/docs/6561/80816)。                                                                                   |
| `VOL_AUDIO_VOICE_TYPE`   | `string` | 否  | —                              | 火山引擎语音合成的 **音色类型**。                                                                                                                                      |
| `VOL_AUDIO_TTS_CLUSTER`  | `string` | 否  | `volcano_tts`                  | 火山引擎 **语音合成集群**。可选值包括：<br>• [volcano_tts](https://www.volcengine.com/docs/6561/1257584)<br>• [volcano_icl](https://www.volcengine.com/docs/6561/1305191) |
| `VOL_END_SMOOTH_WINDOW`  | `int`    | 否  | `1500`                         | 火山引擎音频播放的尾音平滑窗口（单位：毫秒）。                                                                                                                                  |
| `VOL_TTS_SPEAKER`        | `string` | 否  | `zh_female_vv_jupiter_bigtts`  | 火山引擎默认的 **发音人（Speaker）**。                                                                                                                                |
| `VOL_BOT_NAME`           | `string` | 否  | `豆包`                           | 使用火山引擎语音的默认 **机器人名称**。                                                                                                                                   |
| `VOL_SYSTEM_ROLE`        | `string` | 否  | `你使用活泼灵动的女声，性格开朗，热爱生活。`        | 定义语音助手的 **角色设定**。                                                                                                                                        |
| `VOL_SPEAKING_STYLE`     | `string` | 否  | `你的说话风格简洁明了，语速适中，语调自然。`        | 定义语音助手的 **说话风格**。                                                                                                                                        |
| `GEMINI_AUDIO_MODEL`     | `string` | 否  | `gemini-2.5-flash-preview-tts` | Gemini 使用的 **语音合成模型**。                                                                                                                                   |
| `GEMINI_VOICE_NAME`      | `string` | 否  | `Kore`                         | Gemini 使用的 **音色名称（Voice Name）**。                                                                                                                         |
| `OPENAI_AUDIO_MODEL`     | `string` | 否  | `tts-1`                        | OpenAI 的 **语音合成模型**。                                                                                                                                     |
| `OPENAI_VOICE_NAME`      | `string` | 否  | `alloy`                        | OpenAI 的 **音色名称**。                                                                                                                                       |
| `ALIYUN_AUDIO_MODEL`     | `string` | 否  | `qwen3-tts-flash`              | 阿里云（通义千问）使用的 **语音合成模型**。                                                                                                                                 |
| `ALIYUN_AUDIO_VOICE`     | `string` | 否  | `Cherry`                       | 阿里云语音合成的 **发音人名称**。                                                                                                                                      |
| `ALIYUN_AUDIO_REC_MODEL` | `string` | 否  | `qwen-audio-turbo-latest`      | 阿里云使用的 **语音识别模型**。                                                                                                                                       |
| `TTS_TYPE`               | `string` | 否  | —                              | 指定使用的 **TTS 服务类型**，可选值：`vol` / `gemini` / `openai` / `aliyun`。                                                                                           |

进入火山引擎控制台：
![image](https://github.com/user-attachments/assets/6261ee3c-2632-427d-a95e-85e55d85d971)

