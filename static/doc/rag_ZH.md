### 参数列表

| 参数名称             | 类型    | 是否必填 | 描述                           |
|------------------|-------|------|------------------------------|
| `EMBEDDING_TYPE` | `字符串` | 必填   | 向量化方式，支持：openai、gemini、ernie |
| `KNOWLEDGE_PATH` | `字符串` | 必填   | 知识文档路径                       |
| `VECTOR_DB_TYPE` | `字符串` | 可选   | 向量数据库类型，例如：chroma            |
| `CHROMA_URL`     | `字符串` | 可选   | Chroma 数据库的连接地址              |
| `CHROMA_SPACE`   | `字符串` | 可选   | Chroma 中的命名空间（space name）    |
| `CHUNK_SIZE`     | `字符串` | 可选   | RAG 文件的切片大小                  |
| `CHUNK_OVERLAP`  | `字符串` | 可选   | RAG 文件的切片重叠大小                |

