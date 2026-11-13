package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/yincongcyincong/MuseBot/mcpclient/param"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// newOAuthHandler returns an SDK OAuthHandler for streamable HTTP servers
// using the client-credentials grant, or nil when no oauth is configured.
func newOAuthHandler(conf *param.OAuthConfig) auth.OAuthHandler {
	if conf == nil || conf.AuthServerMetadataURL == "" {
		return nil
	}
	return &clientCredentialsHandler{conf: conf}
}

// clientCredentialsHandler discovers the token endpoint from the RFC 8414
// metadata document once and keeps a single long-lived token source; a 401/403
// drops it so the retried request fetches a fresh token.
type clientCredentialsHandler struct {
	conf *param.OAuthConfig

	mu       sync.Mutex
	endpoint string
	source   oauth2.TokenSource
}

// compile-time interface check
var _ auth.OAuthHandler = (*clientCredentialsHandler)(nil)

// TokenSource implements auth.OAuthHandler.
func (h *clientCredentialsHandler) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.resolveEndpointLocked(ctx); err != nil {
		return nil, err
	}
	if h.source == nil {
		h.source = h.newSourceLocked()
	}
	return h.source, nil
}

// Authorize implements auth.OAuthHandler: invalidate the cached token so the
// retried request obtains a fresh one.
func (h *clientCredentialsHandler) Authorize(_ context.Context, _ *http.Request, resp *http.Response) error {
	defer resp.Body.Close()
	h.mu.Lock()
	h.source = nil
	h.mu.Unlock()
	return nil
}

// resolveEndpointLocked fetches the authorization server metadata once; mu
// must be held.
func (h *clientCredentialsHandler) resolveEndpointLocked(ctx context.Context) error {
	if h.endpoint != "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.conf.AuthServerMetadataURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch oauth metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch oauth metadata: status %s", resp.Status)
	}

	var metadata struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return fmt.Errorf("decode oauth metadata: %w", err)
	}
	if metadata.TokenEndpoint == "" {
		return fmt.Errorf("oauth metadata %s has no token_endpoint", h.conf.AuthServerMetadataURL)
	}
	h.endpoint = metadata.TokenEndpoint
	return nil
}

// newSourceLocked builds the client-credentials token source; mu must be held.
// The token lifecycle is detached from any request context so tokens outlive
// the calls that fetched them.
func (h *clientCredentialsHandler) newSourceLocked() oauth2.TokenSource {
	conf := clientcredentials.Config{
		ClientID:     h.conf.ClientID,
		ClientSecret: h.conf.ClientSecret,
		Scopes:       h.conf.Scopes,
		TokenURL:     h.endpoint,
	}
	return conf.TokenSource(context.Background())
}
