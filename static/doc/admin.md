# Telegram DeepSeek Bot Management Platform

This platform is designed to manage your Telegram DeepSeek Bot. It integrates various Large Language Models (LLMs), provides context-aware responses, and supports multiple models for diverse interactions.

-----

## Running the Platform and Parameters

To start the management platform, execute the following command:

```bash
./admin -db_type=sqlite3 -db_conf=./admin/data/telegram_bot.db -session_key=telegram_bot_session_key
```

### Command Parameters

| Variable Name   | Description                                                                                                                              | Default Value                |
| :-------------- | :--------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------- |
| **DB\_TYPE** | Database type: `sqlite3` / `mysql`                                                                                                       | `sqlite3` / `mysql`                   |
| **DB\_CONF** | Database configuration: `./data/telegram_bot.db` or `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` | `./data/telegram_bot.db`     |
| **SESSION\_KEY** | Specifies the key used for session management.                                                                                           | `telegram_bot_session_key` (A string used to encrypt session data) |

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
![image](https://github.com/user-attachments/assets/5097e64e-1a89-41b4-a6a3-f4242be17cd0)

Configure and add new Telegram bots on this page.

### Bot Users
![image](https://github.com/user-attachments/assets/5534971a-e1e2-42d1-9552-0ce37b18444f)

View and manage all users who interact with your bots.

### Add Token to User
![image](https://github.com/user-attachments/assets/b9ffc006-764c-46b7-a5ce-703b052c5368)

Allocate and manage API tokens for specific users to control their access and usage limits for the bot.

### Chat History Page
![image](https://github.com/user-attachments/assets/7b0a834f-0e62-4bec-9d57-1be22da0828d)

This page displays the complete chat history between the bot and users, making it easy to track and analyze conversations.

