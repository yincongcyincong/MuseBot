// Package clients manages the MCP server connections of one bot process: it
// loads mcp.json, keeps a process-wide registry of connected servers, executes
// tool calls, and keeps the connections alive.
//
// It is built on the official MCP Go SDK (modelcontextprotocol/go-sdk) and
// replaces the previous yincongcyincong/mcp-client-go wrapper.
package clients

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/mcpclient/param"
)

const (
	clientImplName    = "MuseBot"
	clientImplVersion = "1.0.0"
	pingInterval      = 10 * time.Second
	pingTimeout       = 10 * time.Second
	// newProtocolVersion is the SEP-2575 protocol revision from which servers
	// require the per-request _meta envelope (protocolVersion, clientInfo,
	// clientCapabilities). The SDK keeps this constant unexported, so it is
	// mirrored here; MCP protocol versions are ISO dates, so the comparison
	// matches the SDK's usesNewProtocol exactly.
	newProtocolVersion = "2026-07-28"
)

// mcpClients maps server name -> *MCPClient for the whole process.
var mcpClients sync.Map

// MCPClient is one registered MCP server connection. Its session is kept
// alive by a ping loop and transparently rebuilt when the server goes away.
type MCPClient struct {
	Conf *param.MCPClientConf

	mu            sync.RWMutex
	session       *mcp.ClientSession
	tools         []*mcp.Tool
	keepaliveCtx  context.Context
	stopKeepalive context.CancelFunc
	closeStderr   func()
}

// InitByConfFile loads an mcp.json document and resolves every enabled server
// into a connectable MCPClientConf.
func InitByConfFile(configFilePath string) ([]*param.MCPClientConf, error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	config := new(param.McpClientGoConfig)
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	mcs := make([]*param.MCPClientConf, 0, len(config.McpServers))
	for mcpName, mcpConf := range config.McpServers {
		if clientConf := GetOneMCPClient(mcpName, mcpConf); clientConf != nil {
			mcs = append(mcs, clientConf)
		}
	}
	return mcs, nil
}

// GetOneMCPClient resolves one MCPConfig into an MCPClientConf, probing the
// URL to tell the legacy SSE transport from streamable HTTP. It returns nil
// for disabled or unresolvable servers.
func GetOneMCPClient(mcpName string, mcpConf *param.MCPConfig) *param.MCPClientConf {
	if mcpConf == nil || mcpConf.Disabled {
		return nil
	}

	switch getMcpType(mcpConf) {
	case param.StdioConfigType:
		return &param.MCPClientConf{
			Name:        mcpName,
			Description: mcpConf.Description,
			ClientType:  param.StdioType,
			StdioConf: &param.StdioClientConfig{
				Command: mcpConf.Command,
				Env:     changeEnvMapToSlice(mcpConf.Env),
				Args:    mcpConf.Args,
			},
		}
	case param.HTTPConfigType:
		httpType, err := checkSSEOrHTTP(mcpConf.Url)
		if err != nil {
			logger.Error("check mcp url transport fail", "server", mcpName, "url", mcpConf.Url, "err", err)
			return nil
		}
		return &param.MCPClientConf{
			Name:        mcpName,
			Description: mcpConf.Description,
			ClientType:  httpType,
			URL:         mcpConf.Url,
			Headers:     mcpConf.Headers,
			OAuth:       mcpConf.OAuth,
		}
	default:
		logger.Error("mcp type not exist", "server", mcpName, "type", mcpConf.Type)
		return nil
	}
}

// RegisterMCPClient connects every given server concurrently and adds it to
// the registry. It never fails as a whole; per-server errors are returned in
// the map keyed by server name (empty map when everything succeeded).
func RegisterMCPClient(ctx context.Context, params []*param.MCPClientConf) map[string]error {
	errs := make(map[string]error)
	if len(params) == 0 {
		return errs
	}

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)
	for _, clientParam := range params {
		wg.Add(1)
		go func(cp *param.MCPClientConf) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mutex.Lock()
					errs[cp.Name] = fmt.Errorf("panic: %v", r)
					mutex.Unlock()
					logger.Error("mcp client panic", "server", cp.Name, "panic", r)
				}
			}()
			if err := connectMCPClient(ctx, cp); err != nil {
				mutex.Lock()
				errs[cp.Name] = err
				mutex.Unlock()
			}
		}(clientParam)
	}
	wg.Wait()
	return errs
}

