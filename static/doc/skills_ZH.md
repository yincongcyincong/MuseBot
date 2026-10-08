# Skill

MuseBot 支持在统一的 `/task` 工作流中使用本地 Skill。

## 定义

一个 Skill 是一个目录，入口文件为 `SKILL.md`。文件名大小写不敏感，但推荐使用 `SKILL.md`。

```markdown
---
name: release_notes
description: 根据变更描述编写简洁的发布说明。
---
# 发布说明

1. 总结本次发布。
2. 列出面向用户的变更。
3. 必要时补充升级注意事项。
```

- `name`：可选，默认使用目录名。仅允许字母、数字、中划线和下划线，最长 64 位。
- `description`：必填。`/task` 规划器会根据它判断何时选择该 Skill。
- Markdown 正文：Skill 被选中后会注入给执行智能体。

Skill 目录下的可读文本辅助文件会按相对路径一起注入。单个文件最大 256 KiB，辅助文件总量最大 2 MiB。二进制文件只会列出路径，不会注入内容。

## 配置

默认递归加载 `conf/skills`：

```bash
./MuseBot -skill_path=/path/to/skills
```

或者：

```bash
export SKILLS_PATH=/path/to/skills
```

同步 MCP 配置时会重新加载 Skill。Skill 名称不能和已有 MCP 任务智能体名称冲突。
