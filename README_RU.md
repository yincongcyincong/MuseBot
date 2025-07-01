Вот перевод на русский язык с сохранением разметки и стиля:

## Группа

telegram-группа: https://t.me/+WtaMcDpaMOlhZTE1, или можете попробовать бота `GWSbot`.  
У каждого есть **3000** токенов для тестирования этого бота, пожалуйста, поставьте звезду!

# DeepSeek Telegram Bot

Этот репозиторий предоставляет **Telegram бота**, написанного на **Golang**, который интегрируется с **DeepSeek API** для предоставления ответов на основе ИИ. Бот поддерживает **потоковые ответы**, делая взаимодействие более естественным и динамичным.  
[Документация на китайском](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/Readme_ZH.md)

## 🚀 Возможности

- 🤖 **Ответы ИИ**: Использует DeepSeek API для ответов чат-бота.
- ⏳ **Потоковый вывод**: Отправляет ответы в реальном времени для улучшения пользовательского опыта.
- 🏗 **Простое развертывание**: Запускайте локально или развертывайте на облачном сервере.
- 👀 **Распознавание изображений**: общение с DeepSeek через изображения, см. [документацию](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/imageconf.md).
- 🎺 **Поддержка голоса**: общение с DeepSeek через голосовые сообщения, см. [документацию](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/audioconf.md).
- 🐂 **Вызов функций**: преобразование протокола MCP в вызов функций, см. [документацию](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/functioncall.md).
- 🌊 **RAG**: Поддержка RAG для заполнения контекста, см. [документацию](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/rag.md).
- ⛰️ **OpenRouter**: Поддержка OpenRouter (более 400 языковых моделей), см. [документацию](https://openrouter.ai/docs/quickstart).

## 🤖 Пример текста

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f6b5cdc7-836f-410f-a784-f7074a672c0e" />  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/621861a4-88d1-4796-bf35-e64698ab1b7b" />

## 🎺 Пример мультимодальности

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b4057dce-9ea9-4fcc-b7fa-bcc297482542" />  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/67ec67e0-37a4-4998-bee0-b50463b87125" />

## 📌 Требования

- [Go 1.24+](https://go.dev/dl/)
- [Токен Telegram бота](https://core.telegram.org/bots/tutorial#obtain-your-bot-token)
- [Токен авторизации DeepSeek](https://api-docs.deepseek.com/zh-cn/)

## 📥 Установка

1. **Клонируйте репозиторий**
   ```sh
   git clone https://github.com/yincongcyincong/telegram-deepseek-bot.git
   cd deepseek-telegram-bot
   ```
2. **Установите зависимости**
   ```sh
   go mod tidy
   ```

3. **Настройте переменные окружения**
   ```sh
   export TELEGRAM_BOT_TOKEN="ваш_токен_телеграм_бота"
   export DEEPSEEK_TOKEN="ваш_ключ_api_deepseek"
   ```

## 🚀 Использование

Запустите бота локально:

   ```sh
   go run main.go -telegram_bot_token=токен-телеграм-бота -deepseek_token=токен-авторизации-deepseek
   ```

Используйте Docker:

   ```sh
   docker pull jackyin0822/telegram-deepseek-bot:latest
   docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="токен-телеграм-бота" -e DEEPSEEK_TOKEN="токен-авторизации-deepseek" --name my-telegram-bot jackyin0822/telegram-deepseek-bot:latest
   ```

## ⚙️ Конфигурация

Вы можете настроить бота через переменные окружения:

| Имя переменной                  | Описание                                                                                                                   | Значение по умолчанию       |
|--------------------------------|---------------------------------------------------------------------------------------------------------------------------|----------------------------|
| TELEGRAM_BOT_TOKEN (обязательно)| Токен вашего Telegram бота                                                                                                | -                          |
| DEEPSEEK_TOKEN (обязательно)    | API-ключ DeepSeek                                                                                                         | -                          |
| OPENAI_TOKEN                    | Токен OpenAI                                                                                                              | -                          |
| GEMINI_TOKEN                    | Токен Gemini                                                                                                              | -                          |
| OPEN_ROUTER_TOKEN               | Токен OpenRouter [документация](https://openrouter.ai/docs/quickstart)                                                     | -                          |
| VOL_TOKEN                       | Токен Vol [документация](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                          | -                          |
| CUSTOM_URL                      | пользовательский URL DeepSeek                                                                                             | https://api.deepseek.com/  |
| TYPE                            | deepseek/openai/gemini/openrouter/vol                                                                                     | deepseek                   |
| VOLC_AK                         | AK для модели фото Volcengine [документация](https://www.volcengine.com/docs/6444/1340578)                                 | -                          |
| VOLC_SK                         | SK для модели фото Volcengine [документация](https://www.volcengine.com/docs/6444/1340578)                                 | -                          |
| Ernie_AK                        | AK для Ernie [документация](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                       | -                          |
| Ernie_SK                        | SK для Ernie [документация](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                                       | -                          |
| DB_TYPE                         | sqlite3 / mysql                                                                                                           | sqlite3                    |
| DB_CONF                         | ./data/telegram_bot.db / root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local                   | ./data/telegram_bot.db     |
| ALLOWED_TELEGRAM_USER_IDS       | ID пользователей Telegram, только они могут использовать бота (разделитель ","). Пусто — все пользователи. 0 — все забанены. | -                          |
| ALLOWED_TELEGRAM_GROUP_IDS      | ID чатов Telegram, только они могут использовать бота (разделитель ","). Пусто — все чаты. 0 — все чаты забанены.          | -                          |
| DEEPSEEK_PROXY                  | прокси для DeepSeek                                                                                                       | -                          |
| TELEGRAM_PROXY                  | прокси для Telegram                                                                                                       | -                          |
| LANG                            | en / zh                                                                                                                   | en                         |
| TOKEN_PER_USER                  | Количество токенов, доступных каждому пользователю                                                                        | 10000                      |
| ADMIN_USER_IDS                  | ID администраторов (могут использовать административные команды)                                                           | -                          |
| NEED_AT_BOT                     | необходимо ли упоминание бота в группе для активации                                                                       | false                      |
| MAX_USER_CHAT                   | максимальное количество активных чатов на пользователя                                                                     | 2                          |
| VIDEO_TOKEN                     | API-ключ Volcengine для видео [документация](https://www.volcengine.com/docs/82379/1399008#b00dee71)                      | -                          |
| HTTP_PORT                       | порт HTTP-сервера                                                                                                         | 36060                      |
| USE_TOOLS                       | использовать ли вызов функций в обычном диалоге                                                                            | false                      |

### CUSTOM_URL

Если вы используете самостоятельно развернутый DeepSeek, вы можете установить `CUSTOM_URL` для перенаправления запросов.

### DEEPSEEK_TYPE

- `deepseek`: прямое использование сервиса DeepSeek (не всегда стабильно).
- Другие варианты: см. [документацию](https://www.volcengine.com/docs/82379/1463946).

### DB_TYPE

Поддерживается `sqlite3` или `mysql`.

### DB_CONF

- Если `DB_TYPE = sqlite3`, укажите путь к файлу, например: `./data/telegram_bot.db`.
- Если `DB_TYPE = mysql`, укажите строку подключения, например:  
  `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` (база данных должна быть создана).

### LANG

Выберите язык бота: английский (`en`), китайский (`zh`), русский (`ru`).

### Другие настройки

- [Конфигурация DeepSeek](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/deepseekconf.md)
- [Конфигурация фото](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/photoconf.md)
- [Конфигурация видео](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/videoconf.md)
- [Конфигурация аудио](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/static/doc/audioconf.md)

## Команды

### /clear

Очищает всю историю вашего общения с DeepSeek (используется для контекста).

### /retry

Повторить последний вопрос.

### /mode

Выбор режима DeepSeek: `chat`, `coder`, `reasoner`.  
`chat` и `coder` — DeepSeek-V3, `reasoner` — DeepSeek-R1.  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/55ac3101-92d2-490d-8ee0-31a5b297e56e" />

### /balance

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/23048b44-a3af-457f-b6ce-3678b6776410" />

### /state

Показывает использование токенов пользователем.  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/0814b3ac-dcf6-4ec7-ae6b-3b8d190a0132" />

### /photo

Создание изображений через модель Volcengine (DeepSeek пока не поддерживает генерацию изображений). Требуются `VOLC_AK` и `VOLC_SK`. [Документация](https://www.volcengine.com/docs/6444/1340578)  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/c8072d7d-74e6-4270-8496-1b4e7532134b" />

### /video

Создание видео. `DEEPSEEK_TOKEN` должен быть API-ключом Volcengine. DeepSeek пока не поддерживает генерацию видео. [Документация](https://www.volcengine.com/docs/82379/1399008#b00dee71)  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/884eeb48-76c4-4329-9446-5cd3822a5d16" />

### /chat

Позволяет боту отвечать в группах через команду `/chat`, даже если бот не является администратором.  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/00a0faf3-6037-4d84-9a33-9aa6c320e44d" />

### /help

<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/869e0207-388b-49ca-b26a-378f71d58818" />

### /task

Мультиагентное взаимодействие!

## Административные команды

### /addtoken

Добавляет токены пользователю.  
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/12d98272-0718-4c9b-bc5c-e0a92e6c8664" />

## Развертывание

### Развертывание с Docker

1. **Соберите Docker-образ**
   ```sh
   docker build -t deepseek-telegram-bot .
   ```

2. **Запустите контейнер**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="токен-телеграм-бота" -e DEEPSEEK_TOKEN="токен-авторизации-deepseek" --name my-telegram-bot telegram-deepseek-bot
   ```

## Участие

Вы можете предлагать улучшения, сообщать об ошибках или отправлять pull-запросы. 🚀

## Лицензия

MIT License © 2025 jack yin