// connectMCPClient builds the transport for conf, performs the handshake,
// lists the tools and swaps the registry entry.
func connectMCPClient(ctx context.Context, conf *param.MCPClientConf) error {
	transport, closeStderr, err := buildTransport(conf)
	if err != nil {
		return err
	}

	session, err := connectSession(ctx, transport)
	if err != nil {
		if closeStderr != nil {
			closeStderr()
		}
		return err
	}

	mc := &MCPClient{Conf: conf, session: session, closeStderr: closeStderr}
	tools, err := mc.GetAllTools(ctx)
	if err != nil {
		_ = session.Close()
		if closeStderr != nil {
			closeStderr()
		}
		return err
	}
	mc.tools = tools

	if old, ok := mcpClients.Load(conf.Name); ok {
		old.(*MCPClient).Close()
	}

	mc.keepaliveCtx, mc.stopKeepalive = context.WithCancel(context.Background())
	mcpClients.Store(conf.Name, mc)
	go mc.handlePing()
	logger.Info("mcp client registered", "server", conf.Name, "type", conf.ClientType, "tools", len(tools))
	return nil
}

// connectSession performs the SDK handshake (initialize + version negotiation)
// over the given transport.
func connectSession(ctx context.Context, transport mcp.Transport) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    clientImplName,
		Version: clientImplVersion,
	}, nil)
	return client.Connect(ctx, transport, nil)
}

// buildTransport constructs the SDK transport matching the client type:
// a spawned command for stdio, SSE or streamable HTTP for remote servers.
func buildTransport(conf *param.MCPClientConf) (mcp.Transport, func(), error) {
	switch conf.ClientType {
	case param.StdioType:
		if conf.StdioConf == nil {
			return nil, nil, fmt.Errorf("stdio server %s has no command", conf.Name)
		}
		cmd := exec.Command(conf.StdioConf.Command, conf.StdioConf.Args...)
		cmd.Env = append(os.Environ(), conf.StdioConf.Env...)
		closeStderr, err := captureStderr(cmd, conf.Name)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CommandTransport{Command: cmd}, closeStderr, nil

	case param.SSEType:
		return &mcp.SSEClientTransport{
			Endpoint:   conf.URL,
			HTTPClient: headerHTTPClient(conf.Headers, nil),
		}, nil, nil

	case param.HTTPStreamer:
		return &mcp.StreamableClientTransport{
			Endpoint:     conf.URL,
			HTTPClient:   headerHTTPClient(conf.Headers, nil),
			OAuthHandler: newOAuthHandler(conf.OAuth),
		}, nil, nil

	default:
		return nil, nil, fmt.Errorf("unsupported mcp client type %q", conf.ClientType)
	}
}

// captureStderr routes the stdio server's stderr into the bot log so startup
// failures of spawned servers are visible.
func captureStderr(cmd *exec.Cmd, name string) (func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = w
	go func() {
		defer r.Close()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logger.Info("mcp server stderr", "server", name, "msg", scanner.Text())
		}
	}()
	// closing w stops the scanner loop once the child exits; the child keeps
	// its inherited end until then, so nothing is cut short while running
	return func() { _ = w.Close() }, nil
}

// headerHTTPClient returns an HTTP client that attaches the configured static
// headers to every request (SSE and streamable HTTP servers).
func headerHTTPClient(headers map[string]string, base *http.Client) *http.Client {
	if len(headers) == 0 {
		if base == nil {
			return nil
		}
		return base
	}
	inner := http.DefaultTransport
	if base != nil && base.Transport != nil {
		inner = base.Transport
	}
	return &http.Client{
		Transport: &headerRoundTripper{base: inner, headers: headers},
		Timeout:   0, // MCP sessions are long-lived; per-request contexts bound the calls
	}
}

type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for key, value := range t.headers {
		clone.Header.Set(key, value)
	}
	return t.base.RoundTrip(clone)
}

// GetMCPClient returns the registered server by name.
func GetMCPClient(name string) (*MCPClient, error) {
	v, ok := mcpClients.Load(name)
	if !ok {
		return nil, fmt.Errorf("client %s not found", name)
	}
	return v.(*MCPClient), nil
}

