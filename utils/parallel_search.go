package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const parallelSearchURL = "https://search.parallel.ai/mcp"

// WebSearchSession belongs to one research run, independently of MCP connections.
// Sharing it between the run's agents keeps retries and follow-up searches together.
type WebSearchSession struct {
	once sync.Once
	id   string
}

func (s *WebSearchSession) ID() string {
	s.once.Do(func() { s.id = uuid.NewString() })
	return s.id
}

// parallelSearchTransport attaches the identifying User-Agent to every request;
// the search endpoint is a plain anonymous MCP server, no auth headers needed.
type parallelSearchTransport struct {
	base http.RoundTripper
}

func (t parallelSearchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", "MuseBot/ParallelSearch")
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

// ParallelWebSearch uses the anonymous Search MCP endpoint. It keeps the built-in
// web_search response contract, so task models receive source URLs and snippets.
func ParallelWebSearch(ctx context.Context, query, sessionID string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryRunes {
		return "", fmt.Errorf("query exceeds %d characters", webSearchMaxQueryRunes)
	}
	if sessionID == "" {
		return "", fmt.Errorf("web search session is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "MuseBot", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: parallelSearchURL,
		HTTPClient: &http.Client{
			Timeout:   webSearchTimeout,
			Transport: parallelSearchTransport{},
		},
		// one-shot tool call: no server-initiated messages are wanted, and
		// skipping the standalone SSE stream avoids a dangling GET per search
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("initialize Parallel search: %w", err)
	}
	defer session.Close()

	found, err := parallelSearchExposes(ctx, session, "web_search")
	if err != nil {
		return "", fmt.Errorf("discover Parallel search: %w", err)
	}
	if !found {
		return "", fmt.Errorf("Parallel MCP server does not expose web_search")
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "web_search",
		Arguments: map[string]interface{}{
			"objective": query, "search_queries": []string{query}, "session_id": sessionID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("Parallel web search: %w", err)
	}
	if result.IsError {
		return "", fmt.Errorf("Parallel web search returned an error: %s", parallelSearchText(result))
	}
	return parseParallelSearch(query, parallelSearchText(result))
}

// parallelSearchExposes walks the server's tool list (with pagination) looking
// for the given tool name.
func parallelSearchExposes(ctx context.Context, session *mcp.ClientSession, name string) (bool, error) {
	found := false
	for cursor := ""; ; {
		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return false, err
		}
		for _, tool := range tools.Tools {
			if tool.Name == name {
				found = true
			}
		}
		if tools.NextCursor == "" {
			return found, nil
		}
		cursor = tools.NextCursor
	}
}

func parallelSearchText(result *mcp.CallToolResult) string {
	var text strings.Builder
	if result == nil {
		return ""
	}
	for _, content := range result.Content {
		if part, ok := content.(*mcp.TextContent); ok {
			text.WriteString(part.Text)
		}
	}
	return text.String()
}

func parseParallelSearch(query, content string) (string, error) {
	if len(content) > webSearchMaxBytes {
		return "", fmt.Errorf("Parallel search response exceeds %d bytes", webSearchMaxBytes)
	}
	var response struct {
		Results []struct {
			Title    string   `json:"title"`
			URL      string   `json:"url"`
			Excerpts []string `json:"excerpts"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return "", fmt.Errorf("decode Parallel search results: %w", err)
	}
	output := WebSearchResponse{Query: query, Results: make([]WebSearchResult, 0)}
	seen := make(map[string]bool)
	for _, item := range response.Results {
		parsed, err := url.Parse(item.URL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || seen[item.URL] {
			continue
		}
		seen[item.URL] = true
		title := item.Title
		if title == "" {
			title = item.URL
		}
		output.Results = append(output.Results, WebSearchResult{Title: title, URL: item.URL, Snippet: strings.Join(item.Excerpts, "\n\n")})
		if len(output.Results) == webSearchMaxResults {
			break
		}
	}
	if len(output.Results) == 0 {
		return "", fmt.Errorf("Parallel web search returned no usable results")
	}
	encoded, err := json.Marshal(output)
	return string(encoded), err
}
