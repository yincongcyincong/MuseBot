package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
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

	c, err := client.NewStreamableHttpClient(parallelSearchURL,
		transport.WithHTTPHeaders(map[string]string{"User-Agent": "MuseBot/ParallelSearch"}),
		transport.WithHTTPTimeout(webSearchTimeout))
	if err != nil {
		return "", err
	}
	defer c.Close()
	if err := c.Start(ctx); err != nil {
		return "", err
	}
	init := mcp.InitializeRequest{}
	init.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	init.Params.ClientInfo = mcp.Implementation{Name: "MuseBot", Version: "1"}
	if _, err := c.Initialize(ctx, init); err != nil {
		return "", fmt.Errorf("initialize Parallel search: %w", err)
	}
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return "", fmt.Errorf("discover Parallel search: %w", err)
	}
	found := false
	for _, tool := range tools.Tools {
		found = found || tool.Name == "web_search"
	}
	if !found {
		return "", fmt.Errorf("Parallel MCP server does not expose web_search")
	}
	request := mcp.CallToolRequest{}
	request.Params.Name = "web_search"
	request.Params.Arguments = map[string]interface{}{
		"objective": query, "search_queries": []string{query}, "session_id": sessionID,
	}
	result, err := c.CallTool(ctx, request)
	if err != nil {
		return "", fmt.Errorf("Parallel web search: %w", err)
	}
	if result.IsError {
		return "", fmt.Errorf("Parallel web search returned an error: %s", parallelSearchText(result))
	}
	return parseParallelSearch(query, parallelSearchText(result))
}

func parallelSearchText(result *mcp.CallToolResult) string {
	var text strings.Builder
	for _, content := range result.Content {
		if part, ok := content.(mcp.TextContent); ok {
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
