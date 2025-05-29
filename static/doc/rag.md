### Parameter List

| Parameter Name   | Type     | Required/Optional | Description                              |
|------------------|----------|-------------------|------------------------------------------|
| `EMBEDDING_TYPE` | `String` | Required          | embedding split api: openai gemini ernie |
| `KNOWLEDGE_PATH` | `String` | Required          | knowledge doc path                       |
| `VECTOR_DB_TYPE` | `String` | Optional          | vector db type: chroma                   |
| `CHROMA_URL`     | `String` | Optional          | chroma url                               |
| `CHROMA_SPACE`   | `String` | Optional          | chroma space name                        |
| `CHUNK_SIZE`     | `String` | Optional          | rag file chunk size                      |
| `CHUNK_OVERLAP`  | `String` | Optional          | rag file chunk overlap                   |
