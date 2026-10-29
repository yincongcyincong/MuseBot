package conf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	remoteSkillQueryMaxBytes  = 512 * 1024
	remoteSkillFileMaxBytes   = int64(maxSkillFileBytes)
	remoteSkillTotalMaxBytes  = int64(maxSkillResourceBytes)
	remoteSkillMaxFiles       = 100
	remoteSkillMaxCandidates  = 5
	remoteSkillQueryMaxRunes  = 128
	remoteSkillHTTPTimeoutSec = 20
	remoteSkillSearchTimeout  = 45 * time.Second
	remoteSkillInstallTimeout = 90 * time.Second
	remoteSkillOutputName     = "SKILL.md"
)

var (
	remoteSkillHTTPClient = &http.Client{Timeout: remoteSkillHTTPTimeoutSec * time.Second}
	duckDuckGoSkillSearch = "https://lite.duckduckgo.com/lite/?q="
	bingSkillSearch       = "https://www.bing.com/search?q="
	githubAPIBaseURL      = "https://api.github.com"
	githubRawBaseURL      = "https://raw.githubusercontent.com"

	githubOwnerRepoReg = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	githubPathReg      = regexp.MustCompile(`^[A-Za-z0-9._/@^{}!$&+,:;=?#|\-~\[\]]+$`)
	htmlHrefReg        = regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["']`)
	githubURLReg       = regexp.MustCompile(`(?i)https://github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)(?:/[^"'\s<>&]+)?`)
)

// isGitHubShorthand reports whether s looks like "owner/repo", the bare form
// of a GitHub repository reference
func isGitHubShorthand(s string) bool {
	owner, repo, found := strings.Cut(s, "/")
	return found && owner != "" && repo != "" &&
		!strings.Contains(repo, "/") &&
		githubOwnerRepoReg.MatchString(owner) && githubOwnerRepoReg.MatchString(repo)
}

type RemoteSkillCandidate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Path        string `json:"path"`
}

type githubRepoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

type githubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type githubTree struct {
	Truncated bool             `json:"truncated"`
	Tree      []githubTreeItem `json:"tree"`
}

type githubRepoSearchResponse struct {
	Items []githubRepoInfoWithName `json:"items"`
}

type githubRepoInfoWithName struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
}

func SearchRemoteSkills(ctx context.Context, query string) ([]*RemoteSkillCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("skill search query is required")
	}
	if len([]rune(query)) > remoteSkillQueryMaxRunes {
		return nil, fmt.Errorf("skill search query exceeds %d characters", remoteSkillQueryMaxRunes)
	}

	candidates := make([]*RemoteSkillCandidate, 0, remoteSkillMaxCandidates)
	seen := make(map[string]bool)
	searchCtx, cancel := context.WithTimeout(ctx, remoteSkillSearchTimeout)
	defer cancel()
	for _, repositoryQuery := range []string{
		fmt.Sprintf("%q skill", query),
		fmt.Sprintf("%s skill", query),
		fmt.Sprintf("%q", query),
	} {
		candidates, err := searchGitHubRepositories(searchCtx, repositoryQuery, query, seen)
		if err == nil && len(candidates) > 0 {
			return candidates, nil
		}
	}

	searchQueries := []string{
		fmt.Sprintf("site:github.com %q %q", query, remoteSkillOutputName),
		fmt.Sprintf("site:github.com %q SKILL.md codex claude", query),
	}
	for _, searchQuery := range searchQueries {
		for _, provider := range []string{duckDuckGoSkillSearch, bingSkillSearch} {
			if len(candidates) >= remoteSkillMaxCandidates {
				return candidates, nil
			}
			content, err := fetchRemoteURL(searchCtx, provider+url.QueryEscape(searchQuery), remoteSkillQueryMaxBytes)
			if err != nil {
				continue
			}
			for _, source := range extractGitHubURLs(string(content)) {
				found, err := candidatesForGitHubURLLimit(searchCtx, source, query, remoteSkillMaxCandidates)
				if err != nil {
					continue
				}
				for _, candidate := range found {
					if seen[candidate.Source] {
						continue
					}
					seen[candidate.Source] = true
					candidates = append(candidates, candidate)
					if len(candidates) >= remoteSkillMaxCandidates {
						return candidates, nil
					}
				}
			}
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid GitHub skill found for %q", query)
	}
	return candidates, nil
}

