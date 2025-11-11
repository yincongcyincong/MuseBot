# ⏰ Scheduled Task Push Platform Configuration Document

This document details the key identifiers and support status required for configuring scheduled tasks to actively push messages across different messaging platforms.

## 🚀 Platforms Supporting Active Push and Configuration Methods

The following platforms support actively pushing messages via scheduled tasks, with the required configuration identifiers listed:

| Platform | Personal Message Configuration Item | Group/Channel Message Configuration Item | Notes |
| :--- | :--- | :--- | :--- |
| **Telegram** | Fill in **`Personal userId`** | Fill in **`Group`** Identifier (usually ChatID) | |
| **WeChat (微信)** | Fill in **`Personal userId`** | **Group Push Not Supported** | |
| **Personal QQ (个人QQ)** | Fill in **`Personal QQ Number`** | Fill in **`Group Number`** | |
| **Lark (飞书)** | Fill in **`chatId`** | Requires **`chatId`** | Lark usually uses ChatID as the session identifier. |
| **Slack** | Fill in **`chatId`** | Requires **`chatId`** | |
| **Com WeChat (企业微信)**| Fill in **`Personal userId`** | **Group Push Not Supported** | |

---

## 🚫 Platforms Not Supporting Active Push

The following platforms **do not support** actively pushing messages via scheduled tasks, or have specific limitations:

| Platform | Reason for Non-Support / Limitation Description |
| :--- | :--- |
| **QQ** | **Does not support** the function of active message pushing. |
| **Discord** | **Does not support** active message sending. Replies only occur after a user actively sends a message. |
| **DingTalk (钉钉)** | **Does not support** active push (often handled via callback/webhooks only). |

---

## 📝 Key Configuration Item Descriptions

* **`Personal userId` / `Personal QQ Number`**: Used to uniquely identify a **personal user** account or session.
* **`Group` / `Group Number`**: Used to uniquely identify a **group** or group chat.
* **`chatId`**: Typically the ID used by platforms like Lark and Slack to uniquely identify a **conversation/channel**.

Please ensure you correctly fill in the corresponding identifiers based on the platform you are using and the target recipient (individual or group).

