# 对话式配置

配置请求现在统一在 `/task` 工作流里执行。管理员直接发给 `/task`，或者开启智能模式（`SMART_MODE=true`）后用自然语言表达——分类器会把系统请求自动路由到 `/task`：

```text
/task 开启智能模式。
/task 添加一个叫 weather 的 MCP，地址是 https://example.com/mcp，描述是查询天气数据。
/task 重新加载 skill。
/task 创建一个叫 release_notes 的 skill，用于编写发布说明，指令：输出简洁、面向用户的发布说明。
/task 每天早上 8 点创建一个定时任务，提示词是“总结 AI 新闻”。
/task 运行 git status。
/task 查看 data 目录。
```

在 `/task` 内部，规划器会分派内建的 `bot_config` agent，由它调用 `bot_config` 工具提交经过校验的动作；MuseBot 在本地执行后把结果并入任务输出。

## 权限

使用前先设置管理员 ID：

```bash
export ADMIN_USER_IDS=123456,654321
./MuseBot -admin_user_ids=123456,654321
```

配置规划器拒绝空管理员列表；查询敏感值时会脱敏。机器人名称、HTTP 监听地址、数据库类型和连接信息属于进程启动配置，不能在运行中修改。

## 内置命令与文件

命令只使用 argv 数组执行，不解析 shell 字符串或 shell 操作符，且必须事先配置白名单：

```bash
export ALLOWED_COMMANDS=git,ls
./MuseBot -allowed_commands=git,ls -command_timeout_sec=60
```

* `-allowed_commands`：逗号分隔的可执行文件白名单，等价环境变量是 `ALLOWED_COMMANDS`。
* `-command_timeout_sec`：单次命令最长执行时间，等价环境变量是 `COMMAND_TIMEOUT_SEC`。
* 命令在文件根目录内执行；也可以通过自然语言指定根目录内的已存在工作目录。
* 命令白名单属于启动配置，不能通过对话先放行再执行。

文件操作被限制在根目录内，符号链接解析后也必须留在根目录内：

```bash
export FILE_ROOT_PATH=/path/to/agent-files
./MuseBot -file_root_path=/path/to/agent-files
```

文件根目录属于启动配置，不能通过对话在运行中修改。

* 单次读取最多 256 KiB，写入和编辑最多 1 MiB。
* 目录列表最多返回 500 项，命令输出最多返回 16 KiB。
* 写入和编辑使用临时文件加原子重命名。

## 支持范围

- 常规配置：列出字段、查询值、修改支持的运行时字段。
- MCP：列出、添加/更新、启用、禁用、删除和同步。
- Skill：列出、创建、重载和修改 Skill 目录。
- 定时任务：列出、查询、创建、更新、启用、禁用、删除和清空。
- 命令：运行白名单内命令，并返回截断后的输出。
- 文件：列出、读取、写入、精确替换编辑、创建目录、删除和移动。

配置动作是 `/task` 多智能体工作流的一部分：任务规划器能看到 `bot_config` agent，被分派的子 agent 生成受限的 JSON 动作，MuseBot 在本地校验并执行，然后返回结果。
