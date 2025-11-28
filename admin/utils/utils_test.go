package utils

import (
	"net/http"
	"testing"
	
	"github.com/yincongcyincong/MuseBot/admin/db"
)

// 测试 NormalizeAddress 函数
func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"example.com", "http://example.com"},
		{"http://example.com", "http://example.com"},
		{"https://secure.com", "https://secure.com"},
	}
	
	for _, tt := range tests {
		got := NormalizeAddress(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeAddress(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

// 测试 ParseCommand 函数
func TestParseCommand(t *testing.T) {
	input := "-a=1 -b=2 -c=hello"
	want := map[string]string{"a": "1", "b": "2", "c": "hello"}
	
	got := ParseCommand(input)
	if len(got) != len(want) {
		t.Fatalf("expected map of len %d, got %d", len(want), len(got))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("expected key %q value %q, got %q", k, v, got[k])
		}
	}
}

// 测试 GetCrtClient 函数 (包括各种分支)
func TestGetCrtClient(t *testing.T) {
	t.Run("empty cert files", func(t *testing.T) {
		bot := &db.Bot{}
		client := GetCrtClient(bot)
		if client == nil {
			t.Fatal("expected non-nil http.Client")
		}
		if client.Transport == nil {
			t.Fatal("expected non-nil transport")
		}
	})
	
	t.Run("invalid cert content", func(t *testing.T) {
		bot := &db.Bot{
			KeyFile: "invalid-key",
			CrtFile: "invalid-cert",
			CaFile:  "invalid-ca",
		}
		client := GetCrtClient(bot)
		if client == nil {
			t.Fatal("expected non-nil client even if cert is invalid")
		}
		// 即使证书无效，也应该返回 client
		if _, ok := client.Transport.(*http.Transport); !ok {
			t.Errorf("expected http.Transport type")
		}
	})
	
	t.Run("valid empty tls config", func(t *testing.T) {
		bot := &db.Bot{}
		client := GetCrtClient(bot)
		tr, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("transport should be *http.Transport")
		}
		tlsCfg := tr.TLSClientConfig
		if tlsCfg == nil {
			t.Fatal("expected non-nil tls.Config")
		}
	})
}
