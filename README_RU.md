# MuseBot

Этот репозиторий предоставляет **Чат-бот** (Telegram, Discord, Slack, Lark（飞书）, 钉钉, 企业微信, QQ, 微信), который
интегрируется
с **LLM API** для предоставления ответов на основе искусственного интеллекта. Бот поддерживает LLM от **openai** *
*deepseek** **gemini** **openrouter** **orcarouter**, что делает взаимодействие более естественным и динамичным.
[中文文档 (Китайская документация)](https://github.com/yincongcyincong/MuseBot/blob/main/README_ZH.md)
[Китайская документация (Русский перевод)](https://github.com/yincongcyincong/MuseBot/blob/main/README_RU.md)

## Видео по использованию

Самый простой способ
использования: [https://www.youtube.com/watch?v=4UHoKRMfNZg](https://www.youtube.com/watch?v=4UHoKRMfNZg)
deepseek: [https://www.youtube.com/watch?v=kPtNdLjKVn0](https://www.youtube.com/watch?v=kPtNdLjKVn0)
gemini: [https://www.youtube.com/watch?v=7mV9RYvdE6I](https://www.youtube.com/watch?v=7mV9RYvdE6I)
chatgpt: [https://www.youtube.com/watch?v=G\_DZYMvd5Ug](https://www.youtube.com/watch?v=G_DZYMvd5Ug)

## 🚀 Функции

- 🤖 **Ответы ИИ**: Использует LLM API для ответов чат-бота.
- ⏳ **Потоковая передача вывода (Streaming Output)**: Отправляет ответы в реальном времени для улучшения
  пользовательского опыта.
- 🏗 **Простое развертывание**: Запуск локально или развертывание на облачном сервере.
- 👀 **Распознавание изображений**: Используйте изображения для общения с LLM,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/imageconf.md).
- 🎺 **Поддержка голоса**: Используйте голос для общения с LLM,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md).
- 🐂 **Вызов функций (Function Call)**: Преобразование протокола mcp в вызов функций,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/functioncall.md).
- 🌊 **RAG**: Поддержка RAG для заполнения контекста,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/rag.md).
- 🌞 **Админ-платформа (AdminPlatform)**: Используйте платформу для управления MuseBot,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/admin.md).
- 🌛 **Регистрация**: С помощью модуля регистрации служб экземпляры роботов могут быть автоматически зарегистрированы в
  центре регистрации [документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/register.md)
- 🌈 **Метрики**: Поддержка Метрик для мониторинга,
  см. [документацию](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/metrics.md).

## 📸 Поддерживаемые платформы

| Платформа            | Поддержка | Описание                                                                                                                                    | Документация / Ссылки                                                                         |
|----------------------|:---------:|---------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| 🟦 **Telegram**      |     ✅     | Поддерживает Telegram-бота (на основе go-telegram-bot-api, обрабатывает команды, встроенные кнопки, ForceReply и т. д.)                     | [Документация](https://github.com/yincongcyincong/MuseBot)                                    |
| 🌈 **Discord**       |     ✅     | Поддерживает Discord-бота                                                                                                                   | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/discord.md)    |
| 🌛 **Web API**       |     ✅     | Предоставляет HTTP/Web API для взаимодействия с LLM (отлично подходит для пользовательских фронтендов/бэкендов)                             | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/web_api.md)    |
| 🔷 **Slack**         |     ✅     | Поддерживает Slack (Socket Mode / Events API / Block Kit взаимодействия)                                                                    | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/slack.md)      |
| 🟣 **Lark (Feishu)** |     ✅     | Поддерживает длинное соединение Lark и обработку сообщений (на основе SDK larksuite, с загрузкой изображений/аудио и обновлением сообщений) | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/lark.md)       |
| 🆙 **DingDing**      |     ✅     | Поддерживает длинное соединение Dingding                                                                                                    | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/dingding.md)   |
| ⚡️ **Work WeChat**   |     ✅     | Поддерживает HTTP-обратный вызов Work WeChat для запуска LLM                                                                                | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/com_wechat.md) |
| 🌞 **QQ**            |     ✅     | Поддерживает HTTP-обратный вызов QQ для запуска LLM                                                                                         | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/qq.md)         |
| 🚇 **Wechat**        |     ✅     | Поддерживает HTTP-обратный вызов Wechat для запуска LLM                                                                                     | [Документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/wechat.md)     |

## Поддерживаемые большие языковые модели

