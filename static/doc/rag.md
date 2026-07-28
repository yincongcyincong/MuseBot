### Parameter List

| Parameter Name    | Type     | Required/Optional | Description                              |
|-------------------|----------|-------------------|------------------------------------------|
| `EMBEDDING_TYPE`  | `String` | Required          | embedding split api: openai gemini ernie |
| `KNOWLEDGE_PATH`  | `String` | Required          | knowledge doc path                       |
| `VECTOR_DB_TYPE`  | `String` | Required          | vector db type: weaviate, milvus, qdrant |
| `CHROMA_URL`      | `String` | Optional          | chroma url:http://localhost:8080         |
| `MILVUS_URL`      | `String` | Optional          | milvus url: http://localhost:19530       |
| `WEAVIATE_URL`    | `String` | Optional          | weaviate url: localhost:8000             |
| `WEAVIATE_SCHEME` | `String` | Optional          | weaviate scheme: http                    |
| `QDRANT_HOST`     | `String` | Optional          | qdrant gRPC host: localhost              |
| `QDRANT_PORT`     | `Integer`| Optional          | qdrant gRPC port: 6334                   |
| `QDRANT_API_KEY`  | `String` | Optional          | qdrant API key                           |
| `QDRANT_USE_TLS`  | `Boolean`| Optional          | use TLS for qdrant gRPC: false           |
| `SPACE`           | `String` | Optional          | collection or vector db space name       |
| `CHUNK_SIZE`      | `String` | Optional          | rag file chunk size                      |
| `CHUNK_OVERLAP`   | `String` | Optional          | rag file chunk overlap                   |

### Qdrant

Set `VECTOR_DB_TYPE=qdrant` and choose the collection with `SPACE`. Qdrant defaults to `localhost:6334`. Set `QDRANT_API_KEY` if the server requires authentication and `QDRANT_USE_TLS=true` for TLS.

MuseBot creates the collection with cosine distance if it does not exist. It sends one embedding request to determine the vector size. MuseBot does not modify existing collections, whose vector size must match the selected embedding model.