func SearchAndInstallRemoteSkill(ctx context.Context, query string) (string, error) {
	// the LLM often pastes the repo URL (or owner/repo) into the query field;
	// treat those as direct sources instead of search keywords
	trimmed := strings.TrimSpace(query)
	if strings.HasPrefix(trimmed, "https://github.com/") {
		return InstallRemoteSkill(ctx, trimmed)
	}
	if isGitHubShorthand(trimmed) {
		return InstallRemoteSkill(ctx, "https://github.com/"+trimmed)
	}
	candidates, err := SearchRemoteSkills(ctx, query)
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no valid GitHub skill found for %q", query)
	}
	return InstallRemoteSkill(ctx, candidates[0].Source)
}

func InstallRemoteSkill(ctx context.Context, source string) (string, error) {
	installCtx, cancel := context.WithTimeout(ctx, remoteSkillInstallTimeout)
	defer cancel()
	candidate, err := resolveRemoteSkill(ctx, source)
	if err != nil {
		return "", err
	}
	// already present and intact -> idempotent no-op; a partial download
	// (unparsable SKILL.md) falls through and is reinstalled
	existingDir := filepath.Join(*ToolsConfInfo.SkillPath, candidate.Name)
	if existingSkill, _, loadErr := loadSkill(filepath.Join(existingDir, "SKILL.md")); loadErr == nil && existingSkill == candidate.Name {
		return "skill " + candidate.Name + " is already installed", nil
	}
	if err := downloadRemoteSkill(installCtx, candidate); err != nil {
		return "", err
	}
	InitSkills(ctx)
	return "installed " + candidate.Name + " from " + candidate.Source, nil
}

func fetchRemoteURL(ctx context.Context, requestURL string, maxBytes int64) ([]byte, error) {
	parsed, err := url.Parse(requestURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return nil, fmt.Errorf("invalid remote URL")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "MuseBot skill installer")
	request.Header.Set("Accept", "text/html,application/json,text/plain,*/*")

	response, err := remoteSkillHTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("remote request returned HTTP %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > maxBytes {
		return nil, fmt.Errorf("remote response exceeds %d bytes", maxBytes)
	}
	return content, nil
}

