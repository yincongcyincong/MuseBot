# Telegram DeepSeek Bot 管理平台

此平台旨在管理您的 **Telegram DeepSeek Bot**。它集成了多种**大型语言模型 (LLMs)**，提供**上下文感知**的回复，并支持**多模型
**以实现多样化交互。

-----

## 运行平台及参数

要启动管理平台，请执行以下命令：

```bash
./admin -db_type=sqlite3 -db_conf=./admin/data/telegram_bot.db -session_key=telegram_bot_session_key
```

### 命令参数

| 变量名称            | 描述                                                                                                                      | 默认值                                       |
|:----------------|:------------------------------------------------------------------------------------------------------------------------|:------------------------------------------|
| **DB_TYPE**     | 数据库类型：`sqlite3` / `mysql`                                                                                               | `sqlite3` / `mysql`                       |
| **DB_CONF**     | 数据库配置：`./data/telegram_admin_bot.db` 或 `root:admin@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` | `./data/telegram_admin_bot.db`            |
| **SESSION_KEY** | 指定用于会话管理的密钥。                                                                                                            | `telegram_bot_session_key` (用于加密会话数据的字符串) |
| **ADMIN_PORT**  | 管理平台端口                                                                                                                  | `18080`                                   |

-----

## 快速入门

### 登录

![image](https://github.com/user-attachments/assets/f6bf8ae6-4c0e-44d9-9115-7e744fc20dc3)
访问管理平台的登录页面。

### 默认账户

首次启动时，您可以使用以下默认凭据登录：

* **用户名：** `admin`
* **密码：** `admin`

-----

## 平台模块概览

### 首页

![image](https://github.com/user-attachments/assets/b12925ca-8d02-4537-84bd-6b0e1ca1686f)
平台首页概览。

### 管理员页面

![image](https://github.com/user-attachments/assets/0f5ccb12-1733-44d4-8922-c0dbd9966372)
管理平台管理员列表。

#### 添加管理员

![image](https://github.com/user-attachments/assets/89c46bc4-4ff5-455d-8dcd-6bfdc275659a)
在此页面，您可以添加新的管理员账户并授予其平台管理权限。

### 机器人管理

![image](https://github.com/user-attachments/assets/518f9341-9e30-41b5-a71f-fff3e398ace0)
管理您已配置的 Telegram 机器人。

#### 添加机器人

在此页面配置和添加新的 Telegram 机器人。为增强安全性，**强烈建议使用 HTTP 双向认证**。

通过以下方式启动 telegram-deepseek-bot：

```
./telegram-deepseek-bot \
-telegram_bot_token=xxx \
-deepseek_token=sk-xxx \
-crt_file=/Users/yincong/go/src/github.com/yincongcyincong/MuseBot/admin/shell/certs/server.crt
-ca_file=/Users/yincong/go/src/github.com/yincongcyincong/MuseBot/admin/shell/certs/ca.crt
-key_file=/Users/yincong/go/src/github.com/yincongcyincong/MuseBot/admin/shell/certs/server.key
```

在管理页面添加配置：
![image](https://github.com/user-attachments/assets/2a518841-abf6-4a31-b1b3-b26b258a5fab)

可以使用此[文件](https://github.com/yincongcyincong/MuseBot/blob/main/admin/shell/generate_cert.sh)生成
ca、key 和 crt 文件。

#### 机器人启动参数

![image](https://github.com/user-attachments/assets/94c65d03-e097-479e-bf2a-f3d5aad431cc)
显示启动 Telegram DeepSeek Bot 时所有参数。

#### 机器人配置

![image](https://github.com/user-attachments/assets/0e6d3c32-5311-4769-ac42-e9591d4651ad)
修改您的机器人配置。

### 机器人用户

![image](https://github.com/user-attachments/assets/5534971a-e1e2-42d1-9552-0ce37b18444f)
查看和管理所有与您的机器人交互的用户。

### 为用户添加 Token

![image](https://github.com/user-attachments/assets/b9ffc006-764c-46b7-a5ce-703b052c5368)
为特定用户分配和管理 API token，以控制他们对机器人的访问和使用限制。

### 聊天记录页面

![image](https://github.com/user-attachments/assets/7b0a834f-0e62-4bec-9d57-1be22da0828d)
此页面显示机器人与用户之间的完整聊天记录，便于跟踪和分析对话。
