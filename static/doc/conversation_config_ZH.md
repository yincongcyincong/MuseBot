# 对话式配置

管理员可以直接用自然语言配置 MuseBot：

```text
开启智能模式。
添加一个叫 weather 的 MCP，地址是 https://example.com/mcp，描述是查询天气数据。
重新加载 skill。
创建一个叫 release_notes 的 skill，用于编写发布说明，指令：输出简洁、面向用户的发布说明。
每天早上 8 点创建一个定时任务，提示词是“总结 AI 新闻”。
运行 git status。
查看 data 目录。
读取 notes/a.md。
把 doc.md 里的 old 改成 new。
把 a.md 移动到 archive/a.md。
```

MuseBot 会先从普通消息中识别系统意图；确认是系统请求后直接执行，其他消息继续走正常聊天和 Smart Mode 流程。咨询概念、假设问题和风险讨论不会触发执行。

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

意图识别使用轻量 LLM 分类器，确认后再生成受限制的 JSON 操作列表；MuseBot 在本地校验并执行这些操作，然后逐项返回结果。