func extractGitHubURLs(content string) []string {
	escaped := strings.NewReplacer("&amp;", "&", "&#x2F;", "/", `\u002F`, "/").Replace(content)
	hrefs := htmlHrefReg.FindAllStringSubmatch(escaped, -1)
	urls := make([]string, 0, len(hrefs))
	seen := make(map[string]bool)
	for _, match := range hrefs {
		rawURL := match[1]
		if strings.Contains(rawURL, "duckduckgo.com/l/?uddg=") {
			parsed, err := url.Parse(rawURL)
			if err != nil {
				continue
			}
			redirect := parsed.Query().Get("uddg")
			if redirect == "" {
				continue
			}
			rawURL = redirect
		}
		if strings.HasPrefix(rawURL, "//") {
			rawURL = "https:" + rawURL
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Hostname() != "github.com" {
			continue
		}
		parsed.Fragment = ""
		parsed.RawQuery = ""
		cleaned := strings.TrimSuffix(parsed.String(), "/")
		if !strings.Contains(strings.ToLower(cleaned), strings.ToLower(remoteSkillOutputName)) {
			if githubMatch := githubURLReg.FindStringSubmatch(cleaned); githubMatch != nil {
				cleaned = "https://github.com/" + githubMatch[1] + "/" + githubMatch[2]
			}
		}
		cleaned = strings.TrimSuffix(cleaned, "/")
		if seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		urls = append(urls, cleaned)
	}
	return urls
}

func candidatesForGitHubURL(ctx context.Context, source string) ([]*RemoteSkillCandidate, error) {
	return candidatesForGitHubURLLimit(ctx, source, "", remoteSkillMaxCandidates)
}

func candidatesForGitHubURLLimit(ctx context.Context, source, query string, limit int) ([]*RemoteSkillCandidate, error) {
	parsed, err := parseGitHubSkillURL(source)
	if err != nil {
		return nil, err
	}
	if parsed.Branch != "" && strings.EqualFold(path.Base(parsed.Path), remoteSkillOutputName) {
		candidate, err := newRemoteSkillCandidate(ctx, parsed.Owner, parsed.Repo, parsed.Branch, parsed.Path)
		if err != nil {
			return nil, err
		}
		return []*RemoteSkillCandidate{candidate}, nil
	}

	branch := parsed.Branch
	prefix := parsed.Path
	if branch == "" {
		info, err := getGitHubRepoInfo(ctx, parsed.Owner, parsed.Repo)
		if err != nil {
			return nil, err
		}
		branch = info.DefaultBranch
	}
	tree, err := getGitHubTree(ctx, parsed.Owner, parsed.Repo, branch)
	if err != nil {
		return nil, err
	}
	sort.Slice(tree.Tree, func(firstIndex, secondIndex int) bool {
		firstScore := remoteSkillPathScore(tree.Tree[firstIndex].Path, query)
		secondScore := remoteSkillPathScore(tree.Tree[secondIndex].Path, query)
		if firstScore != secondScore {
			return firstScore > secondScore
		}
		return tree.Tree[firstIndex].Path < tree.Tree[secondIndex].Path
	})
	candidates := make([]*RemoteSkillCandidate, 0)
	for _, item := range tree.Tree {
		if item.Type != "blob" || !strings.EqualFold(path.Base(item.Path), remoteSkillOutputName) {
			continue
		}
		if prefix != "" && !strings.HasPrefix(item.Path, prefix+"/") {
			continue
		}
		candidate, err := newRemoteSkillCandidate(ctx, parsed.Owner, parsed.Repo, branch, item.Path)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate)
		if len(candidates) >= limit {
			break
		}
	}
	return candidates, nil
}

type parsedGitHubSkill struct {
	Owner  string
	Repo   string
	Branch string
	Path   string
}

func parseGitHubSkillURL(source string) (*parsedGitHubSkill, error) {
	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "github.com" {
		return nil, fmt.Errorf("skill source must be an https://github.com URL")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || !githubOwnerRepoReg.MatchString(parts[0]) || !githubOwnerRepoReg.MatchString(parts[1]) {
		return nil, fmt.Errorf("invalid GitHub repository URL")
	}
	result := &parsedGitHubSkill{Owner: parts[0], Repo: strings.TrimSuffix(parts[1], ".git")}
	if len(parts) <= 2 {
		return result, nil
	}
	if parts[2] != "blob" && parts[2] != "tree" {
		return nil, fmt.Errorf("GitHub URL must identify a repository, tree, blob, or SKILL.md")
	}
	if len(parts) < 5 {
		return nil, fmt.Errorf("GitHub blob URL requires a branch and file path")
	}
	remaining := parts[3:]
	for index := 1; index < len(remaining); index++ {
		branch := strings.Join(remaining[:index], "/")
		filePath := strings.Join(remaining[index:], "/")
		if isSafeGitHubPath(branch) && isSafeGitHubPath(filePath) {
			result.Branch = branch
			result.Path = filePath
			return result, nil
		}
	}
	return nil, fmt.Errorf("invalid GitHub branch or path")
}

func isSafeGitHubPath(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." || !githubPathReg.MatchString(part) {
			return false
		}
	}
	return true
}

