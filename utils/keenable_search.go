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

// Keenable Search API: https://docs.keenable.ai/api-reference/search
const (
	keenableSearchURL       = "https://api.keenable.ai/v1/search"
	keenablePublicSearchURL = "https://api.keenable.ai/v1/search/public"
	keenableAppTitle        = "MuseBot"
	keenableSnippetMaxChars = 1000
)

// KeenableWebSearch calls Keenable's search API. Without KEENABLE_API_KEY it uses
// the keyless public endpoint (rate limited per IP, identified by X-Keenable-Title);
// with a key it uses the keyed endpoint and sends X-API-Key. It keeps the built-in
// web_search response contract, so task models receive source URLs and snippets.
func KeenableWebSearch(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryRunes {
		return "", fmt.Errorf("query exceeds %d characters", webSearchMaxQueryRunes)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	body, err := json.Marshal(map[string]interface{}{
		"query":              query,
		"max_results":        webSearchMaxResults,
		"snippet_max_length": keenableSnippetMaxChars,
	})
	if err != nil {
		return "", err
	}
	apiKey := strings.TrimSpace(os.Getenv("KEENABLE_API_KEY"))
	endpoint := keenablePublicSearchURL
	if apiKey != "" {
		endpoint = keenableSearchURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	if apiKey != "" {
		request.Header.Set("X-API-Key", apiKey)
	} else {
		request.Header.Set("X-Keenable-Title", keenableAppTitle)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "MuseBot/KeenableSearch")

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
		return "", fmt.Errorf("Keenable search response exceeds %d bytes", webSearchMaxBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", keenableSearchError(response.StatusCode, content, apiKey != "")
	}
	return parseKeenableSearch(query, content)
}

// keenableSearchError names the documented failures (https://docs.keenable.ai/errors)
// and keeps the API's own message for the rest.
func keenableSearchError(status int, content []byte, keyed bool) error {
	var body struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(content, &body)
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("Keenable search returned HTTP %d: check KEENABLE_API_KEY", status)
	case http.StatusPaymentRequired:
		return fmt.Errorf("Keenable search returned HTTP %d: no credits left", status)
	case http.StatusTooManyRequests:
		if !keyed {
			return fmt.Errorf("Keenable search returned HTTP %d: keyless rate limit reached; set KEENABLE_API_KEY for higher limits", status)
		}
		return fmt.Errorf("Keenable search returned HTTP %d: rate limited", status)
	}
	if body.Message != "" {
		return fmt.Errorf("Keenable search returned HTTP %d: %s", status, body.Message)
	}
	return fmt.Errorf("Keenable search returned HTTP %d", status)
}

func parseKeenableSearch(query string, content []byte) (string, error) {
	var response struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"results"`
	}
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("decode Keenable search results: %w", err)
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
		output.Results = append(output.Results, WebSearchResult{Title: title, URL: item.URL, Snippet: item.Snippet})
		if len(output.Results) == webSearchMaxResults {
			break
		}
	}
	if len(output.Results) == 0 {
		return "", fmt.Errorf("Keenable web search returned no usable results")
	}
	encoded, err := json.Marshal(output)
	return string(encoded), err
}
