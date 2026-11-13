package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/MuseBot/mcpclient/param"
)

// echoTestTool is the tool exposed by every test server below.
var echoTestTool = &mcp.Tool{
	Name:        "echo",
	Description: "echo the input",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"msg": map[string]any{"type": "string"},
		},
		"required": []any{"msg"},
	},
}

func echoHandler(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "echo: " + args.Msg}},
	}, nil
}

// startInMemoryServer spins up an echo server over an in-memory transport and
// returns the client side.
func startInMemoryServer(t *testing.T) mcp.Transport {
	t.Helper()
	ct, st := mcp.NewInMemoryTransports()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0"}, nil)
	server.AddTool(echoTestTool, echoHandler)
	go func() {
		_ = server.Run(context.Background(), st)
	}()
	// the server's Run returns once the client session closes; no explicit
	// server shutdown API exists
	return ct
}

func registerInMemoryClient(t *testing.T, name string) *MCPClient {
	t.Helper()
	session, err := connectSession(context.Background(), startInMemoryServer(t))
	require.NoError(t, err)
	mc := &MCPClient{
		Conf:    &param.MCPClientConf{Name: name, ClientType: "in-memory"},
		session: session,
	}
	tools, err := mc.GetAllTools(context.Background())
	require.NoError(t, err)
	mc.tools = tools

	mcpClients.Store(name, mc)
	t.Cleanup(func() {
		mcpClients.Delete(name)
		mc.Close()
	})
	return mc
}

