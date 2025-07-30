## 📌 1. Add User Token

* **Endpoint**: `POST /user/token/add`
* **Description**: Add available tokens for a user.
* **Request Body** (JSON):

```json
{
  "user_id": "string",
  "token": 100
}
```

| Parameter | Type   | Required | Description                    |
| --------- | ------ | -------- | ------------------------------ |
| user\_id  | string | Yes      | Unique identifier for the user |
| token     | int    | Yes      | Number of tokens to add        |

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

## 📌 2. Get User List

* **Endpoint**: `GET /user/list`
* **Description**: Get a paginated list of users, optionally filtered by `user_id`.
* **Query Parameters**:

| Parameter  | Type   | Required | Description                           |
| ---------- | ------ | -------- | ------------------------------------- |
| page       | int    | No       | Page number (default 1)               |
| page\_size | int    | No       | Number of items per page (default 10) |
| user\_id   | string | No       | Filter by user ID                     |

* **Response Example**:

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

## 📌 3. Update User Mode

* **Endpoint**: `POST /user/update/mode`
* **Description**: Update the mode (model) for a user.
* **Request Parameters** (Form):

| Parameter | Type   | Required | Description            |
| --------- | ------ | -------- | ---------------------- |
| user\_id  | string | Yes      | Unique user identifier |
| mode      | string | Yes      | New mode to set        |

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

## 📌 4. Get User Records

* **Endpoint**: `GET /record/list`
* **Description**: Retrieve paginated user conversation records, with optional filters for deletion status and user ID.
* **Query Parameters**:

| Parameter | Type   | Required | Description                                                           |
| --------- | ------ | -------- | --------------------------------------------------------------------- |
| page      | int    | No       | Page number (default 1)                                               |
| pageSize  | int    | No       | Items per page (default 10)                                           |
| isDeleted | int    | No       | Filter by deletion status (0 = not deleted, 1 = deleted, default all) |
| user\_id  | string | No       | User ID filter                                                        |

* **Response Example**:

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

## 📄 Data Structure Definitions

### ✅ User Object Fields

| Field        | Type   | Description            |
| ------------ | ------ | ---------------------- |
| id           | int64  | User primary key ID    |
| user\_id     | string | Unique user identifier |
| mode         | string | Current mode           |
| token        | int    | Total tokens           |
| updatetime   | int64  | Update timestamp       |
| avail\_token | int    | Available tokens       |

---

### ✅ Record Object Fields

| Field        | Type   | Description                                   |
| ------------ | ------ | --------------------------------------------- |
| id           | int    | Record ID                                     |
| user\_id     | string | Associated user ID                            |
| question     | string | User question content                         |
| answer       | string | System response                               |
| content      | string | Uploaded special content (e.g., image, audio) |
| token        | int    | Tokens consumed                               |
| is\_deleted  | int    | Deletion status (0 = no, 1 = yes)             |
| create\_time | int64  | Creation timestamp                            |
| record\_type | int    | Record type (e.g., WEB or other)              |

---

## 5. Real-time Communication API — `Communicate`

* **Endpoint**: `POST /communicate`

* **Description**: Real-time client request handling via Server-Sent Events (SSE), supporting text chat, image/video generation, multi-agent tasks, and various commands.

* **Request Method**: `POST`

* **Request Headers**:

    * `Content-Type`: Usually `application/octet-stream` (binary image/video data)

* **Query Parameters**:

| Parameter | Type   | Required | Description                                                           |
| --------- | ------ | -------- | --------------------------------------------------------------------- |
| prompt    | string | Yes      | Request content, can include commands starting with `/` or plain text |
| user_id  | string | Yes      | User unique identifier (numeric string)                               |

* **Request Body**:

    * Binary data such as image or audio, depending on command.

---

### Supported Commands

| Command    | Description                                             |
| ---------- | ------------------------------------------------------- |
| `/chat`    | Start a normal chat session                             |
| `/mode`    | Set the LLM mode                                        |
| `/balance` | Check current balance (tokens or credits)               |
| `/state`   | View current session state and settings                 |
| `/clear`   | Clear all conversation history                          |
| `/retry`   | Retry last question                                     |
| `/photo`   | Generate image based on prompt or uploaded image        |
| `/video`   | Generate video based on prompt                          |
| `/task`    | Let multiple agents collaborate on a task               |
| `/mcp`     | Use multi-agent control panel for complex task planning |
| `/help`    | Show this help message (list of commands)               |

