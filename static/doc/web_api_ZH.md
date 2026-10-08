## 📌 1. 添加用户 Token

* **接口地址**：`POST /user/token/add`
* **功能说明**：添加用户可用的 Token 数量。
* **请求参数**（JSON Body）：

```json
{
  "user_id": "string",
  "token": 100
}
```

| 参数名     | 类型     | 必填 | 说明            |
|---------|--------|----|---------------|
| user_id | string | 是  | 用户唯一标识        |
| token   | int    | 是  | 要增加的 Token 数量 |

* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

## 📌 2. 获取用户列表

* **接口地址**：`GET /user/list`
* **功能说明**：分页获取用户信息列表，可根据 `user_id` 过滤。
* **请求参数**（Query）：

| 参数名       | 类型     | 必填 | 说明          |
|-----------|--------|----|-------------|
| page      | int    | 否  | 页码（默认 1）    |
| page_size | int    | 否  | 每页数量（默认 10） |
| user_id   | string | 否  | 根据用户ID过滤    |

* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": "user123",
        "mode": "default",
        "token": 100,
        "updatetime": 1623456789,
        "avail_token": 50
      }
    ],
    "total": 1
  }
}
```

---

## 📌 3. 更新用户模式

* **接口地址**：`POST /user/update/mode`
* **功能说明**：修改用户的模型（mode）。
* **请求参数**（Form 表单）：

| 参数名     | 类型     | 必填 | 说明      |
|---------|--------|----|---------|
| user_id | string | 是  | 用户唯一标识  |
| mode    | string | 是  | 要设置的新模式 |

* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

## 📌 4. 获取用户记录

* **接口地址**：`GET /record/list`
* **功能说明**：分页获取用户对话记录，可根据是否删除及用户 ID 过滤。
* **请求参数**（Query）：

| 参数名       | 类型     | 必填 | 说明                   |
|-----------|--------|----|----------------------|
| page      | int    | 否  | 页码（默认 1）             |
| pageSize  | int    | 否  | 每页数量（默认 10）          |
| isDeleted | int    | 否  | 是否删除（0：未删，1：已删，默认全部） |
| user_id   | string | 否  | 用户 ID                |

* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": "user123",
        "question": "What's AI?",
        "answer": "Artificial Intelligence...",
        "content": "conversation content",
        "token": 50,
        "is_deleted": 0,
        "create_time": 1623456789,
        "record_type": 1
      }
    ],
    "total": 1
  }
}
```

---

## 📄 数据结构说明

### ✅ User 对象字段说明

| 字段名         | 类型     | 说明          |
|-------------|--------|-------------|
| id          | int64  | 用户主键 ID     |
| user_id     | string | 用户唯一标识      |
| mode        | string | 当前模式        |
| token       | int    | 总 Token     |
| updatetime  | int64  | 更新时间（时间戳）   |
| avail_token | int    | 可用 Token 数量 |

---

### ✅ Record 对象字段说明

| 字段名         | 类型     | 说明                 |
|-------------|--------|--------------------|
| id          | int    | 记录 ID              |
| user_id     | string | 所属用户 ID            |
| question    | string | 用户提问内容             |
| answer      | string | 系统回答               |
| content     | string | 用户上传的特殊数据，如 图片 语音等 |
| token       | int    | 消耗的 token 数量       |
| is_deleted  | int    | 是否删除（0=否，1=是）      |
| create_time | int64  | 创建时间（时间戳）          |
| record_type | int    | 记录类型（如 WEB、其他类型）   |

---

## 5. 实时通信接口 — `Communicate`

* **接口地址**：`POST /communicate`
* **功能说明**：通过 Server-Sent Events (SSE) 实时处理客户端请求，支持文本对话、图片/视频生成及多代理任务等功能，支持多种命令。

---

### 请求说明

* **请求方法**：`POST`
* **请求头**：

    * `Content-Type`: 依据请求体内容，通常为 `application/octet-stream`（图片或视频二进制数据）
* **请求参数（Query）**：

| 参数名     | 类型     | 必填 | 说明                        |
|---------|--------|----|---------------------------|
| prompt  | string | 是  | 请求内容，可包含命令（以 `/` 开头）或普通文本 |
| user_id | string | 是  | 用户唯一标识（数字字符串）             |

* **请求体**：

    * 传入二进制数据，如图片或音频数据（根据命令决定是否需要）。

---

### 命令说明

接口支持的命令如下（均以 `/` 开头）：

| 命令         | 功能说明              |
|------------|-------------------|
| `/chat`    | 开启普通聊天会话          |
| `/mode`    | 设置 LLM 模式         |
| `/state`   | 查看当前会话状态和设置       |
| `/clear`   | 清除所有对话历史          |
| `/retry`   | 重试上一次提问           |
| `/photo`   | 根据提示或上传图片生成图像     |
| `/video`   | 根据提示生成视频          |
| `/task`    | 让多个代理（包括 MCP 和 Skill 代理）协作完成任务 |
| `/help`    | 显示帮助信息（本命令列表）     |

