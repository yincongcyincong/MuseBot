# Конфигурация изображений

Две независимые возможности настраиваются отдельно: **распознавание изображений** (отправить фото и задать вопрос) и **генерация изображений** (`/photo`, `/edit_photo`).

## Генерация изображений

Провайдер выбирается пользователем командой `/photo_type`; значение по умолчанию при запуске определяется настроенными токенами — доступен любой провайдер с заданным токеном (`GEMINI_TOKEN`, `OPENAI_TOKEN`, `ALIYUN_TOKEN`, `VOL_TOKEN`), а `MEDIA_TYPE` фиксирует значение по умолчанию.

Модель переключается пользователем командой `/img_model` или задаётся при запуске переменными окружения ниже.

| Провайдер | Значение `-photo_type` | Модели |
|-----------|------------------------|--------|
| Gemini    | `gemini`  | `imagen-3.0-generate-002`, `imagen-4.0-generate-001`, `imagen-4.0-fast-generate-001`, `gemini-2.5-flash-image`, `gemini-3-pro-image-preview` |
| OpenAI    | `openai`  | `gpt-image-1` |
| Aliyun    | `aliyun`  | `qwen-image-plus`, `qwen-image-edit` |
| Vol       | `vol`     | `doubao-seed-1-6-250615` |

> С Aliyun команда `/edit_photo` (или `/photo` с прикреплённым изображением) использует `qwen-image-edit`, если пользовательская модель не задана.

## Распознавание изображений

Распознавание использует визуальные модели настроенного провайдера. Провайдер переключается командой `/rec_type`, модель — командой `/rec_model`.

| Провайдер | Модели |
|-----------|--------|
| Gemini    | `gemini-2.5-pro`, `gemini-2.5-flash` |
| Vol       | `doubao-seed-1-6-250615` |
| Aliyun    | `qwen-vl-max`, `qwen-vl-max-latest`, `qvq-max`, `qvq-max-latest`, `qwen-audio-turbo-latest` |

## Переменные окружения

| Параметр              | Тип      | Обязательный | Описание |
|-----------------------|----------|--------------|----------|
| `GEMINI_IMAGE_MODEL`  | `String` | Опциональный | модель генерации изображений gemini |
| `GEMINI_REC_MODEL`    | `String` | Опциональный | модель распознавания изображений gemini |
| `OPENAI_IMAGE_MODEL`  | `String` | Опциональный | модель генерации изображений openai |
| `OPENAI_REC_MODEL`    | `String` | Опциональный | модель распознавания изображений openai |
| `OPENAI_IMAGE_SIZE`   | `String` | Опциональный | размер изображений openai |
| `OPENAI_IMAGE_STYLE`  | `String` | Опциональный | стиль изображений openai |
| `VOL_IMAGE_MODEL`     | `String` | Опциональный | модель изображений vol/doubao |
| `VOL_REC_MODEL`       | `String` | Опциональный | модель распознавания vol/doubao |
| `ALIYUN_IMAGE_MODEL`  | `String` | Опциональный | модель изображений aliyun |
| `ALIYUN_REC_MODEL`    | `String` | Опциональный | модель распознавания aliyun |
| `PHOTO_LANGUAGE`      | `int`    | Опциональный | язык водяного знака vol: `0` китайский (`AI生成`), `1` английский (по умолчанию `1`) |

Полные параметры API генерации изображений Volcengine (seed, водяной знак, разрешение и т.д.) см. в [настройке изображений](./photoconf_RU.md) и [документации VolcEngine](https://www.volcengine.com/docs/6791/1366783).
