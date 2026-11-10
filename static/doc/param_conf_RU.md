# Руководство по параметрам `MuseBot`

В этом документе подробно описаны различные параметры конфигурации для запуска `MuseBot`, что позволяет гибко развертывать и использовать бота в соответствии с вашими потребностями.

## Параметры конфигурации (`conf param`)

`MuseBot` настраивается через параметры командной строки. Ниже приведены примеры использования параметров для различных сценариев:

### 1. Базовая конфигурация (`basic`)

Минимально необходимые параметры для подключения бота к Telegram и API DeepSeek.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx
```

* `-telegram_bot_token`: Ваш токен Telegram Bot API
* `-deepseek_token`: Ваш токен DeepSeek API

### 2. Поддержка MySQL (`mysql`)

Для сохранения истории чатов и данных пользователей можно использовать базу данных MySQL.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-db_type=mysql \
-db_conf='root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local'
```

* `-db_type`: Тип БД (указать `mysql`)
* `-db_conf`: Строка подключения к MySQL (замените на свои данные)

### 3. Настройка прокси (`proxy`)

Используйте, если требуется доступ к API Telegram/DeepSeek через прокси.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-robot_proxy=http://127.0.0.1:7890 \
-llm_proxy=http://127.0.0.1:7890
```

* `-robot_proxy`: Адрес прокси для запросов к Telegram API
* `-llm_proxy`: Адрес прокси для запросов к DeepSeek API

### 4. Поддержка моделей OpenAI (`openai`)

Бот поддерживает работу с моделями OpenAI.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-type=openai \
-openai_token=sk-xxxx
```

* `-type`: Тип модели (указать `openai`)
* `-openai_token`: Ваш токен OpenAI API

### 5. Поддержка моделей Gemini (`gemini`)

Бот поддерживает работу с моделями Google Gemini.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-type=gemini \
-gemini_token=xxxxx
```

* `-type`: Тип модели (указать `gemini`)
* `-gemini_token`: Ваш токен Gemini API

### 6. Распознавание изображений (`identify photo`)

Для интеграции с сервисом распознавания изображений VolcEngine.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-volc_ak=xxx \
-volc_sk=xxx
```

* `-volc_ak`: VolcEngine Access Key
* `-volc_sk`: VolcEngine Secret Key

Подробнее: [Документация VolcEngine](https://www.volcengine.com/docs/6790/116987)

### 7. Распознавание голоса (`identify voice`)

Для интеграции с сервисом распознавания речи VolcEngine.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-vol_audio_app_id=xxx \
-vol_audio_rec_cluster=volcengine_input_common \
-vol_audio_token=xxxx
```

* `-vol_audio_app_id`: ID приложения распознавания речи
* `-vol_audio_rec_cluster`: Имя кластера (обычно `volcengine_input_common`)
* `-vol_audio_token`: Токен доступа

Подробнее: [Документация VolcEngine](https://www.volcengine.com/docs/6561/80816)

### 8. Инструменты MCP (`mcp`)

Для использования инструментов Amap (например, геолокации).

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-use_tools=true
```

* `-use_tools`: Активирует инструменты (по умолчанию `false`)

### 9. RAG с Milvus + эмбеддингами OpenAI (`rag milvus`)

Для использования RAG с Milvus и сервисом эмбеддингов OpenAI.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-openai_token=sk-xxxx \
-embedding_type=openai \
-vector_db_type=milvus
```

* `-openai_token`: Токен OpenAI (для эмбеддингов)
* `-embedding_type`: Тип эмбеддингов (указать `openai`)
* `-vector_db_type`: Тип векторной БД (указать `milvus`; также поддерживаются `weaviate`, `qdrant`)

### 10. RAG с Milvus (`rag milvus`)

Для использования RAG с Milvus и сервисом эмбеддингов Gemini.

```bash
./MuseBot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-gemini_token=xxx \
-embedding_type=gemini \
-vector_db_type=milvus
```

* `-gemini_token`: Токен Gemini (для эмбеддингов)
* `-embedding_type`: Тип эмбеддингов (указать `gemini`)
* `-vector_db_type`: Тип векторной БД (указать `milvus`)

### 11. RAG с Weaviate (`rag weaviate`)

Для использования RAG с Weaviate и сервисом эмбеддингов Ernie.

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

* `-ernie_ak`: Ernie Access Key
* `-ernie_sk`: Ernie Secret Key
* `-embedding_type`: Тип эмбеддингов (указать `ernie`)
* `-vector_db_type`: Тип векторной БД (указать `weaviate`)
* `-weaviate_url`: URL базы данных Weaviate