| Модель              | Провайдер    | Генерация текста | Генерация изображений | Генерация видео | Распознавание фото | TTS | Ссылка                                                                                                        |
|---------------------|--------------|------------------|:---------------------:|:---------------:|-------------------:|----:|---------------------------------------------------------------------------------------------------------------|
| 🌟 **Gemini**       | Google       | ✅                |           ✅           |        ✅        |                  ✅ |   ✅ | [doc](https://gemini.google.com/app)                                                                          |
| 💬 **ChatGPT**      | OpenAI       | ✅                |           ✅           |        ❌        |                  ✅ |   ✅ | [doc](https://chat.openai.com)                                                                                |
| 🐦 **Doubao**       | ByteDance    | ✅                |           ✅           |        ✅        |                  ✅ |   ✅ | [doc](https://www.volcengine.com/)                                                                            |
| 🐦 **Qwen**         | Aliyun       | ✅                |           ✅           |        ✅        |                  ✅ |   ✅ | [doc](https://bailian.console.aliyun.com/?spm=5176.12818093_47.overview_recent.1.663b2cc9wXXcVC&tab=api#/api) |
| 🧠 **DeepSeek**     | DeepSeek     | ✅                |           ❌           |        ❌        |                  ❌ |   ❌ | [doc](https://www.deepseek.com/)                                                                              |
| ⚙️ **302.AI**       | 302.AI       | ✅                |           ✅           |        ✅        |                  ✅ |   ❌ | [doc](https://302.ai/)                                                                                        |
| 🌐 **OpenRouter**   | OpenRouter   | ✅                |           ✅           |        ❌        |                  ✅ |   ❌ | [doc](https://openrouter.ai/docs/quickstart)                                                                  |
| 🛡️ **OrcaRouter**  | OrcaRouter   | ✅                |           ✅           |        ❌        |                  ✅ |   ❌ | [doc](https://www.orcarouter.ai)                                                                               |
| 🌐 **ChatAnywhere** | ChatAnywhere | ✅                |           ✅           |        ❌        |                  ✅ |   ❌ | [doc](https://api.chatanywhere.tech/#/)                                                                       |

## 🤖 Пример текста

## 🎺 Пример мультимодальности

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

3. **Настроить переменные среды**

   ```sh
    export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
    export DEEPSEEK_TOKEN="your_deepseek_api_key"
   ```

## 🚀 Использование

Запустить бота локально:

```sh
 go run main.go -telegram_bot_token=telegram-bot-token -deepseek_token=deepseek-auth-token
```

Использовать docker

```sh
  docker pull jackyin0822/musebot:latest
  docker run -d -v /home/user/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-telegram-bot  jackyin0822/musebot:latest
```

```sh
 ALIYUN:
 docker pull crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest
```

команда: [документация](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/param_conf.md)

## ⚙️ Конфигурация

Вы можете настроить бота с помощью переменных среды:

Если вы используете параметры. Пожалуйста, используйте строчные буквы и нижнее подчеркивание. Например:
`./MuseBot -telegram_bot_token=xxx`

| Название переменной                 | Описание                                                                                                                | Значение по умолчанию                                  |
|-------------------------------------|-------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| **TELEGRAM\_BOT\_TOKEN**            | Токен Telegram-бота                                                                                                     | -                                                      |
| **DISCORD\_BOT\_TOKEN**             | Токен Discord-бота                                                                                                      | -                                                      |
| **SLACK\_BOT\_TOKEN**               | Токен Slack-бота                                                                                                        | -                                                      |
| **SLACK\_APP\_TOKEN**               | Токен Slack уровня приложения                                                                                           | -                                                      |
| **LARK\_APP\_ID**                   | ID приложения Lark (Feishu)                                                                                             | -                                                      |
| **LARK\_APP\_SECRET**               | Секрет приложения Lark (Feishu)                                                                                         | -                                                      |
| **DING\_CLIENT\_ID**                | App Key / Client ID DingTalk                                                                                            | -                                                      |
| **DING\_CLIENT\_SECRET**            | App Secret DingTalk                                                                                                     | -                                                      |
| **DING\_TEMPLATE\_ID**              | ID шаблонного сообщения DingTalk                                                                                        | -                                                      |
| **COM\_WECHAT\_TOKEN**              | Токен WeCom (Enterprise WeChat)                                                                                         | -                                                      |
| **COM\_WECHAT\_ENCODING\_AES\_KEY** | EncodingAESKey WeCom                                                                                                    | -                                                      |
| **COM\_WECHAT\_CORP\_ID**           | CorpID WeCom                                                                                                            | -                                                      |
| **COM\_WECHAT\_SECRET**             | App Secret WeCom                                                                                                        | -                                                      |
| **COM\_WECHAT\_AGENT\_ID**          | Agent ID WeCom                                                                                                          | -                                                      |
| **WECHAT\_APP\_ID**                 | AppID официального аккаунта WeChat                                                                                      | -                                                      |
| **WECHAT\_APP\_SECRET**             | AppSecret официального аккаунта WeChat                                                                                  | -                                                      |
| **WECHAT\_ENCODING\_AES\_KEY**      | EncodingAESKey официального аккаунта WeChat                                                                             | -                                                      |
| **WECHAT\_TOKEN**                   | Токен официального аккаунта WeChat                                                                                      | -                                                      |
| **WECHAT\_ACTIVE**                  | Включить ли прослушивание сообщений WeChat (true/false)                                                                 | false                                                  |
| **QQ\_APP\_ID**                     | AppID открытой платформы QQ                                                                                             | -                                                      |
| **QQ\_APP\_SECRET**                 | AppSecret открытой платформы QQ                                                                                         | -                                                      |
| **QQ\_ONEBOT\_RECEIVE\_TOKEN**      | Токен для сообщений о событиях ONEBOT → MuseBot                                                                         | MuseBot                                                |
| **QQ\_ONEBOT\_SEND\_TOKEN**         | Токен для отправки сообщений MuseBot → ONEBOT                                                                           | MuseBot                                                |
| **QQ\_ONEBOT\_HTTP\_SERVER**        | Адрес HTTP-сервера ONEBOT                                                                                               | [http://127.0.0.1:3000](http://127.0.0.1:3000)         |
| **DEEPSEEK\_TOKEN**                 | Ключ API DeepSeek                                                                                                       | -                                                      |
| **OPENAI\_TOKEN**                   | Ключ API OpenAI                                                                                                         | -                                                      |
| **GEMINI\_TOKEN**                   | Токен API Google Gemini                                                                                                 | -                                                      |
| **OPEN\_ROUTER\_TOKEN**             | Токен OpenRouter [doc](https://openrouter.ai/docs/quickstart)                                                           | -                                                      |
| **ORCAROUTER\_TOKEN**             | API-ключ OrcaRouter [doc](https://www.orcarouter.ai)                                                                  | -                                                      |
| **ALIYUN\_TOKEN**                   | Токен Aliyun Bailian [doc](https://bailian.console.aliyun.com/#/doc/?type=model&url=2840915)                            | -                                                      |
| **AI\_302\_TOKEN**                  | Токен 302.AI [doc](https://302.ai/)                                                                                     | -                                                      |
| **VOL\_TOKEN**                      | Общий токен Volcano Engine [doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)                                | -                                                      |
| **VOLC\_AK**                        | Ключ доступа к мультимедиа Volcano Engine [doc](https://www.volcengine.com/docs/6444/1340578)                           | -                                                      |
| **VOLC\_SK**                        | Секретный ключ мультимедиа Volcano Engine [doc](https://www.volcengine.com/docs/6444/1340578)                           | -                                                      |
| **ERNIE\_AK**                       | AK большой модели Baidu ERNIE [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                             | -                                                      |
| **ERNIE\_SK**                       | SK большой модели Baidu ERNIE [doc](https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Sly8bm96d)                             | -                                                      |
| **CUSTOM\_URL**                     | Пользовательская конечная точка API DeepSeek                                                                            | [https://api.deepseek.com/](https://api.deepseek.com/) |
| **TYPE**                            | Тип LLM (deepseek/openai/gemini/openrouter/vol/302-ai/chatanywhere/orcarouter)                                        | deepseek                                               |
| **MEDIA\_TYPE**                     | Источник генерации медиа (openai/gemini/vol/openrouter/aliyun/302-ai/orcarouter)                                      | vol                                                    |
| **DB\_TYPE**                        | Тип базы данных (sqlite3/mysql)                                                                                         | sqlite3                                                |
| **DB\_CONF**                        | Путь к файлу конфигурации базы данных или строка подключения                                                            | ./data/muse\_bot.db                                    |
| **LLM\_PROXY**                      | Сетевой прокси LLM (например, [http://127.0.0.1:7890](http://127.0.0.1:7890))                                           | -                                                      |
| **ROBOT\_PROXY**                    | Сетевой прокси бота (например, [http://127.0.0.1:7890](http://127.0.0.1:7890))                                          | -                                                      |
| **LANG**                            | Язык (en/zh)                                                                                                            | en                                                     |
| **TOKEN\_PER\_USER**                | Максимальное количество токенов, разрешенное для одного пользователя                                                    | 10000                                                  |
| **MAX\_USER\_CHAT**                 | Максимальное количество одновременных чатов для одного пользователя                                                     | 2                                                      |
| **HTTP\_HOST**                      | Порт HTTP-сервера MuseBot                                                                                               | :36060                                                 |
| **USE\_TOOLS**                      | Включить инструменты вызова функций (true/false)                                                                        | false                                                  |
| **MAX\_QA\_PAIR**                   | Максимальное количество пар вопрос-ответ для сохранения в качестве контекста                                            | 100                                                    |
| **CHARACTER**                       | Описание личности ИИ                                                                                                    | -                                                      |
| **CRT\_FILE**                       | Путь к файлу HTTPS-сертификата                                                                                          | -                                                      |
| **KEY\_FILE**                       | Путь к файлу закрытого ключа HTTPS                                                                                      | -                                                      |
| **CA\_FILE**                        | Путь к файлу сертификата центра сертификации HTTPS                                                                      | -                                                      |
| **ADMIN\_USER\_IDS**                | Список ID пользователей-администраторов, разделенных запятыми                                                           | -                                                      |
| **ALLOWED\_USER\_IDS**              | ID пользователей, которым разрешено использовать бота, разделенные запятыми; пусто = разрешено всем; 0 = запрещено всем | -                                                      |
| **ALLOWED\_GROUP\_IDS**             | ID групп, которым разрешено использовать бота, разделенные запятыми; пусто = разрешено всем; 0 = запрещено всем         | -                                                      |
| **BOT\_NAME**                       | Имя бота                                                                                                                | MuseBot                                                |
| **CHAT\_ANY\_WHERE\_TOKEN**         | Токен платформы ChatAnyWhere                                                                                            | -                                                      |

### CUSTOM\_URL

Если вы используете саморазвернутый DeepSeek, вы можете установить CUSTOM\_URL для маршрутизации запросов к вашему
саморазвернутому DeepSeek.

### DEEPSEEK\_TYPE

deepseek: напрямую использовать сервис deepseek. но он не очень стабилен
others: см. [документацию](https://www.volcengine.com/docs/82379/1463946)

### DB\_TYPE

поддержка sqlite3 или mysql

### DB\_CONF

если DB\_TYPE — sqlite3, укажите путь к файлу, например `./data/telegram_bot.db`
если DB\_TYPE — mysql, укажите ссылку на mysql, например
`root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local`, база данных должна быть создана.

### LANG

выберите язык для бота: Английский (`en`), Китайский (`zh`), Русский (`ru`).

### Другая конфигурация

[deepseek\_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/deepseekconf.md)
[photo\_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/photoconf.md)
[video\_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/videoconf.md)
[audio\_conf](https://github.com/yincongcyincong/MuseBot/blob/main/static/doc/audioconf.md)

## Команды

### /clear

Очистить всю историю вашего общения с deepseek. Эта история используется для помощи deepseek в понимании контекста.

### /retry

Повторить последний вопрос.

### /txt\_type /photo\_type /video\_type /rec\_type

Выбрать тип модели для текста/фото/видео/распознавания.

### /txt\_model /img\_model /video\_model /rec\_model

Выбрать модель для текста/фото/видео/распознавания.

### /mode

Показать текущий тип и модель.

### /state

Рассчитать использование токенов одним пользователем.

### /photo /edit\_photo

Использование модели фото volcengine для создания фото, deepseek пока не поддерживает создание фото. VOLC\_AK и VOLC\_SK
обязательны. [doc](https://www.volcengine.com/docs/6444/1340578)

/edit\_photo обновит ваше фото на основе вашего описания.

### /video

Создать видео. `DEEPSEEK_TOKEN` должен быть ключом API volcengine. deepseek пока не поддерживает создание видео.
[doc](https://www.volcengine.com/docs/82379/1399008#b00dee71)

### /chat

Позволяет боту общаться через команду /chat в группах,
без необходимости устанавливать бота в качестве администратора группы.

### /help

### /task

Множество агентов общаются друг с другом\!

### /change\_photo

Только для приложений tencent (wechat, qq, work wechat)
Изменить фото на основе вашего промпта.

### /rec\_photo

Только для приложений tencent (wechat, qq, work wechat)
Распознать фото на основе вашего промпта.

### /save\_voice

Только для приложений tencent (wechat, qq, work wechat)
Сохранить ваш голос на ПК.

## Развертывание

### Развертывание с помощью Docker

1. **Собрать образ Docker**

   ```sh
    docker build -t MuseBot .
   ```

2. **Запустить контейнер**

   ```sh
     docker run -d -v /home/user/xxx/data:/app/data -e TELEGRAM_BOT_TOKEN="telegram-bot-token" -e DEEPSEEK_TOKEN="deepseek-auth-token" --name my-bot MuseBot
   ```

## Участие

Не стесняйтесь отправлять проблемы (issues) и запросы на слияние (pull requests) для улучшения этого бота. 🚀

# Группа

telegram-group: [https://t.me/+WtaMcDpaMOlhZTE1](https://t.me/+WtaMcDpaMOlhZTE1), или вы можете попробовать робота *
*Guanwushan\_bot**.
У каждого есть **10000** токенов, чтобы попробовать этого бота, пожалуйста, поставьте мне звезду\!

## Лицензия

Лицензия MIT © 2025 jack yin