func resolveRemoteSkill(ctx context.Context, source string) (*RemoteSkillCandidate, error) {
	parsed, err := parseGitHubSkillURL(source)
	if err != nil {
		return nil, err
	}
	if parsed.Branch != "" && strings.EqualFold(path.Base(parsed.Path), remoteSkillOutputName) {
		return newRemoteSkillCandidate(ctx, parsed.Owner, parsed.Repo, parsed.Branch, parsed.Path)
	}
	candidates, err := candidatesForGitHubURL(ctx, source)
	if err != nil {
		return nil, err
	}
	if len(candidates) != 1 {
		example := ""
		if len(candidates) > 0 {
			example = fmt.Sprintf(", e.g. %s", candidates[0].Source)
		}
		return nil, fmt.Errorf("source contains %d skills; install a direct SKILL.md URL%s", len(candidates), example)
	}
	return candidates[0], nil
}

func newRemoteSkillCandidate(ctx context.Context, owner, repo, branch, filePath string) (*RemoteSkillCandidate, error) {
	if !isSafeGitHubPath(filePath) || !strings.EqualFold(path.Base(filePath), remoteSkillOutputName) {
		return nil, fmt.Errorf("invalid SKILL.md path")
	}
	content, err := fetchRemoteURL(ctx, githubRawURL(owner, repo, branch, filePath), remoteSkillFileMaxBytes)
	if err != nil {
		return nil, err
	}
	metadata, _, err := parseSkillContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("validate remote SKILL.md: %w", err)
	}
	name := strings.TrimSpace(metadata.Name)
	if name == "" {
		name = path.Base(path.Dir(filePath))
	}
	if !skillNameReg.MatchString(name) {
		return nil, fmt.Errorf("remote skill has invalid name %q", name)
	}
	return &RemoteSkillCandidate{
		Name:        name,
		Description: strings.TrimSpace(metadata.Description),
		Source:      fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", owner, repo, branch, filePath),
		Owner:       owner,
		Repo:        repo,
		Branch:      branch,
		Path:        filePath,
	}, nil
}

func getGitHubRepoInfo(ctx context.Context, owner, repo string) (*githubRepoInfo, error) {
	content, err := fetchGitHubAPI(ctx, fmt.Sprintf("/repos/%s/%s", owner, repo))
	if err != nil {
		return nil, err
	}
	info := new(githubRepoInfo)
	if err := json.Unmarshal(content, info); err != nil {
		return nil, err
	}
	return info, nil
}

func getGitHubTree(ctx context.Context, owner, repo, branch string) (*githubTree, error) {
	content, err := fetchGitHubAPI(ctx, fmt.Sprintf("/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, url.PathEscape(branch)))
	if err != nil {
		return nil, err
	}
	tree := new(githubTree)
	if err := json.Unmarshal(content, tree); err != nil {
		return nil, err
	}
	return tree, nil
}

func searchGitHubRepositories(ctx context.Context, githubQuery, originalQuery string, seen map[string]bool) ([]*RemoteSkillCandidate, error) {
	requestURL := fmt.Sprintf("%s/search/repositories?per_page=10&q=%s", githubAPIBaseURL, url.QueryEscape(githubQuery))
	content, err := fetchGitHubAPIRaw(ctx, requestURL)
	if err != nil {
		return nil, fmt.Errorf("no search results were returned")
	}
	response := new(githubRepoSearchResponse)
	if err := json.Unmarshal(content, response); err != nil {
		return nil, fmt.Errorf("no search results were returned")
	}
	candidates := make([]*RemoteSkillCandidate, 0)
	for _, item := range response.Items {
		owner, repo, found := strings.Cut(item.FullName, "/")
		if !found || seen["https://github.com/"+item.FullName] {
			continue
		}
		repoCandidates, err := candidatesForGitHubURLLimit(ctx, "https://github.com/"+owner+"/"+repo, originalQuery, 1)
		if err != nil {
			continue
		}
		if len(repoCandidates) == 0 {
			continue
		}
		candidate := repoCandidates[0]
		if seen[candidate.Source] {
			continue
		}
		seen[candidate.Source] = true
		candidates = append(candidates, candidate)
		if len(candidates) >= remoteSkillMaxCandidates {
			return candidates, nil
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid GitHub skill found for %q", originalQuery)
	}
	return candidates, nil
}

func remoteSkillPathScore(filePath, query string) int {
	if query == "" {
		return 0
	}
	filePath = strings.ToLower(filePath)
	score := 0
	for _, term := range strings.FieldsFunc(strings.ToLower(query), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsDigit(character)
	}) {
		if term != "" && strings.Contains(filePath, term) {
			score += 10
		}
	}
	if strings.Contains(filePath, "skill") {
		score += 3
	}
	return score
}

func fetchGitHubAPI(ctx context.Context, endpoint string) ([]byte, error) {
	return fetchGitHubAPIRaw(ctx, githubAPIBaseURL+endpoint)
}

func fetchGitHubAPIRaw(ctx context.Context, requestURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "MuseBot skill installer")
	request.Header.Set("Accept", "application/vnd.github+json")
	if token := githubToken(); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := remoteSkillHTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API request returned HTTP %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, remoteSkillQueryMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > remoteSkillQueryMaxBytes {
		return nil, fmt.Errorf("GitHub API response exceeds %d bytes", remoteSkillQueryMaxBytes)
	}
	if strings.Contains(strings.TrimSpace(string(content)), "API rate limit exceeded") {
		return nil, fmt.Errorf("GitHub API rate limit exceeded")
	}
	return content, nil
}

func githubToken() string {
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		return token
	}
	return strings.TrimSpace(os.Getenv("GH_TOKEN"))
}

