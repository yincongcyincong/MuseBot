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
func mockOutcrawlSearch(t *testing.T, handle func(http.ResponseWriter, map[string]interface{})) {
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
		if r.Method != http.MethodPost || r.URL.String() != "https://api.outcrawl.ai/v1/search" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.UserAgent() != "MuseBot/OutcrawlSearch" || r.Header.Get("Authorization") != "Bearer outcrawl-test-key" {
			t.Errorf("unexpected caller identity or authentication headers")
		}
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "outcrawl"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	t.Setenv("OUTCRAWL_API_KEY", "outcrawl-test-key")
}

const outcrawlResultsFixture = `{"hits":[{"url":"https://example.com/reference","title":"MCP reference","snippet":"Useful current information","rank":1},{"url":"http://example.org/","rank":2},{"url":"javascript:alert(1)","rank":3},{"url":"https://example.com/reference","rank":4}],"usage":{"credits":"0.5"}}`

func TestOutcrawlSearchTaskDispatcher(t *testing.T) {
	mockOutcrawlSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
		if request["query"] != "first run" || request["limit"] != float64(8) {
			t.Errorf("request mapping: %v", request)
		}
		_, _ = w.Write([]byte(outcrawlResultsFixture))
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

func TestOutcrawlSearchErrors(t *testing.T) {
	mockOutcrawlSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
		switch request["query"] {
		case "invalid key":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":"unauthorized","message":"authentication failed"}`))
		case "no credits":
			w.WriteHeader(http.StatusPaymentRequired)
		case "rate limited":
			w.WriteHeader(http.StatusTooManyRequests)
		case "http error":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"code":"unavailable","message":"try again later"}`))
		case "malformed":
			_, _ = w.Write([]byte("not JSON"))
		case "no URLs":
			_, _ = w.Write([]byte(`{"hits":[{"url":"file:///etc/passwd"}]}`))
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
		if strings.Contains(state.Results[0].Content, "outcrawl-test-key") {
			t.Fatalf("%q leaked the API key into the tool result", query)
		}
	}
}

func TestOutcrawlSearchCancellation(t *testing.T) {
	mockOutcrawlSearch(t, func(w http.ResponseWriter, request map[string]interface{}) {
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

func TestOutcrawlSearchRejectsMissingKeyAndInvalidQueriesBeforeNetwork(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	conf.ToolsConfInfo.WebSearchProvider = "outcrawl"
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Error("rejected search made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
	t.Setenv("OUTCRAWL_API_KEY", "")
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "query")
	if err != nil || !strings.Contains(state.Results[0].Content, "OUTCRAWL_API_KEY") {
		t.Fatalf("missing key was not surfaced as tool result: %v %+v", err, state)
	}
	t.Setenv("OUTCRAWL_API_KEY", "outcrawl-test-key")
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

func TestOutcrawlSearchLive(t *testing.T) {
	if os.Getenv("MUSEBOT_OUTCRAWL_LIVE_TEST") != "1" || os.Getenv("OUTCRAWL_API_KEY") == "" {
		t.Skip("set MUSEBOT_OUTCRAWL_LIVE_TEST=1 and OUTCRAWL_API_KEY for live validation")
	}
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "outcrawl"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "Go programming language documentation")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state.Results[0].Content)
}
