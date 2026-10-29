package conf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yincongcyincong/MuseBot/logger"
	"gopkg.in/yaml.v3"
)

const (
	maxSkillFileBytes       = 256 * 1024
	maxSkillResourceBytes   = 2 * 1024 * 1024
	maxSkillDescriptionSize = 1024
)

var skillNameReg = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

type skillInfo struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

func InitSkills(ctx context.Context) {
	SkillTools.Range(func(key, _ any) bool {
		TaskTools.Delete(key)
		SkillTools.Delete(key)
		return true
	})

	skillPath := *ToolsConfInfo.SkillPath
	// data/skills is the bot's own skill home; make sure it exists so
	// installs and creates always land there
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		logger.ErrorCtx(ctx, "create skill path fail", "path", skillPath, "err", err)
		return
	}
	matches := make([]string, 0)
	err := filepath.WalkDir(skillPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(entry.Name(), "SKILL.md") {
			return nil
		}
		matches = append(matches, path)
		return nil
	})
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.ErrorCtx(ctx, "init skills fail", "path", skillPath, "err", err)
		}
		return
	}

	sort.Strings(matches)
	for _, match := range matches {
		skillName, skill, err := loadSkill(match)
		if err != nil {
			logger.ErrorCtx(ctx, "load skill fail", "path", match, "err", err)
			continue
		}
		if _, loaded := TaskTools.LoadOrStore(skillName, skill); loaded {
			logger.ErrorCtx(ctx, "skill name conflicts with an existing task agent", "name", skillName, "path", match)
			continue
		}
		SkillTools.Store(skillName, skill)
		logger.InfoCtx(ctx, "load skill", "name", skillName, "path", match)
	}
}

func CreateSkill(ctx context.Context, name, description, instructions string) error {
	name = strings.TrimSpace(name)
	if !skillNameReg.MatchString(name) {
		return errors.New("invalid skill name; use 1-64 letters, digits, hyphens, or underscores")
	}
	description = strings.TrimSpace(description)
	if description == "" {
		return errors.New("skill description is required")
	}

	skillDir := filepath.Join(*ToolsConfInfo.SkillPath, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err == nil {
		return errors.New("skill already exists")
	}

	frontmatter, err := yaml.Marshal(skillInfo{Name: name, Description: description})
	if err != nil {
		return err
	}
	skillContent := "---\n" + string(frontmatter) + "---\n\n" + strings.TrimSpace(instructions) + "\n"
	if len(skillContent) > maxSkillFileBytes {
		return fmt.Errorf("SKILL.md exceeds %d bytes", maxSkillFileBytes)
	}
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		return err
	}

	InitSkills(ctx)
	return nil
}

// DeleteSkills removes the directory of the named skill, or of every loaded
// skill when name is empty, then reloads skills. Only directories that contain
// a SKILL.md inside the configured skill path are ever removed; a bare
// SKILL.md sitting directly in the skill path root has its file removed
// instead of the root directory itself.
func DeleteSkills(ctx context.Context, name string) ([]string, error) {
	skillPath, err := filepath.Abs(*ToolsConfInfo.SkillPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(skillPath)
	if err != nil || !info.IsDir() {
		return nil, errors.New("skill path is not a directory")
	}

	type skillEntry struct {
		name string
		file string
	}
	entries := make([]skillEntry, 0)
	err = filepath.WalkDir(skillPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(entry.Name(), "SKILL.md") {
			return nil
		}
		skillName, _, loadErr := loadSkill(path)
		if loadErr != nil {
			logger.WarnCtx(ctx, "load skill fail, skip delete", "path", path, "err", loadErr)
			return nil
		}
		entries = append(entries, skillEntry{name: skillName, file: path})
		return nil
	})
	if err != nil {
		return nil, err
	}

	deleted := make([]string, 0)
	for _, entry := range entries {
		if name != "" && entry.name != name {
			continue
		}
		dir := filepath.Dir(entry.file)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		if absDir == skillPath {
			// never RemoveAll the skill root; the file is the whole skill
			if err := os.Remove(entry.file); err != nil {
				return nil, err
			}
			deleted = append(deleted, entry.name)
			continue
		}
		if !strings.HasPrefix(absDir, skillPath+string(os.PathSeparator)) {
			return nil, fmt.Errorf("skill directory %s escapes the skill path", absDir)
		}
		if err := os.RemoveAll(absDir); err != nil {
			return nil, err
		}
		deleted = append(deleted, entry.name)
	}

	InitSkills(ctx)
	return deleted, nil
}

