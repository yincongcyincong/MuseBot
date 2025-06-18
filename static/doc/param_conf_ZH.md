## `telegram-deepseek-bot` 参数使用指南

本文档详细介绍了 `telegram-deepseek-bot` 运行时的各种配置参数，帮助用户根据需求灵活部署和使用。

### 配置参数 (`conf param`)

`telegram-deepseek-bot` 通过命令行参数进行配置。以下是不同场景下的参数使用示例：

#### 1\. 基础配置 (`basic`)

这是运行机器人所需的最基本参数，用于连接 Telegram 和 DeepSeek API。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx
```

* `-telegram_bot_token`: 您的 Telegram Bot API Token。
* `-deepseek_token`: 您的 DeepSeek API Token。

#### 2\. MySQL 数据库支持 (`mysql`)

如果需要持久化聊天记录或用户数据，可以使用 MySQL 数据库。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-db_type=mysql \
-db_conf='root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local'
```

* `-db_type`: 数据库类型，设置为 `mysql`。
* `-db_conf`: MySQL 连接字符串。请根据您的数据库配置替换 `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`。

#### 3\. 代理配置 (`proxy`)

当您的网络环境需要通过代理访问 Telegram 或 DeepSeek API 时，可以使用此配置。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-telegram_proxy=http://127.0.0.1:7890 \
-deepseek_proxy=http://127.0.0.1:7890
```

* `-telegram_proxy`: Telegram API 请求使用的代理地址。
* `-deepseek_proxy`: DeepSeek API 请求使用的代理地址。

#### 4\. OpenAI 模型支持 (`openai`)

除了 DeepSeek，机器人也支持使用 OpenAI 模型。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=openai \
-openai_token=sk-xxxx
```

* `-type`: 模型类型，设置为 `openai`。
* `-openai_token`: 您的 OpenAI API Token。

#### 5\. Gemini 模型支持 (`gemini`)

机器人还支持使用 Google Gemini 模型。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=gemini \
-gemini_token=xxxxx
```

* `-type`: 模型类型，设置为 `gemini`。
* `-gemini_token`: 您的 Gemini API Token。

#### 6\. OpenRouter 模型支持 (`openrouter`)

集成 OpenRouter 平台，可以使用其提供的多种模型。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=openrouter \
-openrouter_token=sk-or-v1-xxxx
```

* `-type`: 模型类型，设置为 `openrouter`。
* `-openrouter_token`: 您的 OpenRouter API Token。

#### 7\. 图片识别 (`identify photo`)

集成火山引擎（VolcEngine）的图片识别功能，需要提供火山引擎的 AK/SK。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-volc_ak=xxx \
-volc_sk=xxx
```

* `-volc_ak`: 火山引擎 Access Key。
* `-volc_sk`: 火山引擎 Secret Key。

更多详情请参考：[火山引擎图片识别文档](https://www.volcengine.com/docs/6790/116987)

#### 8\. 语音识别 (`identify voice`)

集成火山引擎（VolcEngine）的语音识别功能。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-audio_app_id=xxx \
-audio_cluster=volcengine_input_common \
-audio_token=xxxx
```

* `-audio_app_id`: 火山引擎语音识别的 App ID。
* `-audio_cluster`: 语音识别集群名称，通常为 `volcengine_input_common`。
* `-audio_token`: 语音识别的 Token。

更多详情请参考：[火山引擎语音识别文档](https://www.volcengine.com/docs/6561/80816)

#### 9\. 高德地图 MCP (`amap mcp`)

如果您的机器人需要使用高德地图的相关工具，例如地理位置查询等。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-use_tools=true
```

* `-use_tools`: 启用工具使用功能，设置为 `true`，默认为`false`。

#### 10\. RAG (Retrieval Augmented Generation) - ChromaDB (`rag chroma`)

结合 ChromaDB 进行 RAG，需要使用 OpenAI 的 Embedding 服务。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-openai_token=sk-xxxx \
-embedding_type=openai \
-vector_db_type=chroma
```

* `-openai_token`: 您的 OpenAI API Token (用于 embedding)。
* `-embedding_type`: Embedding 类型，设置为 `openai`。
* `-vector_db_type`: 向量数据库类型，设置为 `chroma`。

#### 11\. RAG (Retrieval Augmented Generation) - Milvus (`rag milvus`)

结合 Milvus 进行 RAG，需要使用 Gemini 的 Embedding 服务。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-gemini_token=xxx \
-embedding_type=gemini \
-vector_db_type=milvus
```

* `-gemini_token`: 您的 Gemini API Token (用于 embedding)。
* `-embedding_type`: Embedding 类型，设置为 `gemini`。
* `-vector_db_type`: 向量数据库类型，设置为 `milvus`。

#### 12\. RAG (Retrieval Augmented Generation) - Weaviate (`rag weaviate`)

结合 Weaviate 进行 RAG，需要使用 Ernie 的 Embedding 服务。

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-ernie_ak=xxx \
-ernie_sk=xxx \
-embedding_type=ernie \
-vector_db_type=weaviate \
-weaviate_url=127.0.0.1:8080
```

* `-ernie_ak`: 您的 Ernie Access Key (用于 embedding)。
* `-ernie_sk`: 您的 Ernie Secret Key (用于 embedding)。
* `-embedding_type`: Embedding 类型，设置为 `ernie`。
* `-vector_db_type`: 向量数据库类型，设置为 `weaviate`。
* `-weaviate_url`: Weaviate 数据库的 URL。

