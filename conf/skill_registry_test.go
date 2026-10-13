package conf

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractGitHubSkillURLs(t *testing.T) {
	content := `<a href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgithub.com%2Fexample%2Fpatent%2Fblob%2Fmain%2FSKILL.md&amp;rut=x">one</a>
<a href="https://github.com/example/other/tree/main/skills/release/">two</a>`
	urls := extractGitHubURLs(content)
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d: %v", len(urls), urls)
	}
	if urls[0] != "https://github.com/example/patent/blob/main/SKILL.md" {
		t.Errorf("unexpected decoded URL: %q", urls[0])
	}
	if urls[1] != "https://github.com/example/other" {
		t.Errorf("unexpected repository URL: %q", urls[1])
	}
}

func TestParseGitHubSkillURLRejectsNonGitHubSource(t *testing.T) {
	_, err := parseGitHubSkillURL("https://example.com/owner/repo/blob/main/SKILL.md")
	if err == nil || !strings.Contains(err.Error(), "https://github.com") {
		t.Fatalf("expected GitHub-only error, got %v", err)
	}
}

func TestInstallRemoteSkillDownloadsValidatedTextResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.URL.Path == "/repos/example/patent-writing":
			_ = json.NewEncoder(writer).Encode(githubRepoInfo{DefaultBranch: "main"})
		case request.URL.Path == "/repos/example/patent-writing/git/trees/main":
			if request.URL.Query().Get("recursive") != "1" {
				http.Error(writer, "missing recursive flag", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(writer).Encode(githubTree{Tree: []githubTreeItem{
				{Path: "SKILL.md", Type: "blob"},
				{Path: "references/guide.md", Type: "blob"},
				{Path: "install.sh", Type: "blob"},
			}})
		case request.URL.Path == "/raw/example/patent-writing/main/SKILL.md":
			_, _ = writer.Write([]byte("---\nname: patent-writing\ndescription: Draft patent applications.\n---\n\n# Patent writing\n"))
		case request.URL.Path == "/raw/example/patent-writing/main/references/guide.md":
			_, _ = writer.Write([]byte("# Guide\n\nUse claim structure.\n"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	previousClient := remoteSkillHTTPClient
	previousAPI := githubAPIBaseURL
	previousRaw := githubRawBaseURL
	previousSkillPath := ToolsConfInfo.SkillPath
	base := t.TempDir()
	ToolsConfInfo.SkillPath = stringPtr(filepath.Join(base, "skills"))
	remoteSkillHTTPClient = server.Client()
	githubAPIBaseURL = server.URL
	githubRawBaseURL = server.URL + "/raw"
	t.Cleanup(func() {
		remoteSkillHTTPClient = previousClient
		githubAPIBaseURL = previousAPI
		githubRawBaseURL = previousRaw
		ToolsConfInfo.SkillPath = previousSkillPath
		TaskTools.Delete("patent-writing")
		SkillTools.Delete("patent-writing")
	})

	source := "https://github.com/example/patent-writing/blob/main/SKILL.md"
	result, err := InstallRemoteSkill(context.Background(), source)
	if err != nil {
		t.Fatalf("install remote skill: %v", err)
	}
	if !strings.Contains(result, source) {
		t.Errorf("result does not contain source: %q", result)
	}
	if _, err := os.Stat(filepath.Join(*ToolsConfInfo.SkillPath, "patent-writing", "references", "guide.md")); err != nil {
		t.Fatalf("text resource was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(*ToolsConfInfo.SkillPath, "patent-writing", "install.sh")); err == nil {
		t.Fatal("executable resource must not be installed")
	}

	loaded, ok := SkillTools.Load("patent-writing")
	if !ok {
		t.Fatal("installed skill was not loaded")
	}
	agent := loaded.(*AgentInfo)
	if !strings.Contains(agent.Instructions, "Use claim structure.") {
		t.Error("installed skill does not contain resource content")
	}
}
