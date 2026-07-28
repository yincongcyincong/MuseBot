### 参数列表

| 参数名称             | 类型    | 是否必填 | 描述                           |
|------------------|-------|------|------------------------------|
| `EMBEDDING_TYPE` | `字符串` | 必填   | 向量化方式，支持：openai、gemini、ernie |
| `KNOWLEDGE_PATH` | `字符串` | 必填   | 知识文档路径                       |
| `VECTOR_DB_TYPE` | `字符串` | 可选   | 向量数据库类型：milvus、weaviate、qdrant |
| `CHROMA_URL`     | `字符串` | 可选   | Chroma 数据库的连接地址              |
| `QDRANT_HOST`    | `字符串` | 可选   | Qdrant gRPC 主机，默认：localhost     |
| `QDRANT_PORT`    | `整数`  | 可选   | Qdrant gRPC 端口，默认：6334          |
| `QDRANT_API_KEY` | `字符串` | 可选   | Qdrant API 密钥                       |
| `QDRANT_USE_TLS` | `布尔值` | 可选   | 是否为 Qdrant gRPC 启用 TLS，默认：false |
| `SPACE`          | `字符串` | 可选   | 向量数据库的集合或命名空间             |
| `CHUNK_SIZE`     | `字符串` | 可选   | RAG 文件的切片大小                  |
| `CHUNK_OVERLAP`  | `字符串` | 可选   | RAG 文件的切片重叠大小                |

使用 Qdrant 时，将 `VECTOR_DB_TYPE` 设置为 `qdrant`。Qdrant 默认使用 gRPC 地址 `localhost:6334`。如果 `SPACE` 指定的集合不存在，MuseBot 会使用余弦距离创建集合，并发送一次嵌入请求以确定向量维度。MuseBot 不会修改已有集合；其向量维度必须与所选嵌入模型一致。
