# Telegram DeepSeek Bot Management Platform README

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
| **DB\_TYPE** | Database type: `sqlite3` / `mysql`                                                                                                       | `sqlite3`                    |
| **DB\_CONF** | Database configuration: `./data/telegram_bot.db` or `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` | `./data/telegram_bot.db`     |
| **SESSION\_KEY** | Specifies the key used for session management.                                                                                           | `telegram_bot_session_key` (A string used to encrypt session data) |

-----

## Getting Started

### Login

Access the management platform's login page.

### Default Account

Upon first launch, you can log in using the following default credentials:

* **Username:** `admin`
* **Password:** `admin`

-----

## Platform Modules Overview

### Home Page

An overview of the platform's home page.

### Admin Page

A list of administrators for the management platform.

#### Add Admin

On this page, you can add new administrator accounts and grant them platform management permissions.

### Bot Management

Manage your configured Telegram bots.

#### Add Bot

Configure and add new Telegram bots on this page.

### Bot Users

View and manage all users who interact with your bots.

### Add Token to User

Allocate and manage API tokens for specific users to control their access and usage limits for the bot.

### Chat History Page

This page displays the complete chat history between the bot and users, making it easy to track and analyze conversations.

-----

## Deployment and Configuration

The platform supports various configurations, including:

* **LLM Configuration**: You can configure different LLM services as needed, such as DeepSeek types and custom URLs.
* **Database Type**: Supports `sqlite3` or `mysql`; users can choose the configuration based on their needs.
* **Language Settings**: Supports English (en), Chinese (zh), and Russian (ru).

For more detailed configurations, please refer to the documentation in the GitHub repository.