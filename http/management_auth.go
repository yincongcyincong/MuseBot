package http

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func requireManagementAccess(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 {
			next(w, r)
			return
		}

		// Trust the direct peer only. Require a local Host and reject browser
		// cross-origin requests to keep local management inaccessible to websites.
		peer, _, _ := net.SplitHostPort(r.RemoteAddr)
		host := r.Host
		if name, _, err := net.SplitHostPort(host); err == nil {
			host = name
		}
		if isLoopback(peer) && (strings.EqualFold(host, "localhost") || isLoopback(host)) &&
			r.Header.Get("Origin") == "" && r.Header.Get("Sec-Fetch-Site") != "cross-site" {
			next(w, r)
			return
		}
		http.Error(w, "Management API requires local access or a verified client certificate", http.StatusForbidden)
	}
}

func isLoopback(host string) bool {
	ip, err := netip.ParseAddr(strings.Trim(host, "[]"))
	return err == nil && ip.Unmap().IsLoopback()
}
