### Список параметров

| Имя параметра           | Тип      | Обязательный/Необязательный | Описание                                                                                                                                                                                                             |
|-------------------------|----------|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `VOL_AUDIO_APP_ID`      | `string` | Необязательный              | Идентификатор аудиоприложения Volcengine                                                                                                                                                                             |
| `VOL_AUDIO_TOKEN`       | `string` | Необязательный              | Токен доступа к аудио Volcengine                                                                                                                                                                                     |
| `VOL_AUDIO_REC_CLUSTER` | `string` | Необязательный              | Кластер распознавания речи Volcengine (по умолчанию: `volcengine_input_common`) [модель распознавания речи](https://www.volcengine.com/docs/6561/80816)                                                              |
| `VOL_AUDIO_VOICE_TYPE`  | `string` | Необязательный              | Тип голоса Volcengine                                                                                                                                                                                                |
| `VOL_AUDIO_TTS_CLUSTER` | `string` | Необязательный              | Кластер TTS (синтеза речи) Volcengine (по умолчанию: `volcano_tts`). Доступные варианты: [volcano\_tts](https://www.volcengine.com/docs/6561/1257584) / [volcano\_icl](https://www.volcengine.com/docs/6561/1305191) |
| `GEMINI_AUDIO_MODEL`    | `string` | Необязательный              | Аудиомодель Gemini (по умолчанию: `gemini-2.5-flash-preview-tts`)                                                                                                                                                    |
| `GEMINI_VOICE_NAME`     | `string` | Необязательный              | Имя голоса Gemini (по умолчанию: `Kore`)                                                                                                                                                                             |
| `OPENAI_AUDIO_MODEL`    | `string` | Необязательно               | Аудио-модель OpenAI (по умолчанию: `tts-1`)                                                                                                                                                                          |
| `OPENAI_VOICE_NAME`     | `string` | Необязательно               | Имя голоса OpenAI (по умолчанию: `alloy`)                                                                                                                                                                            |
| `TTS_TYPE`              | `string` | Необязательный              | Тип TTS: vol/gemini/openai                                                                                                                                                                                           |

Перейдите в консоль Volcengine:
![image](https://github.com/user-attachments/assets/6261ee3c-2632-427d-a95e-85e55d85d971)

