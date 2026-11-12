# 图片配置

图片**识别**（发图提问）与图片**生成**（`/photo`、`/edit_photo`）是两个独立能力，分别配置。

## 图片生成

厂商由用户用 `/photo_type` 命令切换；启动时的默认值取决于已配置的 token——任何配置了 token 的厂商（`GEMINI_TOKEN`、`OPENAI_TOKEN`、`ALIYUN_TOKEN`、`VOL_TOKEN`）都可用，`MEDIA_TYPE` 可固定启动默认值。

模型由用户用 `/img_model` 切换，或在启动时用下面的环境变量指定。

| 厂商     | `-photo_type` 取值 | 模型 |
|----------|--------------------|------|
| Gemini   | `gemini`  | `gemini-3-pro-image`, `gemini-nano-banana-2.1`, `gemini-3.1-flash-image`, `gemini-3.1-flash-lite-image`, `gemini-2.5-flash-image` |
| OpenAI   | `openai`  | `gpt-image-1.5`, `gpt-image-1` |
| 阿里云   | `aliyun`  | `qwen-image-plus`、`qwen-image-edit` |
| 火山引擎 | `vol`     | `doubao-seed-1-6-250615` |

> 使用阿里云时，`/edit_photo`（或带图的 `/photo`）在未自定义图片模型的情况下会使用 `qwen-image-edit`。

## 图片识别

识别使用所配置厂商的视觉模型能力。用 `/rec_type` 切换厂商，`/rec_model` 切换模型。

| 厂商     | 模型 |
|----------|------|
| Gemini   | `gemini-4-carbon`、`gemini-4-argon`、`gemini-3.8-flash`、`gemini-3.6-flash`、`gemini-3.5-flash-lite`、`gemini-2.5-pro`、`gemini-2.5-flash` |
| 火山引擎 | `doubao-seed-1-6-250615` |
| 阿里云   | `qwen-vl-max`、`qwen-vl-max-latest`、`qvq-max`、`qvq-max-latest`、`qwen-audio-turbo-latest` |

## 环境变量

| 参数名称             | 类型     | 可选/必选 | 说明 |
|----------------------|----------|-----------|------|
| `GEMINI_IMAGE_MODEL` | `String` | 可选      | gemini 图片生成模型 |
| `GEMINI_REC_MODEL`   | `String` | 可选      | gemini 图片识别模型 |
| `OPENAI_IMAGE_MODEL` | `String` | 可选      | openai 图片生成模型 |
| `OPENAI_REC_MODEL`   | `String` | 可选      | openai 图片识别模型 |
| `OPENAI_IMAGE_SIZE`  | `String` | 可选      | openai 图片大小 |
| `OPENAI_IMAGE_STYLE` | `String` | 可选      | openai 图片风格 |
| `VOL_IMAGE_MODEL`    | `String` | 可选      | 火山/豆包 图片生成模型 |
| `VOL_REC_MODEL`      | `String` | 可选      | 火山/豆包 图片识别模型 |
| `ALIYUN_IMAGE_MODEL` | `String` | 可选      | 阿里云 图片生成模型 |
| `ALIYUN_REC_MODEL`   | `String` | 可选      | 阿里云 图片识别模型 |
| `PHOTO_LANGUAGE`     | `int`    | 可选      | 火山水印语言：`0` 中文（`AI生成`）、`1` 英文（默认 `1`） |

火山引擎图片生成 API 的完整参数（seed、水印、分辨率等）见[图片参数](./photoconf_ZH.md)与[火山引擎文档](https://www.volcengine.com/docs/6791/1366783)。