#### /chat
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/be9043ff-513b-4cb3-a8c5-53678ada3fc7" />

#### /mode
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/5a2cead9-5064-41f9-bfab-335efc83e360" />
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/1a135dbb-2367-4ce0-836e-fe367c0e0ea5" />

#### /balance
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/b7dd73f6-adef-4367-90ea-96d0cb5ba692" />

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

#### /mcp
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/9c5db063-23b5-41c2-989c-4eda48b7440c" />

#### /help
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/f2734a79-9d82-4716-8916-86a01865ed97" />



---

### Response

* **Content-Type**: `text/event-stream`

* **Headers**:

    * `Cache-Control: no-cache`
    * `Connection: keep-alive`

* **Body**: Server-sent event stream data pushed in real-time.

* **Error Responses**:

| Status Code | Description                                        | Response Text                                        |
| ----------- | -------------------------------------------------- | ---------------------------------------------------- |
| 400         | Missing required `prompt` param                    | Missing prompt parameter                             |
| 500         | Request body read failure or unsupported streaming | Error reading request body or Streaming unsupported! |

---

### Example Request

```http
POST /api/communicate?prompt=/photo sunset&user_id=12345 HTTP/1.1
Content-Type: application/octet-stream

<binary image data>
```

---

## 6. Get Current Startup Command Line Arguments

* **Endpoint**: `GET /command/get`
* **Description**: Return the current command-line parameters that differ from the config struct defaults, formatted as CLI flags.
* **Request Parameters**: None
* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": "-mcp_conf_path=/path/to/mcp_conf.json -some_flag=value "
}
```

---

## 7. Get Full Current Configuration

* **Endpoint**: `GET /conf/get`
* **Description**: Return the full current configuration from all modules (base, audio, llm, photo, rag, video).
* **Request Parameters**: None
* **Response Example**:

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

## 8. Update Configuration Field

* **Endpoint**: `POST /conf/update`
* **Description**: Dynamically update a specified field in a specific config struct.
* **Request Body** (JSON):

```json
{
  "type": "base|audio|llm|photo|rag|video",
  "key": "json_tag_field",
  "value": "new_value"
}
```

| Parameter | Type   | Required | Description                 |
| --------- | ------ | -------- | --------------------------- |
| type      | string | Yes      | Config type, e.g., `"base"` |
| key       | string | Yes      | Struct field's JSON tag     |
| value     | any    | Yes      | New value for the field     |

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

* **Note**:

    * Special fields (e.g., `allowed_telegram_user_ids`, `admin_user_ids`) are processed specially.
    * Unsupported types return parameter error.

---

## 9. Get MCP Configuration

* **Endpoint**: `GET /mcp/get`
* **Description**: Read and return the MCP configuration file content.
* **Request Parameters**: None
* **Response Example**:

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

## 10. Update MCP Configuration

* **Endpoint**: `POST /mcp/update?name={name}`

* **Description**: Update MCP config for a given server name.

* **Request Parameters**:

    * Query:

        * `name` (string, required): MCP server name

    * JSON Body: MCP configuration object (`mcpParam.MCPConfig` struct)

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 11. Delete MCP Configuration

* **Endpoint**: `DELETE /mcp/delete?name={name}`

* **Description**: Delete MCP config by server name, close the client and remove from task tools.

* **Request Parameters**:

    * Query:

        * `name` (string, required): MCP server name

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 12. Enable or Disable MCP Configuration

* **Endpoint**: `POST /mcp/disable?name={name}&disable={0|1}`

* **Description**: Enable or disable MCP config for the specified server.

* **Request Parameters**:

    * Query:

        * `name` (string, required): MCP server name
        * `disable` (string, required): `"1"` to disable, `"0"` to enable

* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

## 13. Synchronize MCP Configuration

* **Endpoint**: `POST /mcp/sync`
* **Description**: Clear all MCP clients and task tools, then reinitialize.
* **Request Parameters**: None
* **Response Example**:

```json
{
  "code": 0,
  "msg": "success",
  "data": ""
}
```

---

# Notes

* Successful responses all follow the format:

```json
{
  "code": 0,
  "msg": "success",
  "data": <response data or empty string>
}
```

* Failure responses include a non-zero code and an error message:

```json
{
  "code": <error code>,
  "msg": <error message>,
  "data": null
}
```

