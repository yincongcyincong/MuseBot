# 🤖 OneBot + MuseBot 环境变量配置说明

## 🧩 一、启动 OneBot 容器

启动 LLONEBot
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

### 说明：
* `3000`: LLONEBot HTTP 服务端口（用于接收 MuseBot 发来的消息）
* `3080`: LLONEBot Web 管理界面（扫码登录 QQ）

或启动 NapCat：

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
````

### 说明：
* `3000`: NapCat HTTP 服务端口（用于接收 MuseBot 发来的消息）
* `6099`: NapCat Web 管理界面（扫码登录 QQ）

---

## 🔐 二、登录 QQ

1. 打开浏览器访问 http://127.0.0.1:3080/ 或者 http://127.0.0.1:6099/ 
2. 扫码登录 QQ 账号
3. 两个软件都会有个key，妥善保存

---

## ⚙️ 三、MuseBot 环境变量配置

在 MuseBot 启动前，需要配置 OneBot 相关的 3 个环境变量：

| 环境变量名                     | 说明                                                | 示例值                     |
|---------------------------|---------------------------------------------------|-------------------------|
| `QQ_ONEBOT_RECEIVE_TOKEN` | OneBot 向 MuseBot 发送消息时使用的 token（OneBot 客户端 token） | —               |
| `QQ_ONEBOT_SEND_TOKEN`    | MuseBot 向 OneBot 发送消息时使用的 token（OneBot 服务器 token） | `MuseBot`               |
| `QQ_ONEBOT_HTTP_SERVER`   | OneBot HTTP 服务器地址（即 OneBot 接收消息的接口）               | `http://127.0.0.1:3000` |

例如：

```bash
export QQ_ONEBOT_RECEIVE_TOKEN="<your-unique-random-token>"
export QQ_ONEBOT_SEND_TOKEN=MuseBot
export QQ_ONEBOT_HTTP_SERVER=http://127.0.0.1:3000
```

> ⚠️ 这三个变量必须与 OneBot 配置页面中的设置保持一致。

---

## 🔄 四、OneBot 网络配置说明

登录 OneBot Web 控制台 → 「配置」 → 「网络配置」，按以下方式填写：

| 配置项             | 说明                         | 示例值                             |
|-----------------|----------------------------|---------------------------------|
| **HTTP 服务器**    | MuseBot 调用 Onebot 接口的地址    | `http://127.0.0.1:3000`         |
| **HTTP 客户端**    | Onebot 向 MuseBot 推送消息事件的地址 | `http://127.0.0.1:36060/onebot` |
| **HTTP 鉴权 Key** | 与环境变量中 token 一致            | `MuseBot`                       |

![image](https://github.com/user-attachments/assets/a6a7bf64-9f93-436f-8910-1177e1e2749a)
![image](https://github.com/user-attachments/assets/13a118a7-ced0-4427-923d-44cc0df7ca2c)
![image](https://github.com/user-attachments/assets/b6aa893d-6db9-444a-82e6-a185561ad818)
![image](https://github.com/user-attachments/assets/53e86994-a19d-487b-b46f-3b457a38d5c0)




请在两端配置同一个随机接收令牌。媒体地址须使用公网 HTTP(S)，本地文件和私有地址会被拒绝。