// GetMCPClientByToolName finds the server that exposes the given tool.
// clientByToolName returns the client exposing the tool, or nil.
func clientByToolName(toolName string) *MCPClient {
	var res *MCPClient
	mcpClients.Range(func(_, value any) bool {
		for _, tool := range value.(*MCPClient).Tools() {
			if tool.Name == toolName {
				res = value.(*MCPClient)
				return false
			}
		}
		return true
	})
	return res
}

func GetMCPClientByToolName(toolName string) (*MCPClient, error) {
	res := clientByToolName(toolName)
	if res == nil {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}
	return res, nil
}

// ResolveToolName maps a model-invented mcp__server__tool name onto the
// advertised plain tool name. GLM-style models routinely rename advertised
// tools with the mcp__<server>__<tool> convention even though MuseBot
// advertises the bare MCP tool names; an exact match is returned untouched,
// and so is an unknown name (the lookup error names what the model sent).
func ResolveToolName(toolName string) string {
	if clientByToolName(toolName) != nil {
		return toolName
	}
	// try the suffix after the last double underscore: mcp__time__get_current_time -> get_current_time
	if idx := strings.LastIndex(toolName, "__"); idx >= 0 && idx+2 < len(toolName) {
		if suffix := toolName[idx+2:]; clientByToolName(suffix) != nil {
			return suffix
		}
	}
	return toolName
}

// Tools returns a snapshot of the tools the server currently exposes.
func (m *MCPClient) Tools() []*mcp.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*mcp.Tool(nil), m.tools...)
}

// GetAllTools pages through the server's tools/list.
func (m *MCPClient) GetAllTools(ctx context.Context) ([]*mcp.Tool, error) {
	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("mcp client %s is not connected", m.Conf.Name)
	}

	var all []*mcp.Tool
	var cursor string
	for {
		result, err := session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, err
		}
		all = append(all, result.Tools...)
		if result.NextCursor == "" {
			return all, nil
		}
		cursor = result.NextCursor
	}
}

// ExecTools calls one tool of this server after checking the required
// parameters, and renders the result content as a string for the LLM.
func (m *MCPClient) ExecTools(ctx context.Context, name string, params map[string]interface{}) (string, error) {
	m.mu.RLock()
	session := m.session
	var useTool *mcp.Tool
	for _, tool := range m.tools {
		if tool.Name == name {
			useTool = tool
			break
		}
	}
	m.mu.RUnlock()

	if useTool == nil {
		return "", fmt.Errorf("tool %s not found", name)
	}
	_, required := inputSchemaParts(useTool.InputSchema)
	for _, reqParam := range required {
		if _, ok := params[reqParam]; !ok {
			return "", fmt.Errorf("required parameter %s not found", reqParam)
		}
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: params})
	if err != nil {
		return "", err
	}
	return ResultString(result), nil
}

// handlePing keeps the connection alive: ping every 10s and rebuild the
// session when the server stopped responding.
func (m *MCPClient) handlePing() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("mcp keepalive panic", "server", m.Conf.Name, "panic", r)
		}
	}()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.restartMCPServer()
		case <-m.keepaliveCtx.Done():
			return
		}
	}
}

// pingParams builds the ping request params, carrying the SEP-2575 _meta
// envelope (protocolVersion, clientInfo, clientCapabilities) when the session
// negotiated the >= 2026-07-28 protocol.
//
// The SDK's Ping (v1.8.0) is the one client request that skips the envelope
// injection ListTools/ListPrompts/CallTool perform, and servers that enforce
// the new protocol (e.g. the Python MCP SDK) reject the bare ping. Old-protocol
// sessions keep bare pings: per the spec a _meta carrying protocolVersion opts
// the request into envelope validation, so it must not be sent partially or
// unconditionally.
//
// Note that >= 2026-07-28 servers split on ping: Python keeps the method and
// demands the envelope, while Go SDK servers removed ping entirely and answer
// an enveloped ping with "method not found". restartMCPServer treats that
// response as proof of liveness, so both kinds stay connected.
func pingParams(session *mcp.ClientSession) *mcp.PingParams {
	params := &mcp.PingParams{}
	initResult := session.InitializeResult()
	if initResult == nil || initResult.ProtocolVersion < newProtocolVersion {
		return params
	}
	// mirrors what injectRequestMeta sends for this client: the impl passed to
	// NewClient and the default capabilities marshalled as {"roots":{"listChanged":true}}
	params.Meta = mcp.Meta{
		mcp.MetaKeyProtocolVersion: initResult.ProtocolVersion,
		mcp.MetaKeyClientInfo: &mcp.Implementation{
			Name:    clientImplName,
			Version: clientImplVersion,
		},
		mcp.MetaKeyClientCapabilities: map[string]any{
			"roots": map[string]any{"listChanged": true},
		},
	}
	return params
}

