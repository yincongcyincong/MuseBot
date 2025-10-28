# 🤖 OneBot + MuseBot Environment Variable Configuration Guide

## 🧩 1. Start the OneBot Container

### Start **LLONEBot**

```bash
docker run -d \
  --name llonebot \
  -p 3000:3000 \
  -p 3001:3001 \
  -p 5600:5600 \
  -p 3080:3080 \
  -v ./QQ:/root/.config/QQ \
  -v ./llonebot:/root/llonebot \
  --add-host=host.docker.internal:host-gateway \
  --restart unless-stopped \
  initialencounter/llonebot:latest
```

**Explanation:**

* `3000`: LLONEBot HTTP service port (used to receive messages from MuseBot)
* `3080`: LLONEBot web management interface (used for QQ login via QR code)

---

### Or start **NapCat**:

```bash
docker run -d \
  -e NAPCAT_GID=$(id -g) \
  -e NAPCAT_UID=$(id -u) \
  -p 3000:3000 \
  -p 3001:3001 \
  -p 6099:6099 \
  --name napcat \
  --restart=always \
  mlikiowa/napcat-docker:latest
```

**Explanation:**

* `3000`: NapCat HTTP service port (used to receive messages from MuseBot)
* `6099`: NapCat web management interface (used for QQ login via QR code)

---

## 🔐 2. Log in to QQ

1. Open your browser and visit `http://127.0.0.1:3080/` or `http://127.0.0.1:6099/`
2. Scan the QR code to log into your QQ account
3. Both LLONEBot and NapCat will generate a **key** — keep it safe for configuration

---

## ⚙️ 3. MuseBot Environment Variable Configuration

Before starting MuseBot, set the following **three environment variables** related to OneBot:

| Variable Name             | Description                                                      | Example Value           |
| ------------------------- | ---------------------------------------------------------------- | ----------------------- |
| `QQ_ONEBOT_RECEIVE_TOKEN` | Token used by OneBot to send messages to MuseBot (client token)  | `MuseBot`               |
| `QQ_ONEBOT_SEND_TOKEN`    | Token used by MuseBot to send messages to OneBot (server token)  | `MuseBot`               |
| `QQ_ONEBOT_HTTP_SERVER`   | OneBot HTTP server address (OneBot’s message receiving endpoint) | `http://127.0.0.1:3000` |

**Example:**

```bash
export QQ_ONEBOT_RECEIVE_TOKEN=MuseBot
export QQ_ONEBOT_SEND_TOKEN=MuseBot
export QQ_ONEBOT_HTTP_SERVER=http://127.0.0.1:3000
```

> ⚠️ These three variables **must match** the configuration in the OneBot web interface.

---

## 🔄 4. OneBot Network Configuration

Go to the **OneBot Web Console → “Configuration” → “Network Settings”**, and fill in as follows:

| Setting           | Description                                       | Example Value                   |
| ----------------- | ------------------------------------------------- | ------------------------------- |
| **HTTP Server**   | The address MuseBot uses to call OneBot APIs      | `http://127.0.0.1:3000`         |
| **HTTP Client**   | The address where OneBot pushes message events    | `http://127.0.0.1:36060/onebot` |
| **HTTP Auth Key** | Must match the token set in environment variables | `MuseBot`                       |

---

![image](https://github.com/user-attachments/assets/a6a7bf64-9f93-436f-8910-1177e1e2749a)
![image](https://github.com/user-attachments/assets/13a118a7-ced0-4427-923d-44cc0df7ca2c)
![image](https://github.com/user-attachments/assets/b6aa893d-6db9-444a-82e6-a185561ad818)
![image](https://github.com/user-attachments/assets/53e86994-a19d-487b-b46f-3b457a38d5c0)

