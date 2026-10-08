package conf

import (
	"bytes"
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

func InitSkills() {
	SkillTools.Range(func(key, _ any) bool {
		TaskTools.Delete(key)
		SkillTools.Delete(key)
		return true
	})

	skillPath := *ToolsConfInfo.SkillPath
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
			logger.Error("init skills fail", "path", skillPath, "err", err)
		}
		return
	}

	sort.Strings(matches)
	for _, match := range matches {
		skillName, skill, err := loadSkill(match)
		if err != nil {
			logger.Error("load skill fail", "path", match, "err", err)
			continue
		}
		if _, loaded := TaskTools.LoadOrStore(skillName, skill); loaded {
			logger.Error("skill name conflicts with an existing task agent", "name", skillName, "path", match)
			continue
		}
		SkillTools.Store(skillName, skill)
		logger.Info("load skill", "name", skillName, "path", match)
	}
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
