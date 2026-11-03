package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"
)

const outcrawlSearchURL = "https://api.outcrawl.ai/v1/search"

// OutcrawlWebSearch calls the Outcrawl Search API with the key from OUTCRAWL_API_KEY.
// It keeps the built-in web_search response contract, so task models receive
// source URLs and snippets.
func OutcrawlWebSearch(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryRunes {
		return "", fmt.Errorf("query exceeds %d characters", webSearchMaxQueryRunes)
	}
	apiKey := strings.TrimSpace(os.Getenv("OUTCRAWL_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("OUTCRAWL_API_KEY is not set")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	body, err := json.Marshal(map[string]interface{}{"query": query, "limit": webSearchMaxResults})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, outcrawlSearchURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "MuseBot/OutcrawlSearch")

	httpClient := &http.Client{Timeout: webSearchTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(io.LimitReader(response.Body, webSearchMaxBytes+1))
	if err != nil {
		return "", err
	}
	if len(content) > webSearchMaxBytes {
		return "", fmt.Errorf("Outcrawl search response exceeds %d bytes", webSearchMaxBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", outcrawlSearchError(response.StatusCode, content)
	}
	return parseOutcrawlSearch(query, content)
}

// outcrawlSearchError names the common failures (401 invalid key, 402 out of
// credits, 429 rate limited) and keeps the API's own message for the rest.
func outcrawlSearchError(status int, content []byte) error {
	var body struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(content, &body)
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("Outcrawl search returned HTTP %d: check OUTCRAWL_API_KEY", status)
	case http.StatusPaymentRequired:
		return fmt.Errorf("Outcrawl search returned HTTP %d: no credits left", status)
	case http.StatusTooManyRequests:
		return fmt.Errorf("Outcrawl search returned HTTP %d: rate limited", status)
	}
	if body.Message != "" {
		return fmt.Errorf("Outcrawl search returned HTTP %d: %s", status, body.Message)
	}
	return fmt.Errorf("Outcrawl search returned HTTP %d", status)
}

func parseOutcrawlSearch(query string, content []byte) (string, error) {
	var response struct {
		Hits []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("decode Outcrawl search results: %w", err)
	}
	output := WebSearchResponse{Query: query, Results: make([]WebSearchResult, 0)}
	seen := make(map[string]bool)
	for _, item := range response.Hits {
		parsed, err := url.Parse(item.URL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || seen[item.URL] {
			continue
		}
		seen[item.URL] = true
		title := item.Title
		if title == "" {
			title = item.URL
		}
		output.Results = append(output.Results, WebSearchResult{Title: title, URL: item.URL, Snippet: item.Snippet})
		if len(output.Results) == webSearchMaxResults {
			break
		}
	}
	if len(output.Results) == 0 {
		return "", fmt.Errorf("Outcrawl web search returned no usable results")
	}
	encoded, err := json.Marshal(output)
	return string(encoded), err
}
