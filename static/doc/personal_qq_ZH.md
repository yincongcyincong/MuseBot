# 🤖 NapCat + MuseBot 环境变量配置说明

本说明介绍如何在 Docker 环境中启动 NapCat 并配置 MuseBot 与其进行通信。

---

## 🧩 一、启动 NapCat 容器

运行以下命令启动 NapCat：

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
* `3001`: NapCat WebSocket 端口（可忽略）
* `6099`: NapCat Web 管理界面（扫码登录 QQ）
* `NAPCAT_GID` 和 `NAPCAT_UID`：保持宿主机用户权限一致，避免文件权限问题

---

## 🔐 二、登录 QQ

1. 打开浏览器访问 [http://127.0.0.1:6099/](http://127.0.0.1:6099/)
2. 扫码登录 QQ 账号
3. 登录成功后查看容器日志：

   ```bash
   docker logs -f napcat
   ```
4. 日志中会出现：

   ```
    [NapCat] [WebUi] 🔑 token=bcc53c876d56
   ```

   这个 **Key** 是 NapCat 的 HTTP 鉴权密钥，请妥善保存。

---

## ⚙️ 三、MuseBot 环境变量配置

在 MuseBot 启动前，需要配置 NapCat 相关的 3 个环境变量：

| 环境变量名                     | 说明                                                | 示例值                     |
|---------------------------|---------------------------------------------------|-------------------------|
| `QQ_NAPCAT_RECEIVE_TOKEN` | NapCat 向 MuseBot 发送消息时使用的 token（NapCat 客户端 token） | `MuseBot`               |
| `QQ_NAPCAT_SEND_TOKEN`    | MuseBot 向 NapCat 发送消息时使用的 token（NapCat 服务器 token） | `MuseBot`               |
| `QQ_NAPCAT_HTTP_SERVER`   | NapCat HTTP 服务器地址（即 NapCat 接收消息的接口）               | `http://127.0.0.1:3000` |

例如：

```bash
export QQ_NAPCAT_RECEIVE_TOKEN=MuseBot
export QQ_NAPCAT_SEND_TOKEN=MuseBot
export QQ_NAPCAT_HTTP_SERVER=http://127.0.0.1:3000
```

> ⚠️ 这三个变量必须与 NapCat 配置页面中的设置保持一致。

---

## 🔄 四、NapCat 网络配置说明

登录 NapCat Web 控制台 → 「配置」 → 「网络配置」，按以下方式填写：

| 配置项             | 说明                         | 示例值                             |
|-----------------|----------------------------|---------------------------------|
| **HTTP 服务器**    | MuseBot 调用 NapCat 接口的地址    | `http://127.0.0.1:3000`         |
| **HTTP 客户端**    | NapCat 向 MuseBot 推送消息事件的地址 | `http://127.0.0.1:36060/napcat` |
| **HTTP 鉴权 Key** | 与环境变量中 token 一致            | `MuseBot`                       |

![image](https://github.com/user-attachments/assets/b6aa893d-6db9-444a-82e6-a185561ad818)
![image](https://github.com/user-attachments/assets/53e86994-a19d-487b-b46f-3b457a38d5c0)

---

## 🧠 五、完整运行流程

1. 启动 NapCat Docker 容器
2. 访问 [http://127.0.0.1:6099/](http://127.0.0.1:6099/) 扫码登录 QQ
3. 获取 NapCat 日志中的 HTTP Key
4. 在 NapCat 控制台中配置 HTTP 网络：

    * HTTP 服务器：`http://127.0.0.1:3000`
    * HTTP 客户端：`http://127.0.0.1:36060/napcat`
    * 鉴权 Key：`MuseBot`
5. 在 MuseBot 环境中设置环境变量：

   ```bash
   export QQ_NAPCAT_RECEIVE_TOKEN=MuseBot
   export QQ_NAPCAT_SEND_TOKEN=MuseBot
   export QQ_NAPCAT_HTTP_SERVER=http://127.0.0.1:3000
   ```
6. 启动 MuseBot
7. NapCat 会自动将 QQ 消息推送给 MuseBot，MuseBot 可以通过 NapCat HTTP 接口发送回复。

