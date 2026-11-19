package utils

import (
	"net/netip"
	"testing"
)

func TestPublicMediaIP(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "::1", "10.0.0.1", "169.254.0.1", "100.64.0.1", "::ffff:127.0.0.1"} {
		if publicMediaIP(netip.MustParseAddr(ip)) {
			t.Errorf("accepted non-public IP %s", ip)
		}
	}
	if !publicMediaIP(netip.MustParseAddr("1.1.1.1")) {
		t.Error("rejected public IP")
	}
}
