package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/yincongcyincong/MuseBot/metrics"
)

// TestNewPProfServer checks that the server is created with the correct address.
func TestNewPProfServer(t *testing.T) {
	// Case: with custom port
	server := NewHTTPServer(":12345")
	if server.Addr != ":12345" {
		t.Errorf("Expected address :12345, got %s", server.Addr)
	}

	// Case: with empty string (should fallback to :36060)
	server = NewHTTPServer("")
	if server.Addr != ":36060" {
		t.Errorf("Expected default address :36060, got %s", server.Addr)
	}
}

// TestPProfServer_Start starts the server and checks the /metrics endpoint.
func TestPProfServer_Start(t *testing.T) {
	metrics.RegisterMetrics()

	addr := ":18182"
	server := NewHTTPServer(addr)
	server.Start()

	// wait for the server to start
	time.Sleep(300 * time.Millisecond)

	resp, err := http.Get("http://localhost" + addr + "/metrics")
	if err != nil {
		t.Fatalf("Failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if len(body) == 0 {
		t.Error("Expected non-empty /metrics response")
	}
}

func TestHTTPMetricsUseRouteAndPreserveFlushing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test-stream/", func(w http.ResponseWriter, r *http.Request) {
		w.(http.Flusher).Flush()
	})
	handler := WithRequestContext(mux)

	for _, test := range []struct {
		path, route, status string
	}{
		{"/test-stream/token-one", "/test-stream/", "200"},
		{"/test-stream/token-two", "/test-stream/", "200"},
		{"/unknown-test-path", "unmatched", "404"},
	} {
		counter := metrics.HTTPRequests.WithLabelValues(test.route, test.status)
		before := testutil.ToFloat64(counter)
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, test.path, nil))
		if got := testutil.ToFloat64(counter); got != before+1 {
			t.Errorf("route %q status %q: count = %v, want %v", test.route, test.status, got, before+1)
		}
	}
}
