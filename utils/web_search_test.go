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

func TestWebSearchFallsBackWhenDuckDuckGoBlocks(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"bot check status", http.StatusAccepted, `<div class="anomaly-modal">bots use DuckDuckGo too</div>`},
		{"bot check page", http.StatusOK, `<div class="anomaly-modal__title">Unfortunately, bots use DuckDuckGo too.</div>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var bingUA string
			duckServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer duckServer.Close()

			bingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bingUA = r.Header.Get("User-Agent")
				_, _ = w.Write([]byte(`<h2><a href="https://example.com/two">Bing Example</a></h2><p class="b_lineclamp2">Fallback result</p>`))
			}))
			defer bingServer.Close()

			previousURL, previousBingURL := webSearchURL, bingWebSearchURL
			webSearchURL = duckServer.URL
			bingWebSearchURL = bingServer.URL + "?q="
			t.Cleanup(func() {
				webSearchURL = previousURL
				bingWebSearchURL = previousBingURL
			})

			response, err := WebSearch(context.Background(), "blocked search")
			if err != nil {
				t.Fatalf("WebSearch() error = %v", err)
			}
			if !strings.Contains(response, `"url":"https://example.com/two"`) {
				t.Fatalf("expected the Bing fallback to be used, got %s", response)
			}
			// the fallback must not advertise itself as a bot
			if !strings.Contains(bingUA, "Mozilla/5.0") || !strings.Contains(bingUA, "Chrome/") {
				t.Fatalf("bing user agent = %q, want a browser user agent", bingUA)
			}
		})
	}
}

func TestWebSearchSendsBrowserUserAgentToDuckDuckGo(t *testing.T) {
	var duckUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duckUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`<a class='result-link' href="https://example.com/one">Example</a>`))
	}))
	defer server.Close()

	previousURL := webSearchURL
	webSearchURL = server.URL
	t.Cleanup(func() { webSearchURL = previousURL })

	if _, err := WebSearch(context.Background(), "user agent check"); err != nil {
		t.Fatalf("WebSearch() error = %v", err)
	}
	if !strings.Contains(duckUA, "Mozilla/5.0") || !strings.Contains(duckUA, "Chrome/") {
		t.Fatalf("duckduckgo user agent = %q, want a browser user agent", duckUA)
	}
}
