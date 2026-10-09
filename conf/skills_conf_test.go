package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitSkills(t *testing.T) {
	skillRoot := t.TempDir()
	skillDir := filepath.Join(skillRoot, "release-notes")
	referenceDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(referenceDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillContent := `---
name: release_notes
description: Write concise release notes from a change description.
---
# Release Notes

Use a summary, changes list, and upgrade notes.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referenceDir, "style.md"), []byte("# Style\n\nUse user-facing language."), 0644); err != nil {
		t.Fatal(err)
	}

	oldSkillPath := ToolsConfInfo.SkillPath
	ToolsConfInfo.SkillPath = &skillRoot
	defer func() {
		ToolsConfInfo.SkillPath = oldSkillPath
		TaskTools.Delete("release_notes")
		SkillTools.Delete("release_notes")
	}()

	InitSkills()

	value, ok := TaskTools.Load("release_notes")
	if !ok {
		t.Fatal("skill was not loaded into TaskTools")
	}
	skill, ok := value.(*AgentInfo)
	if !ok {
		t.Fatalf("TaskTools contains %T, want *AgentInfo", value)
	}
	if !skill.IsSkill {
		t.Error("skill was not marked with IsSkill")
	}
	if skill.Description != "Write concise release notes from a change description." {
		t.Errorf("description = %q", skill.Description)
	}
	if !strings.Contains(skill.Instructions, "# Release Notes") {
		t.Errorf("instructions do not contain SKILL.md body: %q", skill.Instructions)
	}
	if !strings.Contains(skill.Instructions, "# Skill Resource: references/style.md") {
		t.Errorf("instructions do not contain supporting resource: %q", skill.Instructions)
	}

	emptySkillRoot := t.TempDir()
	ToolsConfInfo.SkillPath = &emptySkillRoot
	InitSkills()
	if _, ok := TaskTools.Load("release_notes"); ok {
		t.Error("skill was not removed after reload")
	}
}

func TestInitSkillsSupportsLowercaseSkillFilename(t *testing.T) {
	skillRoot := t.TempDir()
	skillDir := filepath.Join(skillRoot, "lowercase")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := `---
name: lowercase_skill
description: Verify lowercase skill.md files are loaded.
---
Use lowercase filenames.
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldSkillPath := ToolsConfInfo.SkillPath
	ToolsConfInfo.SkillPath = &skillRoot
	defer func() {
		ToolsConfInfo.SkillPath = oldSkillPath
		TaskTools.Delete("lowercase_skill")
		SkillTools.Delete("lowercase_skill")
	}()

	InitSkills()

	if _, ok := TaskTools.Load("lowercase_skill"); !ok {
		t.Fatal("lowercase skill.md was not loaded")
	}
}

func TestCreateSkillEscapesFrontmatter(t *testing.T) {
	skillRoot := t.TempDir()
	oldSkillPath := ToolsConfInfo.SkillPath
	ToolsConfInfo.SkillPath = &skillRoot
	defer func() {
		ToolsConfInfo.SkillPath = oldSkillPath
		TaskTools.Delete("yaml_skill")
		SkillTools.Delete("yaml_skill")
	}()

	err := CreateSkill("yaml_skill", "Use when prompts say: summarize this.", "Summarize safely.")
	if err != nil {
		t.Fatal(err)
	}

	value, ok := TaskTools.Load("yaml_skill")
	if !ok {
		t.Fatal("created skill was not loaded")
	}
	skill, ok := value.(*AgentInfo)
	if !ok {
		t.Fatalf("TaskTools contains %T, want *AgentInfo", value)
	}
	if skill.Description != "Use when prompts say: summarize this." {
		t.Errorf("description = %q", skill.Description)
	}
}