func githubRawURL(owner, repo, branch, filePath string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", githubRawBaseURL, owner, repo, branch, filePath)
}

func downloadRemoteSkill(ctx context.Context, candidate *RemoteSkillCandidate) error {
	tree, err := getGitHubTree(ctx, candidate.Owner, candidate.Repo, candidate.Branch)
	if err != nil {
		return err
	}
	if tree.Truncated {
		return fmt.Errorf("GitHub file tree is too large to install safely")
	}

	skillDir := path.Dir(candidate.Path)
	files := make([]string, 0, 16)
	for _, item := range tree.Tree {
		if item.Type != "blob" || (skillDir != "." && !strings.HasPrefix(item.Path, skillDir+"/")) {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(item.Path), ".md") &&
			!strings.HasSuffix(strings.ToLower(item.Path), ".markdown") &&
			!strings.HasSuffix(strings.ToLower(item.Path), ".txt") {
			continue
		}
		files = append(files, item.Path)
	}
	if len(files) == 0 || len(files) > remoteSkillMaxFiles {
		return fmt.Errorf("remote skill contains an unsupported number of text files")
	}

	destination := filepath.Join(*ToolsConfInfo.SkillPath, candidate.Name)
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("skill %q already exists", candidate.Name)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(filepath.Dir(destination), "."+candidate.Name+".tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	var totalBytes int64
	for _, filePath := range files {
		content, err := fetchRemoteURL(ctx, githubRawURL(candidate.Owner, candidate.Repo, candidate.Branch, filePath), remoteSkillFileMaxBytes)
		if err != nil {
			return err
		}
		if int64(len(content)) > remoteSkillFileMaxBytes {
			return fmt.Errorf("remote file exceeds %d bytes: %s", remoteSkillFileMaxBytes, filePath)
		}
		totalBytes += int64(len(content))
		if totalBytes > remoteSkillTotalMaxBytes {
			return fmt.Errorf("remote skill exceeds %d total bytes", remoteSkillTotalMaxBytes)
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return fmt.Errorf("remote skill contains a binary file: %s", filePath)
		}

		relativePath := filePath
		if skillDir != "." {
			relativePath = strings.TrimPrefix(filePath, skillDir+"/")
		}
		target := filepath.Join(tempDir, filepath.FromSlash(relativePath))
		if !strings.HasPrefix(target, tempDir+string(os.PathSeparator)) {
			return fmt.Errorf("remote skill contains an unsafe path: %s", filePath)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return err
		}
	}
	return os.Rename(tempDir, destination)
}
