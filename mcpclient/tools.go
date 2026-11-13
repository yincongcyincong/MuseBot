package clients

import (
	"encoding/json"
	"strings"

	"github.com/cohesion-org/deepseek-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"google.golang.org/genai"
)

// ResultString renders a tool result as the string fed back to the LLM: text
// contents are concatenated, other content types are JSON-encoded.
func ResultString(result *mcp.CallToolResult) string {
	if result == nil {
		return ""
	}

	var res strings.Builder
	for _, content := range result.Content {
		switch c := content.(type) {
		case *mcp.TextContent:
			res.WriteString(c.Text)
		default:
			jsonBytes, _ := json.MarshalIndent(content, "", "  ")
			res.Write(jsonBytes)
		}
	}
	return res.String()
}

// inputSchemaParts extracts the properties and required parameter names from
// a tool input schema. On the client side the SDK decodes the server's JSON
// Schema into a generic map, so this walks that map.
func inputSchemaParts(schema any) (map[string]any, []string) {
	root, ok := schema.(map[string]any)
	if !ok {
		return nil, nil
	}

	var props map[string]any
	if p, ok := root["properties"].(map[string]any); ok {
		props = p
	}

	var required []string
	if list, ok := root["required"].([]any); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				required = append(required, s)
			}
		}
	}
	return props, required
}

// TransToolsToDPFunctionCall converts MCP tools to deepseek function calls.
func TransToolsToDPFunctionCall(tools []*mcp.Tool) []deepseek.Tool {
	dpTools := make([]deepseek.Tool, 0, len(tools))
	for _, tool := range tools {
		props, required := inputSchemaParts(tool.InputSchema)
		dpTools = append(dpTools, deepseek.Tool{
			Type: "function",
			Function: deepseek.Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: &deepseek.FunctionParameters{
					Type:       "object",
					Properties: props,
					Required:   required,
				},
			},
		})
	}
	return dpTools
}

// TransToolsToChatGPTFunctionCall converts MCP tools to OpenAI function calls.
func TransToolsToChatGPTFunctionCall(tools []*mcp.Tool) []openai.Tool {
	oaTools := make([]openai.Tool, 0, len(tools))
	for _, tool := range tools {
		props, required := inputSchemaParts(tool.InputSchema)
		oaTool := openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: &deepseek.FunctionParameters{
					Type:       "object",
					Properties: props,
					Required:   required,
				},
			},
		}
		// providers reject an empty parameter object harder than a missing one
		if len(props) == 0 {
			oaTool.Function.Parameters = nil
		}
		oaTools = append(oaTools, oaTool)
	}
	return oaTools
}

// TransToolsToGeminiFunctionCall converts MCP tools to Gemini tool declarations.
func TransToolsToGeminiFunctionCall(tools []*mcp.Tool) []*genai.Tool {
	geminiTools := make([]*genai.Tool, 0, len(tools))
	for _, tool := range tools {
		props, required := inputSchemaParts(tool.InputSchema)
		// round-trip the generic map so genai's schema types get decoded
		prop := make(map[string]*genai.Schema)
		if propByte, err := json.Marshal(props); err == nil {
			if err = json.Unmarshal(propByte, &prop); err != nil {
				continue
			}
		}
		geminiTools = append(geminiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters: &genai.Schema{
						Type:       genai.TypeObject,
						Properties: prop,
						Required:   required,
					},
				},
			},
		})
	}
	return geminiTools
}

// TransToolsToVolFunctionCall converts MCP tools to Volcengine function calls.
func TransToolsToVolFunctionCall(tools []*mcp.Tool) []*model.Tool {
	volTools := make([]*model.Tool, 0, len(tools))
	for _, tool := range tools {
		props, required := inputSchemaParts(tool.InputSchema)
		volTools = append(volTools, &model.Tool{
			Type: "function",
			Function: &model.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": props,
					"required":   required,
				},
			},
		})
	}
	return volTools
}
