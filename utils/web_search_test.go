package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebSearchParsesResults(t *testing.T) {
	bingContent := `<h2><a href="https://www.bing.com/ck/a?!&amp;u=a1aHR0cHM6Ly9leGFtcGxlLmNvbS90d28&amp;ntb=1">Bing Example</a></h2><p class="b_lineclamp2">Fallback result</p>`
	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(bingContent))
	}))
	defer fallbackServer.Close()
	previousBingURL := bingWebSearchURL
	bingWebSearchURL = fallbackServer.URL + "?q="
	t.Cleanup(func() { bingWebSearchURL = previousBingURL })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.FormValue("query") != "mcp function calling" && r.FormValue("q") != "mcp function calling" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`<a class='result-link' href="https://example.com/one">Example</a><td class='result-snippet'>A useful &amp; recent result</td>`))
	}))
	defer server.Close()
	previousURL := webSearchURL
	webSearchURL = server.URL
	t.Cleanup(func() { webSearchURL = previousURL })

	response, err := WebSearch(context.Background(), "mcp function calling")
	if err != nil {
		t.Fatalf("WebSearch() error = %v", err)
	}
	if !strings.Contains(response, `"url":"https://example.com/one"`) ||
		!strings.Contains(response, "useful") || !strings.Contains(response, "recent result") {
		t.Fatalf("unexpected search response: %s", response)
	}
}

func TestWebSearchFallsBackToBing(t *testing.T) {
	previousURL := webSearchURL
	webSearchURL = "http://127.0.0.1:1/"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<h2><a href="https://www.bing.com/ck/a?!&amp;u=a1aHR0cHM6Ly9leGFtcGxlLmNvbS90d28&amp;ntb=1">Bing Example</a></h2><p class="b_lineclamp2">Fallback result</p>`))
	}))
	defer server.Close()
	previousBingURL := bingWebSearchURL
	bingWebSearchURL = server.URL + "?q="
	t.Cleanup(func() {
		webSearchURL = previousURL
		bingWebSearchURL = previousBingURL
	})

	response, err := WebSearch(context.Background(), "fallback search")
	if err != nil {
		t.Fatalf("WebSearch() error = %v", err)
	}
	if !strings.Contains(response, `"url":"https://example.com/two"`) || !strings.Contains(response, "Fallback result") {
		t.Fatalf("unexpected fallback response: %s", response)
	}
}

func TestWebSearchRejectsEmptyQuery(t *testing.T) {
	if _, err := WebSearch(context.Background(), "  "); err == nil {
		t.Fatal("WebSearch() expected empty query error")
	}
}
