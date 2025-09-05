## 视频参数列表

| 参数名称                    | 类型       | 必填/可选 | 描述                                                                                                                                                                               |
|-------------------------|----------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `VOL_AUDIO_APP_ID`      | `string` | 可选    | Volcengine 音频应用的 AppID                                                                                                                                                           |
| `VOL_AUDIO_TOKEN`       | `string` | 可选    | Volcengine 音频访问 Token                                                                                                                                                            |
| `VOL_AUDIO_REC_CLUSTER` | `string` | 可选    | Volcengine 音频识别集群（默认值: `volcengine_input_common`），参考 [语音识别模型文档](https://www.volcengine.com/docs/6561/80816)                                                                      |
| `VOL_AUDIO_VOICE_TYPE`  | `string` | 可选    | Volcengine 音频语音类型                                                                                                                                                                |
| `VOL_AUDIO_TTS_CLUSTER` | `string` | 可选    | Volcengine TTS 集群（默认值: `volcano_tts`），可选值包括： [语音合成 volcano_tts](https://www.volcengine.com/docs/6561/1257584) / [语音复刻 volcano_icl](https://www.volcengine.com/docs/6561/1305191) |
| `GEMINI_AUDIO_MODEL`    | `string` | 可选    | Gemini 音频模型（默认值: `gemini-2.5-flash-preview-tts`）                                                                                                                                 |
| `GEMINI_VOICE_NAME`     | `string` | 可选    | Gemini 音色名称（默认值: `Kore`）                                                                                                                                                         |
| `TTS_TYPE`              | `string` | 可选    | TTS 类型：`vol` (Volcengine) 或 `gemini` (Gemini)                                                                                                                                    |
