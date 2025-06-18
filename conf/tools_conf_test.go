package conf

import (
	"flag"
	"os"
	"testing"
)

func TestInitConf_InitTools(t *testing.T) {
	if UseTools == nil {
		UseTools = getPointBool(true)
		os.Setenv("MCP_CONF_PATH", "./conf/mcp/mcp.json")

		InitToolsConf()
		flag.Parse()
	}

	InitTools()
	if len(DeepseekTools) != len(OpenAITools) {
		t.Errorf("%s expected %d, got %d", "tools number", len(DeepseekTools), len(OpenAITools))
	}
}

func getPointBool(b bool) *bool {
	return &b
}
