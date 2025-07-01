# Руководство по параметрам `telegram-deepseek-bot`

В этом документе подробно описаны различные параметры конфигурации для запуска `telegram-deepseek-bot`, что позволяет гибко развертывать и использовать бота в соответствии с вашими потребностями.

## Параметры конфигурации (`conf param`)

`telegram-deepseek-bot` настраивается через параметры командной строки. Ниже приведены примеры использования параметров для различных сценариев:

### 1. Базовая конфигурация (`basic`)

Минимально необходимые параметры для подключения бота к Telegram и API DeepSeek.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx
```

* `-telegram_bot_token`: Ваш токен Telegram Bot API
* `-deepseek_token`: Ваш токен DeepSeek API

### 2. Поддержка MySQL (`mysql`)

Для сохранения истории чатов и данных пользователей можно использовать базу данных MySQL.

```bash
./telegram-deepseek-bot \
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
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-telegram_proxy=http://127.0.0.1:7890 \
-deepseek_proxy=http://127.0.0.1:7890
```

* `-telegram_proxy`: Адрес прокси для запросов к Telegram API
* `-deepseek_proxy`: Адрес прокси для запросов к DeepSeek API

### 4. Поддержка моделей OpenAI (`openai`)

Бот поддерживает работу с моделями OpenAI.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=openai \
-openai_token=sk-xxxx
```

* `-type`: Тип модели (указать `openai`)
* `-openai_token`: Ваш токен OpenAI API

### 5. Поддержка моделей Gemini (`gemini`)

Бот поддерживает работу с моделями Google Gemini.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=gemini \
-gemini_token=xxxxx
```

* `-type`: Тип модели (указать `gemini`)
* `-gemini_token`: Ваш токен Gemini API

### 6. Поддержка OpenRouter (`openrouter`)

Интеграция с платформой OpenRouter для доступа к различным моделям.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-type=openrouter \
-openrouter_token=sk-or-v1-xxxx
```

* `-type`: Тип модели (указать `openrouter`)
* `-openrouter_token`: Ваш токен OpenRouter API

### 7. Распознавание изображений (`identify photo`)

Для интеграции с сервисом распознавания изображений VolcEngine.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-volc_ak=xxx \
-volc_sk=xxx
```

* `-volc_ak`: VolcEngine Access Key
* `-volc_sk`: VolcEngine Secret Key

Подробнее: [Документация VolcEngine](https://www.volcengine.com/docs/6790/116987)

### 8. Распознавание голоса (`identify voice`)

Для интеграции с сервисом распознавания речи VolcEngine.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-audio_app_id=xxx \
-audio_cluster=volcengine_input_common \
-audio_token=xxxx
```

* `-audio_app_id`: ID приложения распознавания речи
* `-audio_cluster`: Имя кластера (обычно `volcengine_input_common`)
* `-audio_token`: Токен доступа

Подробнее: [Документация VolcEngine](https://www.volcengine.com/docs/6561/80816)

### 9. Инструменты MCP (`mcp`)

Для использования инструментов Amap (например, геолокации).

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-use_tools=true
```

* `-use_tools`: Активирует инструменты (по умолчанию `false`)

### 10. RAG с ChromaDB (`rag chroma`)

Для использования RAG с ChromaDB и сервисом эмбеддингов OpenAI.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-openai_token=sk-xxxx \
-embedding_type=openai \
-vector_db_type=chroma
```

* `-openai_token`: Токен OpenAI (для эмбеддингов)
* `-embedding_type`: Тип эмбеддингов (указать `openai`)
* `-vector_db_type`: Тип векторной БД (указать `chroma`)

### 11. RAG с Milvus (`rag milvus`)

Для использования RAG с Milvus и сервисом эмбеддингов Gemini.

```bash
./telegram-deepseek-bot \
-telegram_bot_token=xxxx \
-deepseek_token=sk-xxx \
-gemini_token=xxx \
-embedding_type=gemini \
-vector_db_type=milvus
```

* `-gemini_token`: Токен Gemini (для эмбеддингов)
* `-embedding_type`: Тип эмбеддингов (указать `gemini`)
* `-vector_db_type`: Тип векторной БД (указать `milvus`)

### 12. RAG с Weaviate (`rag weaviate`)

Для использования RAG с Weaviate и сервисом эмбеддингов Ernie.

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

* `-ernie_ak`: Ernie Access Key
* `-ernie_sk`: Ernie Secret Key
* `-embedding_type`: Тип эмбеддингов (указать `ernie`)
* `-vector_db_type`: Тип векторной БД (указать `weaviate`)
* `-weaviate_url`: URL базы данных Weaviate