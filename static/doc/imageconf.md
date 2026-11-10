# Image Configuration

Two independent capabilities are configured separately: **image recognition** (send a photo and ask about it) and **image generation** (`/photo`, `/edit_photo`).

## Image Generation

The provider is selected per user with the `/photo_type` command; the startup default comes from the token(s) you configured — any provider whose token is set (`GEMINI_TOKEN`, `OPENAI_TOKEN`, `ALIYUN_TOKEN`, `VOL_TOKEN`) is available, and `MEDIA_TYPE` pins the startup default.

The model is switched per user with `/img_model`, or set at startup with the environment variables below.

| Provider | `-photo_type` value | Models |
|----------|--------------------|--------|
| Gemini   | `gemini`  | `imagen-3.0-generate-002`, `imagen-4.0-generate-001`, `imagen-4.0-fast-generate-001`, `gemini-2.5-flash-image`, `gemini-3-pro-image-preview` |
| OpenAI   | `openai`  | `gpt-image-1` |
| Aliyun   | `aliyun`  | `qwen-image-plus`, `qwen-image-edit` |
| Vol      | `vol`     | `doubao-seed-1-6-250615` |

> With Aliyun, `/edit_photo` (or `/photo` with an attached image) uses `qwen-image-edit` unless a custom image model is set.

## Image Recognition

Recognition uses the vision capability of the configured chat/recognition provider. Switch the provider with `/rec_type` and the model with `/rec_model`.

| Provider | Models |
|----------|--------|
| Gemini   | `gemini-2.5-pro`, `gemini-2.5-flash` |
| Vol      | `doubao-seed-1-6-250615` |
| Aliyun   | `qwen-vl-max`, `qwen-vl-max-latest`, `qvq-max`, `qvq-max-latest`, `qwen-audio-turbo-latest` |

## Environment Variables

| Parameter Name       | Type     | Required/Optional | Description |
|----------------------|----------|-------------------|-------------|
| `GEMINI_IMAGE_MODEL` | `String` | Optional          | gemini create photo model |
| `GEMINI_REC_MODEL`   | `String` | Optional          | gemini recognize photo model |
| `OPENAI_IMAGE_MODEL` | `String` | Optional          | openai create photo model |
| `OPENAI_REC_MODEL`   | `String` | Optional          | openai recognize photo model |
| `OPENAI_IMAGE_SIZE`  | `String` | Optional          | openai image size |
| `OPENAI_IMAGE_STYLE` | `String` | Optional          | openai image style |
| `VOL_IMAGE_MODEL`    | `String` | Optional          | vol/doubao photo model |
| `VOL_REC_MODEL`      | `String` | Optional          | vol/doubao recognize model |
| `ALIYUN_IMAGE_MODEL` | `String` | Optional          | aliyun photo model |
| `ALIYUN_REC_MODEL`   | `String` | Optional          | aliyun recognize model |
| `PHOTO_LANGUAGE`     | `int`    | Optional          | Vol watermark language: `0` Chinese (`AI生成`), `1` English (default `1`) |

For the full Volcengine image generation API parameters (seed, watermark, resolution, etc.) see [photo config](./photoconf.md) and the [VolcEngine documentation](https://www.volcengine.com/docs/6791/1366783).