// isMethodNotFound reports whether a ping error is a JSON-RPC "method not
// found" response — the answer sessionless Go SDK servers give because ping
// was removed in the >= 2026-07-28 protocol. Only a responding server can
// produce it, so the connection is demonstrably alive.
func isMethodNotFound(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "method not found")
}

// restartMCPServer rebuilds the session when a ping fails; the registry entry
// (and thus the *MCPClient identity) stays the same.
func (m *MCPClient) restartMCPServer() {
	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()
	if session == nil {
		return
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	err := session.Ping(pingCtx, pingParams(session))
	cancel()
	if err == nil {
		return
	}
	// a "method not found" response means the server answered: ping was removed
	// in the sessionless protocol, so this is the healthy case for those
	// servers, not a dead connection
	if isMethodNotFound(err) {
		logger.Debug("mcp ping not supported by server, connection alive", "server", m.Conf.Name)
		return
	}

	logger.Warn("mcp ping fail, reconnecting", "server", m.Conf.Name, "err", err)
	transport, closeStderr, tErr := buildTransport(m.Conf)
	if tErr != nil {
		logger.Error("mcp rebuild transport fail", "server", m.Conf.Name, "err", tErr)
		return
	}
	newSession, err := connectSession(context.Background(), transport)
	if err != nil {
		logger.Error("mcp reconnect fail", "server", m.Conf.Name, "err", err)
		if closeStderr != nil {
			closeStderr()
		}
		return
	}

	m.mu.Lock()
	oldSession := m.session
	oldCloseStderr := m.closeStderr
	m.session = newSession
	m.closeStderr = closeStderr
	m.mu.Unlock()

	if oldSession != nil {
		_ = oldSession.Close()
	}
	if oldCloseStderr != nil {
		oldCloseStderr()
	}

	if tools, err := m.GetAllTools(context.Background()); err == nil {
		m.mu.Lock()
		m.tools = tools
		m.mu.Unlock()
		logger.Info("mcp client reconnected", "server", m.Conf.Name, "tools", len(tools))
	} else {
		logger.Error("mcp list tools after reconnect fail", "server", m.Conf.Name, "err", err)
	}
}

// Close tears the connection and its keepalive loop down.
func (m *MCPClient) Close() {
	if m.stopKeepalive != nil {
		m.stopKeepalive()
	}
	m.mu.Lock()
	session := m.session
	closeStderr := m.closeStderr
	m.session = nil
	m.tools = nil
	m.mu.Unlock()
	if session != nil {
		_ = session.Close()
	}
	if closeStderr != nil {
		closeStderr()
	}
}

// RemoveMCPClient closes and unregisters one server.
func RemoveMCPClient(name string) error {
	c, err := GetMCPClient(name)
	if err != nil {
		return err
	}
	c.Close()
	mcpClients.Delete(name)
	return nil
}

// ClearAllMCPClient closes and unregisters every server.
func ClearAllMCPClient() {
	mcpClients.Range(func(key, value any) bool {
		value.(*MCPClient).Close()
		mcpClients.Delete(key)
		return true
	})
}

// getMcpType derives the config-file type: explicit Type wins, then Command
// (stdio), then Url (http).
func getMcpType(conf *param.MCPConfig) string {
	if conf.Type != "" {
		return conf.Type
	}
	if conf.Command != "" {
		return param.StdioConfigType
	}
	return param.HTTPConfigType
}

// changeEnvMapToSlice flattens an env map into KEY=VALUE slices for exec.
func changeEnvMapToSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	res := make([]string, 0, len(env))
	for k, v := range env {
		res = append(res, k+"="+v)
	}
	return res
}

// checkSSEOrHTTP probes the endpoint once: a text/event-stream response means
// the legacy SSE transport, anything else is treated as streamable HTTP.
func checkSSEOrHTTP(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if strings.EqualFold(resp.Header.Get("Content-Type"), "text/event-stream") {
		return param.SSEType, nil
	}
	return param.HTTPStreamer, nil
}
