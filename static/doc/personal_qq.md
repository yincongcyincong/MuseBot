# 🤖 NapCat + MuseBot Environment Variable Configuration Guide

This guide explains how to launch NapCat in a Docker environment and configure MuseBot to communicate with it.

---

## 🧩 1. Start the NapCat Container

Run the following command to start NapCat:

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

### Explanation:

* `3000`: NapCat HTTP server port (used to receive messages from MuseBot)
* `3001`: NapCat WebSocket port (optional)
* `6099`: NapCat Web admin interface (for QQ login via QR code)
* `NAPCAT_GID` & `NAPCAT_UID`: Keep permissions consistent with the host user to avoid file access issues

---

## 🔐 2. Log in to QQ

1. Open your browser and visit [http://127.0.0.1:6099/](http://127.0.0.1:6099/)
2. Scan the QR code to log in with your QQ account
3. After logging in, check the container logs:

   ```bash
   docker logs -f napcat
   ```
4. You’ll find a line like this:

   ```
   [NapCat] [WebUi] 🔑 token=bcc53c876d56
   ```

   This **Key** is NapCat’s HTTP authentication token — save it carefully.

---

## ⚙️ 3. Configure MuseBot Environment Variables

Before starting MuseBot, set the following environment variables related to NapCat:

| Environment Variable Name | Description                                                            | Example Value           |
|---------------------------|------------------------------------------------------------------------|-------------------------|
| `QQ_NAPCAT_RECEIVE_TOKEN` | Token used when NapCat sends messages to MuseBot (NapCat client token) | `MuseBot`               |
| `QQ_NAPCAT_SEND_TOKEN`    | Token used when MuseBot sends messages to NapCat (NapCat server token) | `MuseBot`               |
| `QQ_NAPCAT_HTTP_SERVER`   | NapCat HTTP server address (the endpoint for receiving messages)       | `http://127.0.0.1:3000` |

Example setup:

```bash
export QQ_NAPCAT_RECEIVE_TOKEN=MuseBot
export QQ_NAPCAT_SEND_TOKEN=MuseBot
export QQ_NAPCAT_HTTP_SERVER=http://127.0.0.1:3000
```

> ⚠️ These variables **must match** the configuration in NapCat’s web interface.

---

## 🔄 4. NapCat Network Configuration

In the NapCat Web Console → “Settings” → “Network Configuration”, fill in the following fields:

| Configuration     | Description                                        | Example                         |
|-------------------|----------------------------------------------------|---------------------------------|
| **HTTP Server**   | Address MuseBot uses to call NapCat APIs           | `http://127.0.0.1:3000`         |
| **HTTP Client**   | Address NapCat uses to push events to MuseBot      | `http://127.0.0.1:36060/napcat` |
| **HTTP Auth Key** | Must match the token in your environment variables | `MuseBot`                       |

![image](https://github.com/user-attachments/assets/b6aa893d-6db9-444a-82e6-a185561ad818)
![image](https://github.com/user-attachments/assets/53e86994-a19d-487b-b46f-3b457a38d5c0)


---

## 🧠 5. Full Workflow

1. Start the NapCat Docker container
2. Open [http://127.0.0.1:6099/](http://127.0.0.1:6099/) and log in to QQ
3. Retrieve the NapCat HTTP Key from the logs
4. Configure NapCat network settings:

    * HTTP Server: `http://127.0.0.1:3000`
    * HTTP Client: `http://127.0.0.1:36060/napcat`
    * Auth Key: `MuseBot`
5. Set MuseBot environment variables:

   ```bash
   export QQ_NAPCAT_RECEIVE_TOKEN=MuseBot
   export QQ_NAPCAT_SEND_TOKEN=MuseBot
   export QQ_NAPCAT_HTTP_SERVER=http://127.0.0.1:3000
   ```
6. Start MuseBot
7. NapCat will automatically forward QQ messages to MuseBot, and MuseBot can reply via the NapCat HTTP API.
