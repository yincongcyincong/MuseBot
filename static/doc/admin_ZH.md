# Telegram DeepSeek Bot 管理平台 README

这是一个用于管理 Telegram DeepSeek Bot 的平台，它集成了多种大型语言模型（LLMs），提供上下文感知的响应，并支持多模型以实现多样化的交互。

## 执行命令与参数

要启动管理平台，请执行以下命令：

```bash
./admin -db_type=sqlite3 -db_conf=./admin/data/telegram_bot.db -session_key=telegram_bot_session_key
```

### 命令参数列表

| 变量名                            | 描述                                                                                                            | 默认值                       |
| :----------- | :---------------------------------- | :--------------------------- |
| **DB_TYPE**  | 数据库类型：sqlite3 / mysql                                                                                         | sqlite3                   |
| **DB_CONF**                    | 数据库配置：./data/telegram_bot.db 或 root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local | ./data/telegram_bot.db    |
| **SESSION_KEY** | 指定用于会话管理的密钥。            | `telegram_bot_session_key` (一个用于加密会话数据的字符串) |

## 开始使用

### 登录

访问管理平台登录页面。

### 默认账号

首次启动时，可以使用以下默认凭据登录：

* **用户名：** `admin`
* **密码：** `admin`

## 平台模块介绍

### 首页

平台首页概览。

### 管理员页面

管理平台的管理员列表。

#### 添加管理员

在此页面可以添加新的管理员账号，赋予其管理平台的权限。

### 机器人管理

对已配置的 Telegram 机器人进行管理。

#### 添加机器人

在此页面配置并添加新的 Telegram 机器人。

### 机器人用户

查看和管理所有与机器人交互的用户。

### 给用户添加 Token

为特定用户分配和管理 API Token，用于控制其使用机器人的权限和额度。

### 聊天记录页面

此页面展示机器人与用户的完整聊天记录，便于追溯和分析对话内容。

## 部署与配置

该平台支持多种配置，包括：

* **LLM 配置**：可根据需求配置不同的 LLM 服务，例如 DeepSeek 类型和自定义 URL。
* **数据库类型**：支持 `sqlite3` 或 `mysql`，用户可以根据实际情况选择配置。
* **语言设置**：支持英文 (en)、中文 (zh) 和俄文 (ru)。

更多详细配置请参考 GitHub 仓库中的文档。

