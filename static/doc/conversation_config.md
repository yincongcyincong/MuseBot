# Conversation Configuration

Configuration requests run inside the unified `/task` workflow. Administrators
send the request to `/task` directly, or enable smart mode (`SMART_MODE=true`)
and phrase it naturally—the classifier routes system requests to `/task`
automatically:

```text
/task Enable smart mode.
/task Add an MCP server named weather, URL https://example.com/mcp, description "Query weather data".
/task Reload skills.
/task Create a skill named release_notes for writing release notes; instructions: write concise user-facing release notes.
/task Create a cron at 08:00 every day with prompt "summarize AI news".
/task Run git status.
```

Inside `/task`, the planner assigns the built-in `bot_config` agent, which
calls the `bot_config` tool with a validated action. MuseBot executes the
action locally and reports each result as part of the task output.

## Access Control

Set admin IDs before using this command:

```bash
export ADMIN_USER_IDS=123456,654321
./MuseBot -admin_user_ids=123456,654321
```

The planner rejects empty admin lists and redacts sensitive values when queried. Bot name, HTTP host, database type, database connection, the command allowlist, and the file root cannot be changed at runtime because they define the process security boundary.

## Built-in Commands and Files

Commands are executed as argv arrays only—no shell strings or shell operators are parsed—and must be allowlisted before startup:

```bash
export ALLOWED_COMMANDS=git,ls
./MuseBot -allowed_commands=git,ls -command_timeout_sec=60
```

- `-allowed_commands`: comma-separated executable allowlist; equivalent environment variable is `ALLOWED_COMMANDS`.
- `-command_timeout_sec`: maximum command runtime (default 600 seconds); equivalent environment variable is `COMMAND_TIMEOUT_SEC`.
- Commands run inside the file root. A natural-language request may select an existing working directory inside that root.
- A command outside the allowlist triggers an interactive approval: the user can add it to the allowlist and run it right away. Approvals are persisted to `data/runtime_boundary.json` and survive restarts.
- `-task_auto_approve`: auto approve every task tool operation (command runs and runtime config changes) without the interactive `/y` confirmation; equivalent environment variable is `TASK_AUTO_APPROVE` (default `false`).

There is no separate file tool: all file operations (list, read, write, edit,
create, move) go through allowlisted shell commands (`ls`, `cat`, `tee`,
`mkdir`, `mv`, ...). Commands run with a working directory inside the
configured file root; a path outside the root used as a working directory is
rejected. Approving a new command persists it to `data/runtime_boundary.json`.

## Supported Operations

- General config: list fields, get a value, and set a supported runtime field.
- MCP: list, add/update, enable, disable, delete, and synchronize servers.
- Skills: list, create, reload, and change the skill directory.
- Cron: list, get, create, update, enable, disable, delete, and clear.
- Commands: run an allowlisted command (including all file operations) and return truncated output.

Configuration actions are part of the `/task` multi-agent workflow: the task
planner sees the `bot_config` agent, the assigned sub-agent emits a constrained
JSON action, and MuseBot validates and executes it locally before reporting
the result.