// TestPingKeepaliveKeepsSession covers the >= 2026-07-28 protocol split: the
// enveloped ping a strict (Python) server demands is answered with "method
// not found" by sessionless Go SDK servers, which removed ping. The keepalive
// must treat that response as liveness instead of tearing the session down.
//
// The helper subprocess negotiates the same new protocol and is a stdio
// server, so a regression (missing liveness classification) would actually
// rebuild the session — assert.Same detects that.
func TestPingKeepaliveKeepsSession(t *testing.T) {
	if os.Getenv("GO_MCP_HELPER") == "1" {
		t.Skip("helper process")
	}

	conf := &param.MCPClientConf{
		Name:       "stdio-ping",
		ClientType: param.StdioType,
		StdioConf: &param.StdioClientConfig{
			Command: os.Args[0],
			Args:    []string{"-test.run=^TestHelperMCPServer$", "-test.count=1"},
			Env:     []string{"GO_MCP_HELPER=1"},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	require.NoError(t, connectMCPClient(ctx, conf))
	t.Cleanup(func() {
		require.NoError(t, RemoveMCPClient("stdio-ping"))
	})

	mc, err := GetMCPClient("stdio-ping")
	require.NoError(t, err)

	// direct evidence for the protocol split on this pair: bare ping passes,
	// the enveloped ping is the removed method
	require.NoError(t, mc.session.Ping(context.Background(), nil))
	envelopedErr := mc.session.Ping(context.Background(), pingParams(mc.session))
	require.Error(t, envelopedErr)
	require.True(t, isMethodNotFound(envelopedErr))

	before := mc.session
	mc.restartMCPServer()

	mc.mu.RLock()
	after := mc.session
	mc.mu.RUnlock()
	assert.Same(t, before, after, "a method-not-found ping response must not trigger a reconnect")
}

func TestGetByToolNameAndExecTools(t *testing.T) {
	registerInMemoryClient(t, "mem-echo")

	mc, err := GetMCPClientByToolName("echo")
	require.NoError(t, err)
	assert.Equal(t, "mem-echo", mc.Conf.Name)

	out, err := mc.ExecTools(context.Background(), "echo", map[string]interface{}{"msg": "hi"})
	require.NoError(t, err)
	assert.Equal(t, "echo: hi", out)

	// required parameter enforcement happens before any request is sent
	_, err = mc.ExecTools(context.Background(), "echo", map[string]interface{}{})
	assert.EqualError(t, err, "required parameter msg not found")

	_, err = mc.ExecTools(context.Background(), "nope", nil)
	assert.EqualError(t, err, "tool nope not found")
}

func TestResolveToolName(t *testing.T) {
	registerInMemoryClient(t, "mem-resolve")

	// exact advertised names are returned untouched
	assert.Equal(t, "echo", ResolveToolName("echo"))
	// GLM's mcp__server__tool convention maps back onto the advertised name
	assert.Equal(t, "echo", ResolveToolName("mcp__time__echo"))
	assert.Equal(t, "echo", ResolveToolName("mcp__mem-resolve__echo"))
	// unknown names are returned as-is so the lookup error names what the
	// model actually sent
	assert.Equal(t, "mcp__time__nope", ResolveToolName("mcp__time__nope"))
}

func TestRemoveAndClearRegistry(t *testing.T) {
	registerInMemoryClient(t, "mem-remove")

	require.NoError(t, RemoveMCPClient("mem-remove"))
	_, err := GetMCPClient("mem-remove")
	assert.EqualError(t, err, "client mem-remove not found")

	// removing a missing client is an error, clearing twice is fine
	assert.Error(t, RemoveMCPClient("mem-remove"))
	ClearAllMCPClient()
}

func TestInitByConfFile(t *testing.T) {
	// a live endpoint for the http transport probe: anything not answering
	// text/event-stream resolves to streamable HTTP
	probe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(probe.Close)

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	content := fmt.Sprintf(`{
	  "mcpServers": {
	    "fs": {"command": "npx", "args": ["-y", "server-filesystem", "/tmp"],
	           "env": {"A": "1"}, "description": "filesystem"},
	    "remote": {"url": "%s", "headers": {"X-Token": "t"}},
	    "off": {"command": "whatever", "disabled": true}
	  }
	}`, probe.URL)
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	confs, err := InitByConfFile(path)
	require.NoError(t, err)
	require.Len(t, confs, 2) // the disabled server is skipped

	byName := map[string]*param.MCPClientConf{}
	for _, c := range confs {
		byName[c.Name] = c
	}

	fs := byName["fs"]
	require.NotNil(t, fs)
	assert.Equal(t, param.StdioType, fs.ClientType)
	assert.Equal(t, "npx", fs.StdioConf.Command)
	assert.Equal(t, []string{"A=1"}, fs.StdioConf.Env)
	assert.Equal(t, "filesystem", fs.Description)

	remote := byName["remote"]
	require.NotNil(t, remote)
	assert.Equal(t, param.HTTPStreamer, remote.ClientType)
	assert.Equal(t, probe.URL, remote.URL)
	assert.Equal(t, map[string]string{"X-Token": "t"}, remote.Headers)

	_, err = InitByConfFile(filepath.Join(dir, "missing.json"))
	assert.Error(t, err)
}

func TestTransTools(t *testing.T) {
	tools := []*mcp.Tool{echoTestTool}

	oa := TransToolsToChatGPTFunctionCall(tools)
	require.Len(t, oa, 1)
	assert.Equal(t, "echo", oa[0].Function.Name)
	oaParams, ok := oa[0].Function.Parameters.(*deepseek.FunctionParameters)
	require.True(t, ok)
	assert.Equal(t, []string{"msg"}, oaParams.Required)

	dp := TransToolsToDPFunctionCall(tools)
	require.Len(t, dp, 1)
	assert.Equal(t, "echo", dp[0].Function.Name)
	assert.Equal(t, []string{"msg"}, dp[0].Function.Parameters.Required)

	vol := TransToolsToVolFunctionCall(tools)
	require.Len(t, vol, 1)
	assert.Equal(t, "echo", vol[0].Function.Name)
	volParams, ok := vol[0].Function.Parameters.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, []string{"msg"}, volParams["required"])

	gm := TransToolsToGeminiFunctionCall(tools)
	require.Len(t, gm, 1)
	assert.Equal(t, "echo", gm[0].FunctionDeclarations[0].Name)
	assert.Equal(t, []string{"msg"}, gm[0].FunctionDeclarations[0].Parameters.Required)
}

// TestStdioRegisterAndExec runs the full stdio path: spawn the test binary as
// an MCP server subprocess, handshake, list tools, call a tool, and tear the
// client down.
func TestStdioRegisterAndExec(t *testing.T) {
	if os.Getenv("GO_MCP_HELPER") == "1" {
		t.Skip("helper process")
	}

	conf := &param.MCPClientConf{
		Name:       "stdio-echo",
		ClientType: param.StdioType,
		StdioConf: &param.StdioClientConfig{
			Command: os.Args[0],
			Args:    []string{"-test.run=^TestHelperMCPServer$", "-test.count=1"},
			Env:     []string{"GO_MCP_HELPER=1"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	require.NoError(t, connectMCPClient(ctx, conf))
	t.Cleanup(func() {
		require.NoError(t, RemoveMCPClient("stdio-echo"))
	})

	mc, err := GetMCPClient("stdio-echo")
	require.NoError(t, err)
	assert.Equal(t, param.StdioType, mc.Conf.ClientType)

	out, err := mc.ExecTools(ctx, "echo", map[string]interface{}{"msg": "stdio"})
	require.NoError(t, err)
	assert.Equal(t, "echo: stdio", out)
}

// TestHelperMCPServer is the helper subprocess: a real MCP server speaking
// stdio over the test binary's stdin/stdout. It only runs when re-executed
// with GO_MCP_HELPER=1 (see TestStdioRegisterAndExec).
func TestHelperMCPServer(t *testing.T) {
	if os.Getenv("GO_MCP_HELPER") != "1" {
		t.Skip("helper process only")
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "helper-server", Version: "1.0"}, nil)
	server.AddTool(echoTestTool, echoHandler)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, "helper server exited:", err)
	}
	// exit without letting the test framework print "PASS" to stdout: stdout
	// is the MCP protocol channel
	os.Exit(0)
}
