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

const sofyaSearchURL = "https://sofya.co/v1/search"

// SofyaWebSearch calls the Sofya Search API with the key from SOFYA_API_KEY.
// It keeps the built-in web_search response contract, so task models receive
// source URLs and snippets.
func SofyaWebSearch(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryRunes {
		return "", fmt.Errorf("query exceeds %d characters", webSearchMaxQueryRunes)
	}
	apiKey := strings.TrimSpace(os.Getenv("SOFYA_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("SOFYA_API_KEY is not set")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	// "snippets" returns the result snippets without reading each page, which
	// is what the web_search contract carries.
	body, err := json.Marshal(map[string]interface{}{
		"query":        query,
		"max_results":  webSearchMaxResults,
		"search_depth": "snippets",
	})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sofyaSearchURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "MuseBot/SofyaSearch")

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
		return "", fmt.Errorf("Sofya search response exceeds %d bytes", webSearchMaxBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", sofyaSearchError(response.StatusCode, content)
	}
	return parseSofyaSearch(query, content)
}

// sofyaSearchError names the common failures (401 invalid key, 402 out of
// credits or over the key's limit, 429 rate limited) and keeps the API's own
// message for the rest. Sofya does not charge for a failed request.
func sofyaSearchError(status int, content []byte) error {
	var body struct {
		Detail interface{} `json:"detail"`
	}
	_ = json.Unmarshal(content, &body)
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("Sofya search returned HTTP %d: check SOFYA_API_KEY", status)
	case http.StatusPaymentRequired:
		return fmt.Errorf("Sofya search returned HTTP %d: no credits left, or the key reached its limit", status)
	case http.StatusTooManyRequests:
		return fmt.Errorf("Sofya search returned HTTP %d: rate limited", status)
	}
	if detail, ok := body.Detail.(string); ok && detail != "" {
		return fmt.Errorf("Sofya search returned HTTP %d: %s", status, detail)
	}
	return fmt.Errorf("Sofya search returned HTTP %d", status)
}

func parseSofyaSearch(query string, content []byte) (string, error) {
	var response struct {
		Results []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Content     string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("decode Sofya search results: %w", err)
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
		snippet := item.Description
		if snippet == "" {
			snippet = item.Content
		}
		output.Results = append(output.Results, WebSearchResult{Title: title, URL: item.URL, Snippet: snippet})
		if len(output.Results) == webSearchMaxResults {
			break
		}
	}
	if len(output.Results) == 0 {
		return "", fmt.Errorf("Sofya web search returned no usable results")
	}
	encoded, err := json.Marshal(output)
	return string(encoded), err
}
