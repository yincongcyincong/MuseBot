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
	"sync"
	"testing"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type searchRoundTripper func(*http.Request) (*http.Response, error)

func (f searchRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// The fixture replaces only the network transport. Production endpoint, headers,
// MCP SDK, resolver, tool dispatcher and downstream result collection still run.
func mockParallelSearch(t *testing.T, call func(http.ResponseWriter, *http.Request, map[string]interface{})) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var result interface{}
		switch request["method"] {
		case "initialize":
			result = map[string]interface{}{"protocolVersion": "2025-03-26", "capabilities": map[string]interface{}{}, "serverInfo": map[string]string{"name": "fixture", "version": "1"}}
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
			return
		case "tools/list":
			result = map[string]interface{}{"tools": []interface{}{map[string]interface{}{"name": "web_search", "inputSchema": map[string]interface{}{"type": "object"}}}}
		case "tools/call":
			call(w, r, request)
			return
		default:
			t.Errorf("unexpected method: %v", request["method"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request["id"], "result": result})
	}))
	t.Cleanup(server.Close)
	original := http.DefaultTransport
	destination, _ := url.Parse(server.URL)
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://search.parallel.ai/mcp" {
			t.Errorf("unexpected destination: %s", r.URL)
		}
		if r.UserAgent() != "MuseBot/ParallelSearch" || r.Header.Get("Authorization") != "" || r.Header.Get("x-api-key") != "" {
			t.Errorf("unexpected caller identity or authentication headers")
		}
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "parallel"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	// Deliberately hostile ambient keys must not be sent to the anonymous server.
	t.Setenv("PARALLEL_API_KEY", "ambient-key-must-not-be-used")
	t.Setenv("OPENAI_API_KEY", "ambient-model-key-must-not-be-used")
}

func searchFixtureResponse(w http.ResponseWriter, request map[string]interface{}, text string, isError bool) {
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request["id"], "result": map[string]interface{}{
		"content": []interface{}{map[string]interface{}{"type": "text", "text": text}}, "isError": isError,
	}})
}

const parallelResultsFixture = `{"results":[{"title":"MCP reference","url":"https://example.com/reference","excerpts":["Useful current information","A second excerpt"]},{"url":"http://example.org/"},{"url":"javascript:alert(1)"},{"url":"https://example.com/reference"}]}`

func runSearchTool(ctx context.Context, l *LLM, query string) (*LLMToolState, error) {
	arguments, _ := json.Marshal(map[string]string{"query": query})
	state := &LLMToolState{Calls: []LLMToolCall{{ID: "search-1", Name: conf.BuiltinWebSearchTool, Arguments: string(arguments)}}}
	return state, state.execute(ctx, l)
}

func newSearchTask() *LLMTaskReq {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi}})
	return &LLMTaskReq{Ctx: ctx, Cs: new(param.ContextState)}
}

func searchTaskAgent(d *LLMTaskReq) *LLM {
	return d.newTaskLLM("research", &conf.AgentInfo{IsBuiltin: true}, false)
}

func TestParallelSearchTaskDispatcher(t *testing.T) {
	var mu sync.Mutex
	sessions := make(map[string][]string)
	mockParallelSearch(t, func(w http.ResponseWriter, r *http.Request, request map[string]interface{}) {
		params := request["params"].(map[string]interface{})
		if params["name"] != "web_search" {
			t.Errorf("wrong tool: %v", params["name"])
		}
		args := params["arguments"].(map[string]interface{})
		query := args["objective"].(string)
		queries := args["search_queries"].([]interface{})
		if len(queries) != 1 || queries[0] != query {
			t.Errorf("query mapping: %v", args)
		}
		id, _ := args["session_id"].(string)
		if len(id) != 36 {
			t.Errorf("not an opaque UUID: %q", id)
		}
		mu.Lock()
		sessions[query] = append(sessions[query], id)
		mu.Unlock()
		searchFixtureResponse(w, request, parallelResultsFixture, false)
	})
	a, b := newSearchTask(), newSearchTask()
	// Different agents and follow-up searches in a run share identity. Independent
	// and restarted research runs remain isolated, including interleaved calls.
	for _, task := range []*LLMTaskReq{a, b, a, b, newSearchTask()} {
		query := "first run"
		if task == b {
			query = "second run"
		} else if task != a {
			query = "new run"
		}
		state, err := runSearchTool(task.Ctx, searchTaskAgent(task), query)
		if err != nil {
			t.Fatal(err)
		}
		var result utils.WebSearchResponse
		if len(state.Results) != 1 || state.Results[0].ID != "search-1" {
			t.Fatalf("missing tool result: %+v", state.Results)
		}
		if err := json.Unmarshal([]byte(state.Results[0].Content), &result); err != nil {
			t.Fatal(err)
		}
		if result.Query != query || len(result.Results) != 2 || result.Results[0].URL != "https://example.com/reference" || !strings.Contains(result.Results[0].Snippet, "second excerpt") {
			t.Fatalf("unusable downstream result: %+v", result)
		}
	}
	if sessions["first run"][0] != sessions["first run"][1] || sessions["second run"][0] != sessions["second run"][1] || sessions["first run"][0] == sessions["second run"][0] || sessions["first run"][0] == sessions["new run"][0] {
		t.Fatalf("incorrect run identity ownership: %v", sessions)
	}
	var wg sync.WaitGroup
	for _, task := range []*LLMTaskReq{a, b} {
		wg.Add(1)
		go func(d *LLMTaskReq) {
			defer wg.Done()
			// Construction stays outside the goroutine; only execution is concurrent.
			l := &LLM{Ctx: d.Ctx, searchSession: &d.searchSession}
			if _, err := runSearchTool(d.Ctx, l, "concurrent"); err != nil {
				t.Error(err)
			}
		}(task)
	}
	wg.Wait()
	if len(sessions["concurrent"]) != 2 || sessions["concurrent"][0] == sessions["concurrent"][1] {
		t.Fatal("concurrent research runs shared an identity")
	}
}

