### Parameter List

| Parameter Name          | Type     | Required/Optional | Description                                                                                                                                                                           |
|-------------------------|----------|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `VOL_AUDIO_APP_ID`      | `string` | Optional          | Volcengine audio app id                                                                                                                                                               |
| `VOL_AUDIO_TOKEN`       | `string` | Optional          | Volcengine audio access token                                                                                                                                                         |
| `VOL_AUDIO_REC_CLUSTER` | `string` | Optional          | Volcengine audio recognition cluster (default: `volcengine_input_common`) [speech model](https://www.volcengine.com/docs/6561/80816)                                                  |
| `VOL_AUDIO_VOICE_TYPE`  | `string` | Optional          | Volcengine audio voice type                                                                                                                                                           |
| `VOL_AUDIO_TTS_CLUSTER` | `string` | Optional          | Volcengine TTS cluster (default: `volcano_tts`) , include:  [volcano_tts](https://www.volcengine.com/docs/6561/1257584) / [volcano_icl](https://www.volcengine.com/docs/6561/1305191) |
| `GEMINI_AUDIO_MODEL`    | `string` | Optional          | Gemini audio model (default: `gemini-2.5-flash-preview-tts`)                                                                                                                          |
| `GEMINI_VOICE_NAME`     | `string` | Optional          | Gemini voice name (default: `Kore`)                                                                                                                                                   |
| `TTS_TYPE`              | `string` | Optional          | TTS type: `vol` (Volcengine) or `gemini` (Gemini)                                                                                                                                     |

enter vol engine console.
![image](https://github.com/user-attachments/assets/6261ee3c-2632-427d-a95e-85e55d85d971)