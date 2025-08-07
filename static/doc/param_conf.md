## `MuseBot` Parameter Usage Guide

This document provides a detailed overview of the various configuration parameters for running `MuseBot`, helping users deploy and utilize it flexibly according to their needs.

### Configuration Parameters (`conf param`)

`MuseBot` is configured via command-line parameters. Below are examples of parameter usage for different scenarios:

#### 1\. Basic Configuration (`basic`)

These are the most essential parameters required to run the bot, connecting it to Telegram and the DeepSeek API.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx
```

* `-telegram_bot_token`: Your Telegram Bot API Token.
* `-deepseek_token`: Your DeepSeek API Token.

#### 2\. MySQL Database Support (`mysql`)

If you need to persist chat history or user data, you can use a MySQL database.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-db_type=mysql \
-db_conf='root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local'
```

* `-db_type`: Database type, set to `mysql`.
* `-db_conf`: MySQL connection string. Please replace `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` with your database configuration.

#### 3\. Proxy Configuration (`proxy`)

Use this configuration if your network environment requires accessing Telegram or DeepSeek API through a proxy.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-telegram_proxy=http://127.0.0.1:7890 \
-deepseek_proxy=http://127.0.0.1:7890
```

* `-telegram_proxy`: The proxy address used for Telegram API requests.
* `-deepseek_proxy`: The proxy address used for DeepSeek API requests.

#### 4\. OpenAI Model Support (`openai`)

In addition to DeepSeek, the bot also supports using OpenAI models.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-type=openai \
-openai_token=sk-xxxx
```

* `-type`: Model type, set to `openai`.
* `-openai_token`: Your OpenAI API Token.

#### 5\. Gemini Model Support (`gemini`)

The bot also supports using Google Gemini models.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-type=gemini \
-gemini_token=xxxxx
```

* `-type`: Model type, set to `gemini`.
* `-gemini_token`: Your Gemini API Token.

#### 6\. OpenRouter Model Support (`openrouter`)

Integrate with the OpenRouter platform to use various models it provides.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-type=openrouter \
-openrouter_token=sk-or-v1-xxxx
```

* `-type`: Model type, set to `openrouter`.
* `-openrouter_token`: Your OpenRouter API Token.

#### 7\. Photo Identification (`identify photo`)

To integrate with VolcEngine's photo identification feature, you'll need to provide your VolcEngine AK/SK.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-volc_ak=xxx \
-volc_sk=xxx
```

* `-volc_ak`: VolcEngine Access Key.
* `-volc_sk`: VolcEngine Secret Key.

For more details, please refer to: [VolcEngine Image Recognition Documentation](https://www.volcengine.com/docs/6790/116987)

#### 8\. Voice Identification (`identify voice`)

To integrate with VolcEngine's voice recognition feature.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-audio_app_id=xxx \
-audio_cluster=volcengine_input_common \
-audio_token=xxxx
```

* `-audio_app_id`: VolcEngine Voice Recognition App ID.
* `-audio_cluster`: Voice recognition cluster name, typically `volcengine_input_common`.
* `-audio_token`: Voice recognition Token.

For more details, please refer to: [VolcEngine Voice Recognition Documentation](https://www.volcengine.com/docs/6561/80816)

#### 9\. MCP (`mcp`)

If your bot needs to use Amap (Gaode Map) related tools, such as geolocation queries.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-use_tools=true
```
* `-use_tools`: Enables tool usage functionality, set to `true` default is `false`.

#### 10\. RAG (Retrieval Augmented Generation) - ChromaDB (`rag milvus`)

To perform RAG with ChromaDB, you'll need to use OpenAI's Embedding service.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-openai_token=sk-xxxx \
-embedding_type=openai \
-vector_db_type=milvus
```

* `-openai_token`: Your OpenAI API Token (for embedding).
* `-embedding_type`: Embedding type, set to `openai`.
* `-vector_db_type`: Vector database type, set to `chroma`.

#### 11\. RAG (Retrieval Augmented Generation) - Milvus (`rag milvus`)

To perform RAG with Milvus, you'll need to use Gemini's Embedding service.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-gemini_token=xxx \
-embedding_type=gemini \
-vector_db_type=milvus
```

* `-gemini_token`: Your Gemini API Token (for embedding).
* `-embedding_type`: Embedding type, set to `gemini`.
* `-vector_db_type`: Vector database type, set to `milvus`.

#### 12\. RAG (Retrieval Augmented Generation) - Weaviate (`rag weaviate`)

To perform RAG with Weaviate, you'll need to use Ernie's Embedding service.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-ernie_ak=xxx \
-ernie_sk=xxx \
-embedding_type=ernie \
-vector_db_type=weaviate \
-weaviate_url=127.0.0.1:8080
```

* `-ernie_ak`: Your Ernie Access Key (for embedding).
* `-ernie_sk`: Your Ernie Secret Key (for embedding).
* `-embedding_type`: Embedding type, set to `ernie`.
* `-vector_db_type`: Vector database type, set to `weaviate`.
* `-weaviate_url`: The URL of your Weaviate database.

