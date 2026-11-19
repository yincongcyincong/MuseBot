package utils

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

const maxMediaBytes = 64 << 20

var mediaHTTPClient = &http.Client{
	Timeout: 45 * time.Second,
	// Connect to the validated address directly. A proxy could resolve the
	// hostname itself and bypass the destination check.
	Transport: &http.Transport{
		DialContext:         dialPublicMedia,
		ForceAttemptHTTP2:   true,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConns:        32,
		TLSHandshakeTimeout: 10 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many media redirects")
		}
		return validateMediaURL(req.URL)
	},
}

var nonPublicMediaRanges = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
}

func publicMediaIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.Zone() != "" {
		return false
	}
	for _, prefix := range nonPublicMediaRanges {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func validateMediaURL(u *url.URL) error {
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return errors.New("media URL must use HTTP or HTTPS without credentials")
	}
	return nil
}

func dialPublicMedia(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if !publicMediaIP(ip) {
			return nil, errors.New("media host resolves to a non-public address")
		}
	}
	dialer := net.Dialer{Timeout: 10 * time.Second}
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		err = dialErr
	}
	if err == nil {
		err = errors.New("media host has no IP addresses")
	}
	return nil, err
}

func DownloadFile(rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New("invalid media URL")
	}
	if err := validateMediaURL(u); err != nil {
		return nil, err
	}
	resp, err := mediaHTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to download file: " + resp.Status)
	}
	if resp.ContentLength > maxMediaBytes {
		return nil, errors.New("media exceeds size limit")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxMediaBytes+1))
	if err == nil && len(data) > maxMediaBytes {
		return nil, errors.New("media exceeds size limit")
	}
	return data, err
}