func TestParallelSearchErrorsAndRetryIdentity(t *testing.T) {
	var sessions []string
	mockParallelSearch(t, func(w http.ResponseWriter, r *http.Request, request map[string]interface{}) {
		args := request["params"].(map[string]interface{})["arguments"].(map[string]interface{})
		sessions = append(sessions, args["session_id"].(string))
		switch args["objective"] {
		case "quota":
			searchFixtureResponse(w, request, "Free anonymous quota exceeded", true)
		case "http error":
			w.WriteHeader(http.StatusServiceUnavailable)
		case "rpc error":
			_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%v,"error":{"code":-32000,"message":"unavailable"}}`, request["id"])
		case "malformed":
			searchFixtureResponse(w, request, "not JSON", false)
		case "no URLs":
			searchFixtureResponse(w, request, `{"results":[{"url":"file:///etc/passwd"}]}`, false)
		case "large":
			searchFixtureResponse(w, request, strings.Repeat("x", 512*1024+1), false)
		default:
			searchFixtureResponse(w, request, parallelResultsFixture, false)
		}
	})
	task := newSearchTask()
	for _, query := range []string{"quota", "http error", "rpc error", "malformed", "no URLs", "large"} {
		state, err := runSearchTool(task.Ctx, searchTaskAgent(task), query)
		// search failures are fed back to the LLM as the tool result instead of
		// aborting the request, so the agent can adapt or report them
		if err != nil {
			t.Fatalf("%q returned unexpected error: %v", query, err)
		}
		if len(state.Results) != 1 || !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("%q did not surface failure as tool result: %+v", query, state.Results)
		}
	}
	if _, err := runSearchTool(task.Ctx, searchTaskAgent(task), "retry"); err != nil {
		t.Fatal(err)
	}
	for _, id := range sessions {
		if id != sessions[0] {
			t.Fatal("retry/quota changed research identity")
		}
	}
}

func TestParallelSearchCancellation(t *testing.T) {
	mockParallelSearch(t, func(w http.ResponseWriter, r *http.Request, request map[string]interface{}) {
		<-r.Context().Done()
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

func TestWebSearchDisabledAndInvalidProvider(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Fatal("disabled/invalid search made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
	for _, provider := range []string{"disabled", "misspelled"} {
		conf.ToolsConfInfo.WebSearchProvider = provider
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "query")
		// provider errors are fed back to the LLM instead of aborting the request
		if err != nil {
			t.Fatalf("%q search returned unexpected error: %v", provider, err)
		}
		if !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("%q search did not surface failure as tool result: %+v", provider, state.Results)
		}
	}
}

func TestWebSearchDefaultProviderIsPreserved(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<a class="result-link" href="https://example.com/default">Default source</a><td class="result-snippet">Original search backend</td>`))
	}))
	t.Cleanup(server.Close)
	destination, _ := url.Parse(server.URL)
	calls := 0
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "lite.duckduckgo.com" {
			t.Errorf("default search changed provider: %s", r.URL)
		}
		calls++
		copy := r.Clone(r.Context())
		copy.URL.Scheme, copy.URL.Host = destination.Scheme, destination.Host
		return original.RoundTrip(copy)
	})
	for _, provider := range []string{"", "default"} {
		conf.ToolsConfInfo.WebSearchProvider = provider
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "default search")
		if err != nil || !strings.Contains(state.Results[0].Content, "Original search backend") {
			t.Fatalf("default search failed: %v", err)
		}
	}
	if calls != 2 {
		t.Fatalf("default search issued %d calls, expected 2", calls)
	}
}

func TestParallelSearchRejectsInvalidQueriesBeforeNetwork(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	original := http.DefaultTransport
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous; http.DefaultTransport = original })
	conf.ToolsConfInfo.WebSearchProvider = "parallel"
	http.DefaultTransport = searchRoundTripper(func(r *http.Request) (*http.Response, error) {
		t.Error("invalid query made a network request")
		return nil, fmt.Errorf("unexpected network")
	})
	for _, query := range []string{"", "  ", strings.Repeat("界", 257)} {
		state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), query)
		// invalid queries are fed back to the LLM so it can retry with a corrected query
		if err != nil {
			t.Fatalf("invalid query %q returned unexpected error: %v", query, err)
		}
		if !strings.Contains(state.Results[0].Content, "failed") {
			t.Fatalf("invalid query %q did not surface failure as tool result: %+v", query, state.Results)
		}
	}
}

func TestParallelSearchLive(t *testing.T) {
	if os.Getenv("MUSEBOT_PARALLEL_LIVE_TEST") != "1" {
		t.Skip("set MUSEBOT_PARALLEL_LIVE_TEST=1 for anonymous MCP validation")
	}
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "parallel"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })
	state, err := runSearchTool(context.Background(), searchTaskAgent(newSearchTask()), "Go programming language documentation")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state.Results[0].Content)
}
