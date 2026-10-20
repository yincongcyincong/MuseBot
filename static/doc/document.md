# Document Recognition

Send a document (txt / pdf / csv / html) to the bot and ask questions about it. The document is parsed to plain text
and injected into the LLM context, so even text-only models such as DeepSeek can read it.

## Usage

Documents are handled automatically — there is no dedicated command:

| Action | Behavior |
|---|---|
| Send document **without** text | The bot analyzes it right away and replies with a summary |
| Send document **with** text (e.g. a caption) | The document is parsed and the question is answered directly |
| Any follow-up question | The document content stays in the conversation context, so you can keep asking about it like a normal chat |

The parsed document text is written into the conversation history (the same Q/A records the bot replays
for every request), so context survives within the session's history window.

Supported formats: `.txt`, `.pdf`, `.csv`, `.html` (also `.htm`). Any other file whose content is text (e.g. `.bat`,
`.diff`, `.log`, `.md`, `.json`, `Makefile`) is accepted and parsed as plain text by default; binary files (zip/exe/media)
are rejected with a hint. Extracted text is truncated to ~50 KB (head + tail) to fit the model context.

## Platform Support

| Platform             | Supported | Notes                                             |
|----------------------|:---------:|---------------------------------------------------|
| 🟦 Telegram          |     ✅     | Up to 20 MB per file (Bot API download limit)     |
| 🌈 Discord           |     ✅     | Attachments with `application/pdf` or `text/*`    |
| 🔷 Slack             |     ✅     | File shares with pdf/txt/csv/html mimetype        |
| 🟣 Lark (Feishu)     |     ✅     | File messages; file type detected by content      |
| 🆙 DingDing          |     ✅     | Files inside richText messages                    |
| 🌞 QQ                |     ✅     | Attachments with pdf/text content type            |
| 🚇 PersonalQQ        |     ✅     | OneBot `file` segment (needs a downloadable url)  |
| 🌛 Web API           |     ✅     | Raw body + `file_name` query param, see [doc](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api.md) |
| ⚡️ Work WeChat       |     ❌     | Callback model has no file message type           |
| 🐶 Wechat            |     ❌     | Callback model has no file message type           |

## Local Media Storage

Since this feature landed, images, videos and documents received or generated in a conversation are stored as **local
files** under `data/media/{platform}_{chatId}/` (one folder per IM conversation). The database `records` table stores
only the relative file path in `answer`/`content`, instead of inlined base64. Reading old records stays compatible:
rows written before the change (base64 data URIs or URLs) are still decoded.

### Parameter List

| Parameter Name         | Type     | Required/Optional | Default Value | Description                                            |
|------------------------|----------|-------------------|---------------|--------------------------------------------------------|
| `MEDIA_RETENTION_DAYS` | `int`    | Optional          | `30`          | Delete media files older than N days. `0` disables cleanup. A daily job removes expired files and emptied conversation folders. The db records are kept; referencing a cleaned file reports an error. |

> Tip (Docker): media files live in `data/` next to the binary. Mount a volume (e.g. `-v musebot-data:/app/data`) so
> files survive container restarts.
