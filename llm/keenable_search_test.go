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
func mockKeenableSearch(t *testing.T, apiKey string, handle func(http.ResponseWriter, map[string]interface{})) {
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
		wantURL := "https://api.keenable.ai/v1/search/public"
		if apiKey != "" {
			wantURL = "https://api.keenable.ai/v1/search"
		}
		if r.Method != http.MethodPost || r.URL.String() != wantURL {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.UserAgent() != "MuseBot/KeenableSearch" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected caller identity or authorization header")
		}
		if apiKey != "" && (r.Header.Get("X-API-Key") != apiKey || r.Header.Get("X-Keenable-Title") != "") {
			t.Errorf("keyed request headers: %v", r.Header)
		}
		if apiKey == "" && (r.Header.Get("X-API-Key") != "" || r.Header.Get("X-Keenable-Title") != "MuseBot") {
			t.Errorf("keyless request headers: %v", r.Header)
		}
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "keenable"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	t.Setenv("KEENABLE_API_KEY", apiKey)
}

// Shape of a documented 200 from POST /v1/search: `description` is deprecated and empty.
const keenableResultsFixture = `{"query":"first run","mode":"pro","results":[{"title":"MCP reference","url":"https://example.com/reference","snippet":"Useful current information","description":""},{"url":"http://example.org/","snippet":"","description":""},{"title":"bad","url":"javascript:alert(1)"},{"title":"dup","url":"https://example.com/reference"}]}`

func TestKeenableSearchTaskDispatcher(t *testing.T) {
	for _, apiKey := range []string{"", "kn-test-key"} {
		t.Run(fmt.Sprintf("keyed=%v", apiKey != ""), func(t *testing.T) {
			mockKeenableSearch(t, apiKey, func(w http.ResponseWriter, request map[string]interface{}) {
				if request["query"] != "first run" || request["max_results"] != float64(8) || request["snippet_max_length"] != float64(1000) {
					t.Errorf("request mapping: %v", request)
				}
				_, _ = w.Write([]byte(keenableResultsFixture))
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
		})
	}
}

func TestKeenableSearchErrors(t *testing.T) {
	for _, apiKey := range []string{"", "kn-test-key"} {
		t.Run(fmt.Sprintf("keyed=%v", apiKey != ""), func(t *testing.T) {
			mockKeenableSearch(t, apiKey, func(w http.ResponseWriter, request map[string]interface{}) {
				switch request["query"] {
				case "invalid key":
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":"Invalid API key","code":"invalid_api_key","message":"The API key is not valid."}`))
				case "revoked":
					w.WriteHeader(http.StatusForbidden)
				case "no credits":
					w.WriteHeader(http.StatusPaymentRequired)
				case "rate limited":
					w.WriteHeader(http.StatusTooManyRequests)
				case "http error":
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(`{"error":"Service unavailable","message":"try again later"}`))
				case "malformed":
					_, _ = w.Write([]byte("not JSON"))
				case "no URLs":
					_, _ = w.Write([]byte(`{"results":[{"url":"file:///etc/passwd"}]}`))
				case "large":
					_, _ = w.Write([]byte(strings.Repeat("x", 512*1024+1)))
				}
			})
			for _, query := range []string{"invalid key", "revoked", "no credits", "rate limited", "http error", "malformed", "no URLs", "large"} {
				state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), query)
				// search failures are fed back to the LLM as the tool result instead of
				// aborting the request, so the agent can adapt or report them
				if err != nil {
					t.Fatalf("%q returned unexpected error: %v", query, err)
				}
				if len(state.Results) != 1 || !strings.Contains(state.Results[0].Content, "failed") {
					t.Fatalf("%q did not surface failure as tool result: %+v", query, state.Results)
				}
				if strings.Contains(state.Results[0].Content, "kn-test-key") {
					t.Fatalf("%q leaked the API key into the tool result", query)
				}
			}
		})
	}
}

func TestKeenableSearchCancellation(t *testing.T) {
	mockKeenableSearch(t, "", func(w http.ResponseWriter, request map[string]interface{}) {
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

func TestKeenableSearchRejectsInvalidQueriesBeforeNetwork(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	conf.ToolsConfInfo.WebSearchProvider = "keenable"
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Error("rejected search made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
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

func TestKeenableSearchLive(t *testing.T) {
	if os.Getenv("MUSEBOT_KEENABLE_LIVE_TEST") != "1" {
		t.Skip("set MUSEBOT_KEENABLE_LIVE_TEST=1 for live validation (KEENABLE_API_KEY optional)")
	}
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "keenable"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "Go programming language documentation")
	if err != nil {
		t.Fatal(err)
	}
	var result utils.WebSearchResponse
	if len(state.Results) != 1 || json.Unmarshal([]byte(state.Results[0].Content), &result) != nil || len(result.Results) == 0 {
		t.Fatalf("live search returned no usable results: %+v", state.Results)
	}
}
