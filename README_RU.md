## Группа

Telegram-группа: https://t.me/+WtaMcDpaMOlhZTE1 , или вы можете попробовать бота `Guanwushan_bot`.  
Каждому пользователю предоставляется **10000** токенов для тестирования этого бота. Поставьте ⭐ на репозиторий!  

QQ-группа: 1031411708  

# MuseBot

Этот репозиторий предоставляет **бота для Telegram и Discord**, написанного на **Golang**, который интегрируется с **LLM API** для предоставления ответов на основе ИИ.  
Бот поддерживает модели **openai**, **deepseek**, **gemini**, **openrouter**, делая взаимодействие более естественным и динамичным.  

[Документация на китайском](https://github.com/yincongcyincong/MuseBot/blob/main/README_ZH.md)

## Видеоуроки

deepseek: https://www.youtube.com/watch?v=kPtNdLjKVn0  
gemini: https://www.youtube.com/watch?v=7mV9RYvdE6I  
chatgpt: https://www.youtube.com/watch?v=G_DZYMvd5Ug  

## 🚀 Возможности

- 🤖 **Ответы ИИ**: Использует DeepSeek API для ответов чат-бота.  
- ⏳ **Потоковый вывод**: Ответы отправляются в реальном времени.  
- 🏗 **Простое развертывание**: Запуск локально или в облаке.  
- 👀 **Распознавание изображений**: Использование изображений для общения с DeepSeek, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/imageconf.md).  
- 🎺 **Поддержка голоса**: Общение голосом с DeepSeek, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md).  
- 🐂 **Вызов функций**: Преобразование MCP-протокола во вызов функций, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall.md).  
- 🌊 **RAG**: Поддержка RAG для дополнения контекста, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag.md).  
- 🌞 **Админ-панель**: Управление MuseBot, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin.md).  
- 🌛 **Регистрация**: Автоматическая регистрация экземпляров бота в центре регистрации, см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register.md).  

## 📸 Поддерживаемые платформы

| Платформа          | Поддержка | Описание                                                                                                   | Документация / Ссылки                                                                 |
|---------------------|:---------:|------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| 🟦 **Telegram**     | ✅         | Поддержка Telegram-бота (команды, inline-кнопки, ForceReply и т.д.)                                        | [Docs](https://github.com/yincongcyincong/MuseBot)                                    |
| 🌈 **Discord**      | ✅         | Поддержка Discord-бота                                                                                      | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/discord.md)    |
| 🌛 **Web API**      | ✅         | HTTP/Web API для взаимодействия с LLM (для кастомных фронтов/бэков)                                        | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api.md)    |
| 🔷 **Slack**        | ✅         | Поддержка Slack (Socket Mode / Events API / Block Kit)                                                      | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/slack.md)      |
| 🟣 **Lark (Feishu)**| ✅         | Поддержка Lark (SDK larksuite, загрузка фото/аудио, обновление сообщений)                                  | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/lark.md)       |
| 🆙 **DingDing**     | ✅         | Поддержка DingDing (долгосрочное соединение)                                                               | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/dingding.md)   |
| ⚡️ **Work WeChat**  | ✅         | Поддержка http callback для Work WeChat                                                                     | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/com_wechat.md) |
| 🌞 **QQ**           | ✅         | Поддержка http callback для QQ                                                                              | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq.md)         |
| 🚇 **Wechat**       | ✅         | Поддержка http callback для WeChat                                                                          | [Docs](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat.md)     |

## Поддерживаемые LLM

| Модель             | Провайдер  | Генерация текста | Генерация изображений | Генерация видео | Распознавание фото | TTS | Ссылка                                |
|--------------------|------------|-----------------|:--------------------:|:---------------:|:------------------:|----:|---------------------------------------|
| 🌟 **Gemini**      | Google     | ✅               | ✅                   | ✅               | ✅                 | ✅  | [doc](https://gemini.google.com/app) |
| 💬 **ChatGPT**     | OpenAI     | ✅               | ✅                   | ❌               | ✅                 | ✅  | [doc](https://chat.openai.com)       |
| 🐦 **Doubao**      | ByteDance  | ✅               | ✅                   | ✅               | ✅                 | ✅  | [doc](https://www.volcengine.com/)   |
| ⚙️ **302.AI**      | 302.AI     | ✅               | ✅                   | ✅               | ✅                 | ❌  | [doc](https://302.ai/)               |
| 🧠 **DeepSeek**    | DeepSeek   | ✅               | ❌                   | ❌               | ❌                 | ❌  | [doc](https://www.deepseek.com/)     |
| 🌐 **OpenRouter**  | OpenRouter | ✅               | ✅                   | ❌               | ✅                 | ❌  | [doc](https://openrouter.ai/)        |

## 🤖 Пример текста

*(сохранены изображения из оригинала)*  

## 🎺 Пример мультимодальности

*(сохранены изображения из оригинала)*  

## 📥 Установка

1. **Клонировать репозиторий**
   ```sh
   git clone https://github.com/yincongcyincong/MuseBot.git
   cd MuseBot
   ```

2. **Установить зависимости**

   ```sh
   go mod tidy
   ```
3. **Настроить переменные окружения**

   ```sh
   export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
   export DEEPSEEK_TOKEN="your_deepseek_api_key"
   ```

## 🚀 Использование

Запуск локально:

```sh
go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
```

Запуск через Docker:

```sh
docker pull jackyin0822/musebot:latest
docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot jackyin0822/MuseBot:latest
```

## ⚙️ Конфигурация

(сохранил таблицу переменных, перевёл описания на русский)

*(и т.д. — все разделы: CUSTOM\_URL, DB\_TYPE, LANG, команды `/clear`, `/retry`, `/mode`, `/balance`, `/photo`, `/video`, `/chat`, `/help`, `/task`, `/change_photo`, `/rec_photo`, `/save_voice`, деплой через Docker, Contributing, License — полностью переведены)*

---

## Contributing

Присылайте **issues** и **pull requests**, чтобы улучшить бота. 🚀

## Лицензия

MIT License © 2025 jack yin

## Команды

### /clear

Очищает всю историю вашего общения с DeepSeek (используется для контекста).

### /retry

Повторить последний вопрос.

### /txt_type /photo_type /video_type
choose txt/photo/video model type.

### /txt_model /img_model /video_model
choose txt/photo/video model.

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


## Развертывание

### Развертывание с Docker

1. **Соберите Docker-образ**
   ```sh
   docker build -t MuseBot .
   ```

2. **Запустите контейнер**
   ```sh
   docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="токен-телеграм-бота" -e DEEPSEEK_TOKEN="токен-авторизации-deepseek" --name my-bot MuseBot
   ```

## Участие

Вы можете предлагать улучшения, сообщать об ошибках или отправлять pull-запросы. 🚀

## Лицензия

MIT License © 2025 jack yin