func loadSkill(skillFile string) (string, *AgentInfo, error) {
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return "", nil, err
	}
	if len(content) > maxSkillFileBytes {
		return "", nil, fmt.Errorf("SKILL.md exceeds %d bytes", maxSkillFileBytes)
	}

	metadata, body, err := parseSkillContent(string(content))
	if err != nil {
		return "", nil, err
	}
	skillDir := filepath.Dir(skillFile)
	if metadata.Name == "" {
		metadata.Name = filepath.Base(skillDir)
	}
	metadata.Name = strings.TrimSpace(metadata.Name)
	if !skillNameReg.MatchString(metadata.Name) {
		return "", nil, errors.New("invalid skill name; use 1-64 letters, digits, hyphens, or underscores")
	}
	metadata.Description = strings.TrimSpace(metadata.Description)
	if metadata.Description == "" {
		return "", nil, errors.New("skill description is required")
	}
	if len(metadata.Description) > maxSkillDescriptionSize {
		return "", nil, fmt.Errorf("skill description exceeds %d bytes", maxSkillDescriptionSize)
	}

	instructions, err := buildSkillInstructions(skillDir, strings.TrimSpace(body))
	if err != nil {
		return "", nil, err
	}

	return metadata.Name, &AgentInfo{
		Description:  metadata.Description,
		Instructions: instructions,
		IsSkill:      true,
	}, nil
}

func parseSkillContent(content string) (*skillInfo, string, error) {
	normalized := strings.TrimLeft(content, "\r\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", errors.New("SKILL.md must start with YAML frontmatter delimited by ---")
	}

	end := -1
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			end = index
			break
		}
	}
	if end < 0 {
		return nil, "", errors.New("SKILL.md YAML frontmatter is not closed")
	}

	metadata := new(skillInfo)
	frontmatter := strings.Join(lines[1:end], "\n")
	if err := yaml.Unmarshal([]byte(frontmatter), metadata); err != nil {
		return nil, "", err
	}
	body := strings.Join(lines[end+1:], "\n")
	return metadata, body, nil
}

func buildSkillInstructions(skillDir, body string) (string, error) {
	resourceFiles := make([]string, 0)
	err := filepath.WalkDir(skillDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.EqualFold(entry.Name(), "SKILL.md") {
			return nil
		}
		resourceFiles = append(resourceFiles, path)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(resourceFiles)

	var instructions bytes.Buffer
	if body != "" {
		instructions.WriteString("# Skill Instructions\n\n")
		instructions.WriteString(body)
	}

	totalBytes := 0
	for _, resourceFile := range resourceFiles {
		relativePath, err := filepath.Rel(skillDir, resourceFile)
		if err != nil {
			return "", err
		}
		relativePath = filepath.ToSlash(relativePath)
		info, err := os.Stat(resourceFile)
		if err != nil {
			return "", err
		}
		if info.Size() > maxSkillFileBytes || totalBytes+int(info.Size()) > maxSkillResourceBytes {
			writeSkillResourceNotice(&instructions, relativePath, "not loaded because it exceeds the skill size limit")
			continue
		}

		content, err := os.ReadFile(resourceFile)
		if err != nil {
			return "", err
		}
		if len(content) > maxSkillFileBytes {
			writeSkillResourceNotice(&instructions, relativePath, "not loaded because it exceeds the skill size limit")
			continue
		}
		if bytes.IndexByte(content, 0) >= 0 {
			writeSkillResourceNotice(&instructions, relativePath, "not loaded because it is a binary file")
			continue
		}
		totalBytes += len(content)
		if instructions.Len() > 0 {
			instructions.WriteString("\n\n")
		}
		instructions.WriteString("# Skill Resource: " + relativePath + "\n\n")
		instructions.Write(content)
	}

	return strings.TrimSpace(instructions.String()), nil
}

func writeSkillResourceNotice(instructions *bytes.Buffer, path, reason string) {
	if instructions.Len() > 0 {
		instructions.WriteString("\n\n")
	}
	instructions.WriteString("# Skill Resource: " + path + "\n\n" + reason)
}
