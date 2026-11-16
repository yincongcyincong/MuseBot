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
func mockSpecStreamSearch(t *testing.T, handle func(http.ResponseWriter, url.Values)) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handle(w, r.URL.Query())
	}))
	t.Cleanup(server.Close)
	original := http.DefaultTransport
	destination, _ := url.Parse(server.URL)
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		endpoint := *r.URL
		endpoint.RawQuery = ""
		if r.Method != http.MethodGet || endpoint.String() != "https://api.specstream.ai/v1/web-search" || r.Body != nil && r.Body != http.NoBody {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.UserAgent() != "MuseBot/SpecStreamSearch" || r.Header.Get("Authorization") != "Bearer specstream-test-key" {
			t.Errorf("unexpected caller identity or authentication headers")
		}
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "specstream"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	t.Setenv("SPECSTREAM_API_KEY", "specstream-test-key")
}

const specstreamResultsFixture = `{"data":{"request_id":"r1","query":"first run","kind":"fact","columns":[],"answer":null,"rows":[],"handoff":null,"took_ms":900,"price_usd":0.005,"sources":[{"url":"https://example.com/reference","title":"MCP reference","site":"example.com","snippet":"Useful current information","cited":true,"published":"2026-10-01"},{"url":"http://example.org/","title":"","site":"example.org","snippet":"","cited":false},{"url":"javascript:alert(1)","title":"x","site":"","snippet":"","cited":false},{"url":"https://example.com/reference","title":"again","site":"example.com","snippet":"","cited":false}]}}`

func TestSpecStreamSearchTaskDispatcher(t *testing.T) {
	mockSpecStreamSearch(t, func(w http.ResponseWriter, query url.Values) {
		if query.Get("q") != "first run" || query.Get("num") != "8" {
			t.Errorf("request mapping: %v", query)
		}
		_, _ = w.Write([]byte(specstreamResultsFixture))
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

func TestSpecStreamSearchErrors(t *testing.T) {
	mockSpecStreamSearch(t, func(w http.ResponseWriter, query url.Values) {
		switch query.Get("q") {
		case "invalid key":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_token","message":"invalid key"}}`))
		case "no balance":
			w.WriteHeader(http.StatusPaymentRequired)
		case "over limit":
			w.WriteHeader(http.StatusForbidden)
		case "busy":
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
		case "http error":
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"error":{"code":"bad_request","message":"q is too long"}}`))
		case "malformed":
			_, _ = w.Write([]byte("not JSON"))
		case "no URLs":
			_, _ = w.Write([]byte(`{"data":{"sources":[{"url":"file:///etc/passwd"}]}}`))
		case "large":
			_, _ = w.Write([]byte(strings.Repeat("x", 512*1024+1)))
		}
	})
	for _, query := range []string{"invalid key", "no balance", "over limit", "busy", "http error", "malformed", "no URLs", "large"} {
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), query)
		// search failures are fed back to the LLM as the tool result instead of
		// aborting the request, so the agent can adapt or report them
		if err != nil {
			t.Fatalf("%q returned unexpected error: %v", query, err)
		}
		if len(state.Results) != 1 || !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("%q did not surface failure as tool result: %+v", query, state.Results)
		}
		if strings.Contains(state.Results[0].Content, "specstream-test-key") {
			t.Fatalf("%q leaked the API key into the tool result", query)
		}
	}
}

func TestSpecStreamSearchCancellation(t *testing.T) {
	mockSpecStreamSearch(t, func(w http.ResponseWriter, query url.Values) {
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

func TestSpecStreamSearchRejectsMissingKeyAndInvalidQueriesBeforeNetwork(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	conf.ToolsConfInfo.WebSearchProvider = "specstream"
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Error("rejected search made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
	t.Setenv("SPECSTREAM_API_KEY", "")
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "query")
	if err != nil || !strings.Contains(state.Results[0].Content, "SPECSTREAM_API_KEY") {
		t.Fatalf("missing key was not surfaced as tool result: %v %+v", err, state)
	}
	t.Setenv("SPECSTREAM_API_KEY", "specstream-test-key")
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

func TestSpecStreamSearchLive(t *testing.T) {
	if os.Getenv("MUSEBOT_SPECSTREAM_LIVE_TEST") != "1" || os.Getenv("SPECSTREAM_API_KEY") == "" {
		t.Skip("set MUSEBOT_SPECSTREAM_LIVE_TEST=1 and SPECSTREAM_API_KEY for live validation")
	}
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "specstream"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "Go programming language documentation")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state.Results[0].Content)
}
