package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const specstreamSearchURL = "https://api.specstream.ai/v1/web-search"

// SpecStreamWebSearch calls the SpecStream Search API with the key from
// SPECSTREAM_API_KEY. It keeps the built-in web_search response contract, so
// task models receive source URLs and snippets.
func SpecStreamWebSearch(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if utf8.RuneCountInString(query) > webSearchMaxQueryRunes {
		return "", fmt.Errorf("query exceeds %d characters", webSearchMaxQueryRunes)
	}
	apiKey := strings.TrimSpace(os.Getenv("SPECSTREAM_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("SPECSTREAM_API_KEY is not set")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	params := url.Values{}
	params.Set("q", query)
	params.Set("num", strconv.Itoa(webSearchMaxResults))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, specstreamSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "MuseBot/SpecStreamSearch")

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
		return "", fmt.Errorf("SpecStream search response exceeds %d bytes", webSearchMaxBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", specstreamSearchError(response.StatusCode, content)
	}
	return parseSpecStreamSearch(query, content)
}

// specstreamSearchError names the common failures (401 invalid key, 402 no
// balance left, 403 over the key's spending limit, 503 busy or rate limited)
// and keeps the API's own message for the rest.
func specstreamSearchError(status int, content []byte) error {
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(content, &body)
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("SpecStream search returned HTTP %d: check SPECSTREAM_API_KEY", status)
	case http.StatusPaymentRequired:
		return fmt.Errorf("SpecStream search returned HTTP %d: no balance left", status)
	case http.StatusForbidden:
		return fmt.Errorf("SpecStream search returned HTTP %d: over the key's spending limit", status)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("SpecStream search returned HTTP %d: busy or rate limited, try again later", status)
	}
	if body.Error.Message != "" {
		return fmt.Errorf("SpecStream search returned HTTP %d: %s", status, body.Error.Message)
	}
	return fmt.Errorf("SpecStream search returned HTTP %d", status)
}

func parseSpecStreamSearch(query string, content []byte) (string, error) {
	var response struct {
		Data struct {
			Sources []struct {
				URL     string `json:"url"`
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
			} `json:"sources"`
		} `json:"data"`
	}
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("decode SpecStream search results: %w", err)
	}
	output := WebSearchResponse{Query: query, Results: make([]WebSearchResult, 0)}
	seen := make(map[string]bool)
	for _, item := range response.Data.Sources {
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
		return "", fmt.Errorf("SpecStream web search returned no usable results")
	}
	encoded, err := json.Marshal(output)
	return string(encoded), err
}
