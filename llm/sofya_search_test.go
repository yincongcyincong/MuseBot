package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/utils"
)

// The fixture replaces only the network transport. Production endpoint, headers,
// tool dispatcher and downstream result collection still run.
func mockSofyaSearch(t *testing.T, handle func(http.ResponseWriter, map[string]interface{})) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		handle(w, request)
	}))
	t.Cleanup(server.Close)
	original := http.DefaultTransport
	destination, _ := url.Parse(server.URL)
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.String() != "https://sofya.co/v1/search" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.UserAgent() != "MuseBot/SofyaSearch" || r.Header.Get("Authorization") != "Bearer sofya-test-key" {
			t.Errorf("unexpected caller identity or authentication headers")
		}
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "sofya"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	t.Setenv("SOFYA_API_KEY", "sofya-test-key")
}

const sofyaResultsFixture = `{"query":"first run","answer":null,"results":[{"title":"MCP reference","url":"https://example.com/reference","content":"Useful current information","description":"Useful current information","fetched":false,"published_date":"2026-10-01"},{"title":"","url":"http://example.org/","content":"","description":null},{"title":"x","url":"javascript:alert(1)"},{"title":"again","url":"https://example.com/reference"}],"search_depth":"snippets","credits_used":1,"credits_remaining":1999}`

func TestSofyaSearchTaskDispatcher(t *testing.T) {
	mockSofyaSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
		if request["query"] != "first run" || request["max_results"] != float64(8) || request["search_depth"] != "snippets" {
			t.Errorf("request mapping: %v", request)
		}
		_, _ = w.Write([]byte(sofyaResultsFixture))
	})
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "first run")
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Results) != 1 || state.Results[0].ID != "search-1" {
		t.Fatalf("missing tool result: %+v", state.Results)
	}
	var result utils.WebSearchResponse
	if err := json.Unmarshal([]byte(state.Results[0].Content), &result); err != nil {
		t.Fatal(err)
	}
	if result.Query != "first run" || len(result.Results) != 2 ||
		result.Results[0].URL != "https://example.com/reference" || result.Results[0].Snippet != "Useful current information" ||
		result.Results[1].Title != "http://example.org/" {
		t.Fatalf("unusable downstream result: %+v", result)
	}
}

func TestSofyaSearchErrors(t *testing.T) {
	mockSofyaSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
		switch request["query"] {
		case "invalid key":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"detail":"Invalid API key","error":"unauthorized"}`))
		case "no credits":
			w.WriteHeader(http.StatusPaymentRequired)
		case "rate limited":
			w.WriteHeader(http.StatusTooManyRequests)
		case "http error":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"detail":"Search capacity is temporarily unavailable","error":"unavailable"}`))
		case "malformed":
			_, _ = w.Write([]byte("not JSON"))
		case "no URLs":
			_, _ = w.Write([]byte(`{"results":[{"url":"file:///etc/passwd"}]}`))
		case "large":
			_, _ = w.Write([]byte(strings.Repeat("x", 512*1024+1)))
		}
	})
	for _, query := range []string{"invalid key", "no credits", "rate limited", "http error", "malformed", "no URLs", "large"} {
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), query)
		// search failures are fed back to the LLM as the tool result instead of
		// aborting the request, so the agent can adapt or report them
		if err != nil {
			t.Fatalf("%q returned unexpected error: %v", query, err)
		}
		if len(state.Results) != 1 || !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("%q did not surface failure as tool result: %+v", query, state.Results)
		}
		if strings.Contains(state.Results[0].Content, "sofya-test-key") {
			t.Fatalf("%q leaked the API key into the tool result", query)
		}
	}
}

func TestSofyaSearchCancellation(t *testing.T) {
	mockSofyaSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
		time.Sleep(200 * time.Millisecond)
	})
	for _, timeout := range []time.Duration{0, 30 * time.Millisecond} {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		started := time.Now()
		_, err := runSearchTool(ctx, searchTaskAgent(newSearchTask()), "cancel")
		cancel()
		if err == nil || time.Since(started) > time.Second {
			t.Fatalf("cancellation/timeout not propagated: %v", err)
		}
	}
}

func TestSofyaSearchRejectsMissingKeyAndInvalidQueriesBeforeNetwork(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	conf.ToolsConfInfo.WebSearchProvider = "sofya"
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Error("rejected search made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
	t.Setenv("SOFYA_API_KEY", "")
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "query")
	if err != nil || !strings.Contains(state.Results[0].Content, "SOFYA_API_KEY") {
		t.Fatalf("missing key was not surfaced as tool result: %v %+v", err, state)
	}
	t.Setenv("SOFYA_API_KEY", "sofya-test-key")
	for _, query := range []string{"", "  ", strings.Repeat("界", 257)} {
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), query)
		if err != nil {
			t.Fatalf("invalid query %q returned unexpected error: %v", query, err)
		}
		if !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("invalid query %q did not surface failure as tool result: %+v", query, state.Results)
		}
	}
}

func TestSofyaSearchLive(t *testing.T) {
	if os.Getenv("MUSEBOT_SOFYA_LIVE_TEST") != "1" || os.Getenv("SOFYA_API_KEY") == "" {
		t.Skip("set MUSEBOT_SOFYA_LIVE_TEST=1 and SOFYA_API_KEY for live validation")
	}
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "sofya"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "Go programming language documentation")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state.Results[0].Content)
}
