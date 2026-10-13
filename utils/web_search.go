package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	webSearchMaxQueryRunes   = 256
	webSearchMaxResults      = 8
	webSearchMaxBytes        = 512 * 1024
	webSearchTimeout         = 45 * time.Second
	webSearchPrimaryTimeout  = 10 * time.Second
	webSearchFallbackTimeout = 30 * time.Second
)

var (
	webSearchURL        = "https://lite.duckduckgo.com/lite/"
	bingWebSearchURL    = "https://www.bing.com/search?mkt=en-US&setlang=en-US&q="
	bingWebSearchResult = regexp.MustCompile(`(?is)<h2[^>]*>\s*<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	bingWebSearchNote   = regexp.MustCompile(`(?is)<p[^>]+class=["'][^"']*b_lineclamp[^"']*["'][^>]*>(.*?)</p>`)
)

type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

type WebSearchResponse struct {
	Query   string            `json:"query"`
	Results []WebSearchResult `json:"results"`
}

func WebSearch(ctx context.Context, query string) (string, error) {
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

	content, err := duckDuckGoWebSearch(ctx, query)
	if err == nil {
		results := parseWebSearchHTML(content, webSearchMaxResults)
		if len(results) > 0 {
			encoded, encodeErr := json.Marshal(WebSearchResponse{Query: query, Results: results})
			if encodeErr != nil {
				return "", encodeErr
			}
			return string(encoded), nil
		}
	}

	return bingWebSearch(ctx, query)
}

func duckDuckGoWebSearch(ctx context.Context, query string) (string, error) {
	requestCtx, cancel := context.WithTimeout(ctx, webSearchPrimaryTimeout)
	defer cancel()

	form := url.Values{"q": {query}}
	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, webSearchURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MuseBot web search)")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	httpClient := &http.Client{Timeout: webSearchPrimaryTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("web search returned HTTP %d", response.StatusCode)
	}

	content, err := io.ReadAll(io.LimitReader(response.Body, webSearchMaxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(content)) > webSearchMaxBytes {
		return "", fmt.Errorf("web search response exceeds %d bytes", webSearchMaxBytes)
	}
	return string(content), nil
}

func bingWebSearch(ctx context.Context, query string) (string, error) {
	requestCtx, cancel := context.WithTimeout(ctx, webSearchFallbackTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, bingWebSearchURL+url.QueryEscape(query), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MuseBot web search)")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	httpClient := &http.Client{Timeout: webSearchFallbackTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("fallback web search returned HTTP %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, webSearchMaxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(content)) > webSearchMaxBytes {
		return "", fmt.Errorf("fallback web search response exceeds %d bytes", webSearchMaxBytes)
	}

	results := parseBingWebSearchHTML(string(content), webSearchMaxResults)
	if len(results) == 0 {
		return "", fmt.Errorf("web search returned no results")
	}
	encoded, err := json.Marshal(WebSearchResponse{Query: query, Results: results})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

var (
	webSearchLinkReg     = regexp.MustCompile(`(?is)<a[^>]+class=["']?result-link["']?[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>|<a[^>]+href=["']([^"']+)["'][^>]*class=["']?result-link["']?[^>]*>(.*?)</a>`)
	webSearchSnippetReg  = regexp.MustCompile(`(?is)<td[^>]+class=['"]result-snippet['"][^>]*>(.*?)</td>`)
	webSearchTagReg      = regexp.MustCompile(`(?s)<[^>]*>`)
	webSearchSpacesReg   = regexp.MustCompile(`\s+`)
	webSearchRedirectReg = regexp.MustCompile(`[?&]uddg=([^&]+)`)
)

func parseWebSearchHTML(content string, limit int) []WebSearchResult {
	links := findWebSearchLinks(content)
	snippets := webSearchSnippetReg.FindAllStringSubmatch(content, -1)
	results := make([]WebSearchResult, 0, limit)
	seen := make(map[string]bool)

	for index, link := range links {
		if link.URL == "" || !strings.HasPrefix(link.URL, "http") || seen[link.URL] {
			continue
		}
		seen[link.URL] = true
		result := WebSearchResult{Title: stripWebSearchHTML(link.Title), URL: link.URL}
		if index < len(snippets) {
			result.Snippet = stripWebSearchHTML(snippets[index][1])
		}
		if result.Title == "" {
			continue
		}
		results = append(results, result)
		if len(results) >= limit {
			break
		}
	}
	return results
}

type webSearchLink struct {
	URL   string
	Title string
}

func parseBingWebSearchHTML(content string, limit int) []WebSearchResult {
	links := findBingWebSearchLinks(content)
	snippets := bingWebSearchNote.FindAllStringSubmatch(content, -1)
	results := make([]WebSearchResult, 0, limit)
	seen := make(map[string]bool)
	for index, link := range links {
		if link.URL == "" || !strings.HasPrefix(link.URL, "http") || seen[link.URL] {
			continue
		}
		seen[link.URL] = true
		result := WebSearchResult{Title: stripWebSearchHTML(link.Title), URL: link.URL}
		if index < len(snippets) {
			result.Snippet = stripWebSearchHTML(snippets[index][1])
		}
		if result.Title == "" {
			continue
		}
		results = append(results, result)
		if len(results) >= limit {
			break
		}
	}
	return results
}

func findBingWebSearchLinks(content string) []webSearchLink {
	matches := bingWebSearchResult.FindAllStringSubmatch(content, -1)
	links := make([]webSearchLink, 0, len(matches))
	for _, match := range matches {
		links = append(links, webSearchLink{
			URL:   decodeBingRedirectURL(html.UnescapeString(match[1])),
			Title: html.UnescapeString(match[2]),
		})
	}
	return links
}

func decodeBingRedirectURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() != "www.bing.com" || parsed.Path != "/ck/a" {
		return rawURL
	}
	encoded := strings.TrimPrefix(parsed.Query().Get("u"), "a1")
	if encoded == "" {
		return rawURL
	}
	for _, encoding := range []*base64.Encoding{base64.RawURLEncoding, base64.StdEncoding, base64.RawStdEncoding} {
		if decoded, decodeErr := encoding.DecodeString(encoded); decodeErr == nil {
			decodedURL := string(decoded)
			if strings.HasPrefix(decodedURL, "http") {
				return decodedURL
			}
		}
	}
	return rawURL
}

func findWebSearchLinks(content string) []webSearchLink {
	matches := webSearchLinkReg.FindAllStringSubmatch(content, -1)
	links := make([]webSearchLink, 0, len(matches))
	for _, match := range matches {
		rawURL, title := match[1], match[2]
		if rawURL == "" {
			rawURL, title = match[3], match[4]
		}
		parsed, err := url.Parse(html.UnescapeString(rawURL))
		if err == nil && parsed.Query().Has("uddg") {
			if redirect := webSearchRedirectReg.FindStringSubmatch(rawURL); redirect != nil {
				if decoded, decodeErr := url.QueryUnescape(redirect[1]); decodeErr == nil {
					rawURL = decoded
				}
			}
		}
		links = append(links, webSearchLink{URL: html.UnescapeString(rawURL), Title: html.UnescapeString(title)})
	}
	return links
}

func stripWebSearchHTML(value string) string {
	value = webSearchTagReg.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return webSearchSpacesReg.ReplaceAllString(strings.TrimSpace(value), " ")
}
