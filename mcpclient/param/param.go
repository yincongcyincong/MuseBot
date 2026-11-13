// Package param holds the MCP server configuration types.
//
// The JSON shapes are a wire and disk contract: they are the mcp.json format
// persisted by the bot, exchanged between the admin service and the bot over
// HTTP, and rendered by the admin frontend. They must stay compatible with
// previously written configurations.
package param

// client types (ClientType field of MCPClientConf)
const (
	SSEType      = "sse"
	StdioType    = "stdio"
	HTTPStreamer = "http-streamer"
)

// config file types (Type field of MCPConfig)
const (
	HTTPConfigType  = "http"
	StdioConfigType = "stdio"
)

// McpClientGoConfig is the root document of the mcp.json config file.
type McpClientGoConfig struct {
	McpServers map[string]*MCPConfig `json:"mcpServers"`
}

// MCPConfig describes one MCP server: a stdio command (Command/Args/Env) or a
// remote endpoint (Url, with optional Headers and OAuth).
type MCPConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	Url     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`

	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`

	Disabled bool `json:"disabled,omitempty"`

	OAuth *OAuthConfig `json:"oauth,omitempty"`
}

// OAuthConfig configures OAuth for streamable HTTP servers using the
// client-credentials grant.
type OAuthConfig struct {
	ClientID string `json:"client_id,omitEmpty"` // note: option spelled as in the original format; do not "fix"
	// ClientSecret is the OAuth client secret.
	ClientSecret string `json:"client_secret,omitempty"`
	// Scopes are the OAuth scopes to request.
	Scopes []string `json:"scopes,omitempty"`
	// AuthServerMetadataURL is the RFC 8414 metadata document URL; the token
	// endpoint is discovered from it.
	AuthServerMetadataURL string `json:"auth_server_metadata_url,omitempty"`
	// RedirectURL and PKCEEnabled belong to the authorization-code flow, which
	// a headless bot cannot run; kept for config compatibility, ignored at
	// runtime.
	RedirectURL string `json:"redirect_url,omitempty"`
	PKCEEnabled bool   `json:"pkce_enabled,omitempty"`
}

// MCPClientConf is the resolved, ready-to-connect form of one MCPConfig.
type MCPClientConf struct {
	Name        string
	Description string
	ClientType  string // StdioType | SSEType | HTTPStreamer

	StdioConf *StdioClientConfig
	URL       string
	Headers   map[string]string
	OAuth     *OAuthConfig
}

// StdioClientConfig carries the command line of a stdio server.
type StdioClientConfig struct {
	Command string
	Env     []string
	Args    []string
}
