package http

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManagementAccess(t *testing.T) {
	handler := requireManagementAccess(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	for _, tc := range []struct {
		name, host, peer, origin string
		verified                 bool
		want                     int
	}{
		{"local", "localhost:36060", "127.0.0.1:1234", "", false, 204},
		{"remote", "localhost:36060", "192.0.2.10:1234", "", false, 403},
		{"nonlocal host", "example.com", "127.0.0.1:1234", "", false, 403},
		{"browser origin", "localhost", "127.0.0.1:1234", "https://example.com", false, 403},
		{"verified TLS", "bot.example.com", "192.0.2.10:1234", "", true, 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://"+tc.host+"/", nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("Origin", tc.origin)
			if tc.verified {
				r.TLS = &tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{{}}}}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Errorf("got status %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestValidRagFileName(t *testing.T) {
	if !validRagFileName("部署手册.txt") {
		t.Fatal("expected a plain filename to be valid")
	}
	for _, name := range []string{"", ".", "..", "folder/file.txt", "folder\\file.txt", "/file.txt"} {
		if validRagFileName(name) {
			t.Errorf("unexpectedly accepted non-flat filename %q", name)
		}
	}
}
