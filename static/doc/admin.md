# Telegram DeepSeek Bot Management Platform

This platform is designed to manage your Telegram DeepSeek Bot. It integrates various Large Language Models (LLMs),
provides context-aware responses, and supports multiple models for diverse interactions.

-----

## Running the Platform and Parameters

To start the management platform, execute the following command:

```bash
./admin -db_type=sqlite3 -db_conf=./data/telegram_admin_bot.db -session_key=telegram_bot_session_key
```

### Command Parameters

| Variable Name    | Description                                                                                                                                | Default Value                                                      |
|:-----------------|:-------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------|
| **DB\_TYPE**     | Database type: `sqlite3` / `mysql`                                                                                                         | `sqlite3` / `mysql`                                                |
| **DB\_CONF**     | Database configuration: `./data/telegram_admin_bot.db` or `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` | `./data/telegram_admin_bot.db`                                     |
| **SESSION\_KEY** | Specifies the key used for session management.                                                                                             | `telegram_bot_session_key` (A string used to encrypt session data) |
| **ADMIN\_PORT**  | port for admin platform                                                                                                                    | `18080`                                                            |

-----

## Getting Started

### Login

Access the management platform's login page.
![image](https://github.com/user-attachments/assets/f6bf8ae6-4c0e-44d9-9115-7e744fc20dc3)

### Default Account

Upon first launch, you can log in using the following default credentials:

* **Username:** `admin`
* **Password:** `admin`

-----

## Platform Modules Overview

### Home Page

![image](https://github.com/user-attachments/assets/b12925ca-8d02-4537-84bd-6b0e1ca1686f)

An overview of the platform's home page.

### Admin Page

![image](https://github.com/user-attachments/assets/0f5ccb12-1733-44d4-8922-c0dbd9966372)

A list of administrators for the management platform.

#### Add Admin

![image](https://github.com/user-attachments/assets/89c46bc4-4ff5-455d-8dcd-6bfdc275659a)

On this page, you can add new administrator accounts and grant them platform management permissions.

### Bot Management

![image](https://github.com/user-attachments/assets/518f9341-9e30-41b5-a71f-fff3e398ace0)

Manage your configured Telegram bots.

#### Add Bot

Configure and add new Telegram bots on this page. For enhanced security, it's highly recommended to use **HTTP mutual
authentication**.
Start telegram-deepseek-bot in this way:

```
./telegram-deepseek-bot \
-telegram_bot_token=xxx \
-deepseek_token=sk-xxx \
-crt_file=/Users/yincong/go/src/github.com/yincongcyincong/telegram-deepseek-bot/admin/shell/certs/server.crt
-ca_file=/Users/yincong/go/src/github.com/yincongcyincong/telegram-deepseek-bot/admin/shell/certs/ca.crt
-key_file=/Users/yincong/go/src/github.com/yincongcyincong/telegram-deepseek-bot/admin/shell/certs/server.key
```

add configurations in admin page:
![image](https://github.com/user-attachments/assets/2a518841-abf6-4a31-b1b3-b26b258a5fab)

can use this [file](https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/admin/shell/generate_cert.sh) to
generate ca, key and crt.

#### Bot Start Parameter

![image](https://github.com/user-attachments/assets/94c65d03-e097-479e-bf2a-f3d5aad431cc)

Show all Parameters when starting the Telegram DeepSeek Bot.

#### Bot Config

![image](https://github.com/user-attachments/assets/0e6d3c32-5311-4769-ac42-e9591d4651ad)

Modify the configuration of your bot.

### MCP Shop
![image](https://github.com/user-attachments/assets/9ade4136-b261-462c-b59b-8755d71fb7a5)
multi mcp clients for telegram deepseek bot

### Add MCP 
![image](https://github.com/user-attachments/assets/9c6679d4-1417-49fa-ad55-3279e2b55995)
add mcp client for telegram deepseek bot


### Bot Users

![image](https://github.com/user-attachments/assets/5534971a-e1e2-42d1-9552-0ce37b18444f)

View and manage all users who interact with your bots.

### Add Token to User

![image](https://github.com/user-attachments/assets/b9ffc006-764c-46b7-a5ce-703b052c5368)

Allocate and manage API tokens for specific users to control their access and usage limits for the bot.

### Chat History Page

![image](https://github.com/user-attachments/assets/7b0a834f-0e62-4bec-9d57-1be22da0828d)

This page displays the complete chat history between the bot and users, making it easy to track and analyze
conversations.

### Chat Page
![image](https://github.com/user-attachments/assets/b8c9c3e0-467b-44b2-9186-f0c9344b5633)
chat with your telegram deepseek bot 
