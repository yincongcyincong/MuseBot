### Parameter List

| Parameter Name   | Type     | Required/Optional | Description                              |
|------------------|----------|-------------------|------------------------------------------|
| `EMBEDDING_TYPE` | `String` | Required          | embedding split api: openai gemini ernie |
| `KNOWLEDGE_PATH` | `String` | Required          | knowledge doc path                       |
| `VECTOR_DB_TYPE` | `String` | Optional          | vector db type: chroma  weaviate milvus  |
| `CHROMA_URL`     | `String` | Optional          | chroma url                               |
| `SPACE`          | `String` | Optional          | vector db space name                     |
| `CHUNK_SIZE`     | `String` | Optional          | rag file chunk size                      |
| `CHUNK_OVERLAP`  | `String` | Optional          | rag file chunk overlap                   |

