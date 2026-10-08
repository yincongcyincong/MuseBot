# Skills

MuseBot supports local prompt skills in the unified `/task` workflow.

## Definition

A skill is a directory containing a `SKILL.md` entry file. The filename is case-insensitive, though `SKILL.md` is recommended.

```markdown
---
name: release_notes
description: Write concise release notes from a change description.
---
# Release Notes

1. Summarize the release.
2. List user-facing changes.
3. Add upgrade notes when needed.
```

- `name`: optional if the directory name is a valid skill name. Allowed characters are letters, digits, hyphens, and underscores; maximum length is 64.
- `description`: required. The `/task` planner uses it to decide when the skill applies.
- Markdown body: operating instructions injected when the planner selects this skill.

Supporting text files under the skill directory are also injected with their relative paths. A single file is limited to 256 KiB, and total supporting resources are limited to 2 MiB. Binary files are listed but not loaded.

## Configuration

By default, skills are loaded recursively from `conf/skills`.

```bash
./MuseBot -skill_path=/path/to/skills
```

Or:

```bash
export SKILLS_PATH=/path/to/skills
```

Skills are reloaded when MCP configuration is synchronized. A skill name cannot conflict with an existing MCP task-agent name.