#### /chat
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/be9043ff-513b-4cb3-a8c5-53678ada3fc7" />

#### /mode
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/5a2cead9-5064-41f9-bfab-335efc83e360" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/1a135dbb-2367-4ce0-836e-fe367c0e0ea5" />

#### /state
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/c85224b9-ed70-4c24-bc30-1c3c57174670" />


#### /clear
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/40b7ce66-6a58-4367-800e-9c909658f4ea" />

#### /retry
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/71c3611f-9087-4e76-9502-b928e4af3137" />


#### /photo
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/14424e36-169c-41c6-a58c-63f3625fd0a3" />


#### /video
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/e07a1ce3-2dae-44a9-b7ba-804649f24f05" />


#### /task
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f58adc7c-4436-4908-baf9-0a7aed8b140c" />

#### /help
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f2734a79-9d82-4716-8916-86a01865ed97" />



---

### 响应说明

* **响应类型**：`text/event-stream`

* **响应头**：

    * `Content-Type: text/event-stream`
    * `Cache-Control: no-cache`
    * `Connection: keep-alive`

* **响应体**：

    * 实时推送的事件流数据，格式由后端业务逻辑定义。

* **错误响应**：

| 状态码 | 说明              | 响应示例文本                                              |
|-----|-----------------|-----------------------------------------------------|
| 400 | 缺少必要参数 `prompt` | Missing prompt parameter                            |
| 500 | 请求体读取失败或不支持流    | Error reading request body 或 Streaming unsupported! |

---

### 示例请求

```http
POST /api/communicate?prompt=/photo sunset&user_id=12345 HTTP/1.1
Content-Type: application/octet-stream

<二进制图片数据>
```

---
## 6. 获取当前启动参数命令行信息

* **接口地址**：`GET /command/get`
* **功能说明**：返回当前服务各配置结构体字段与启动参数 flag 不一致的部分，以命令行参数格式拼接，便于调试或配置检查。
* **请求参数**：无
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": "-mcp_conf_path=/path/to/mcp_conf.json -some_flag=value "
}
```

---

## 2. 获取完整当前配置

* **接口地址**：`GET /conf/get`
* **功能说明**：返回所有配置模块当前的完整配置信息，包括 base、audio、llm、photo、rag、video。
* **请求参数**：无
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "base": { ... },
    "audio": { ... },
    "llm": { ... },
    "photo": { ... },
    "rag": { ... },
    "video": { ... }
  }
}
```

---

## 3. 更新配置字段

* **接口地址**：`POST /conf/update`
* **功能说明**：动态更新指定类型配置结构体的指定字段值。
* **请求参数**（JSON Body）：

```json
{
  "type": "base|audio|llm|photo|rag|video",
  "key": "json_tag字段名",
  "value": "更新后的值"
}
```

| 参数名   | 类型     | 必填 | 说明              |
| ----- | ------ | -- | --------------- |
| type  | string | 是  | 配置类型，如 `"base"` |
| key   | string | 是  | 结构体字段的 json tag |
| value | any    | 是  | 需要更新的值          |

* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

* **注意**：

    * 对于特殊字段（如 `allowed_user_ids`、`admin_user_ids` 等）会做特殊格式转换。
    * 不支持的类型会返回参数错误。

---

## 4. 获取 MCP 配置

* **接口地址**：`GET /mcp/get`
* **功能说明**：读取并返回 MCP 配置文件的内容。
* **请求参数**：无
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "McpServers": {
      "server1": { ... },
      "server2": { ... }
    },
    ...
  }
}
```

---

## 5. 更新 MCP 配置

* **接口地址**：`POST /mcp/update?name={name}`
* **功能说明**：更新 MCP 配置文件中指定服务器名称的配置。
* **请求参数**：

    * Query参数：

        * `name` (string, 必填)：MCP服务器名称
    * JSON Body：MCP 配置对象，结构体 `mcpParam.MCPConfig`
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 6. 删除 MCP 配置

* **接口地址**：`DELETE /mcp/delete?name={name}`
* **功能说明**：删除 MCP 配置文件中指定服务器名称的配置，同时关闭对应客户端并从任务工具中删除。
* **请求参数**：

    * Query参数：

        * `name` (string, 必填)：MCP服务器名称
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 7. 启用或禁用 MCP 配置

* **接口地址**：`POST /mcp/disable?name={name}&disable={0|1}`
* **功能说明**：启用或禁用指定名称的 MCP 配置。
* **请求参数**：

    * Query参数：

        * `name` (string, 必填)：MCP服务器名称
        * `disable` (string, 必填)：`1` 表示禁用，`0` 表示启用
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 8. 同步 MCP 配置

* **接口地址**：`POST /mcp/sync`
* **功能说明**：清除所有 MCP 客户端及任务工具，重新初始化。
* **请求参数**：无
* **响应示例**：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

# 备注

* 所有成功响应格式：

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

* 失败响应格式类似，但 `code` 和 `msg` 会体现具体错误，通常为：

```json
{
  "code": <错误码>,
  "msg": <错误信息>,
  "data": null
}
```


