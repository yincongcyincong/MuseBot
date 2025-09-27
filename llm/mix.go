package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go/constants"
	openrouter "github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
)

type AIRouterReq struct {
	ToolCall           []openrouter.ToolCall
	ToolMessage        []openrouter.ChatCompletionMessage
	CurrentToolMessage []openrouter.ChatCompletionMessage
	
	OpenRouterMsgs []openrouter.ChatCompletionMessage
}

var (
	pngFetch = regexp.MustCompile(`https?://[^\s)]+\.png[^\s)]*`)
)

// AI302FetchResp fetch response
type AI302FetchResp struct {
	TaskID         string           `json:"task_id"`
	UpstreamTaskID string           `json:"upstream_task_id"`
	Status         string           `json:"status"`
	VideoURL       string           `json:"video_url"`
	RawResponse    AI302RawResponse `json:"raw_response"`
	Model          string           `json:"model"`
	ExecutionTime  int              `json:"execution_time"`
	CreatedAt      *string          `json:"created_at"`
	CompletedAt    string           `json:"completed_at"`
}

type AI302RawResponse struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Prompt    string `json:"prompt"`
	State     string `json:"state"`
	Video     string `json:"video"`
}

// Create video response
type CreateResp struct {
	TaskID string `json:"task_id"`
}

func (d *AIRouterReq) GetModel(l *LLM) {
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
		return
	}
	
	switch *conf.BaseConfInfo.Type {
	case param.AI302:
		l.Model = openai.GPT3Dot5Turbo
		if userInfo != nil && userInfo.Mode != "" {
			l.Model = userInfo.Mode
		}
	case param.OpenRouter:
		l.Model = param.DeepseekDeepseekR1_0528Free
		if userInfo != nil && userInfo.Mode != "" && param.OpenRouterModels[userInfo.Mode] {
			l.Model = userInfo.Mode
		}
	}
	
	logger.Info("User info", "userID", userInfo.UserId, "mode", l.Model)
	
}

func (d *AIRouterReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	
	client := GetMixClient()
	
	request := openrouter.ChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: &openrouter.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Tools:            l.OpenRouterTools,
	}
	
	request.Messages = d.OpenRouterMsgs
	
	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	hasTools := false
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Info("Stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.Warn("Stream error", "updateMsgID", l.MsgId, "err", err)
			return err
		}
		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = d.requestToolsCall(ctx, choice, l)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
					}
				}
			}
			
			if !hasTools {
				msgInfoContent = l.SendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	if hasTools && len(d.CurrentToolMessage) != 0 {
		d.CurrentToolMessage = append([]openrouter.ChatCompletionMessage{
			{
				Role: openrouter.ChatMessageRoleAssistant,
				Content: openrouter.Content{
					Text: l.WholeContent,
				},
				ToolCalls: d.ToolCall,
			},
		}, d.CurrentToolMessage...)
		
		d.ToolMessage = append(d.ToolMessage, d.CurrentToolMessage...)
		d.OpenRouterMsgs = append(d.OpenRouterMsgs, d.CurrentToolMessage...)
		d.CurrentToolMessage = make([]openrouter.ChatCompletionMessage, 0)
		d.ToolCall = make([]openrouter.ToolCall, 0)
		return d.Send(ctx, l)
	}
	
	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (d *AIRouterReq) GetUserMessage(msg string) {
	d.GetMessage(openrouter.ChatMessageRoleUser, msg)
}

func (d *AIRouterReq) GetAssistantMessage(msg string) {
	d.GetMessage(openrouter.ChatMessageRoleAssistant, msg)
}

func (d *AIRouterReq) AppendMessages(client LLMClient) {
	if len(d.OpenRouterMsgs) == 0 {
		d.OpenRouterMsgs = make([]openrouter.ChatCompletionMessage, 0)
	}
	
	d.OpenRouterMsgs = append(d.OpenRouterMsgs, client.(*AIRouterReq).OpenRouterMsgs...)
}

func (d *AIRouterReq) GetMessage(role, msg string) {
	if len(d.OpenRouterMsgs) == 0 {
		d.OpenRouterMsgs = []openrouter.ChatCompletionMessage{
			{
				Role: role,
				Content: openrouter.Content{
					Text: msg,
					//Multi: []openrouter.ChatMessagePart{
					//	{
					//		Type: openrouter.ChatMessagePartTypeText,
					//		Text: msg,
					//	},
					//},
				},
			},
		}
		return
	}
	
	d.OpenRouterMsgs = append(d.OpenRouterMsgs, openrouter.ChatCompletionMessage{
		Role: role,
		Content: openrouter.Content{
			Text: msg,
			//Multi: []openrouter.ChatMessagePart{
			//	{
			//		Type: openrouter.ChatMessagePartTypeText,
			//		Text: msg,
			//	},
			//},
		},
	})
}

func (d *AIRouterReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	client := GetMixClient()
	
	request := openrouter.ChatCompletionRequest{
		Model:            l.Model,
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Tools:            l.OpenRouterTools,
		Messages:         d.OpenRouterMsgs,
	}
	
	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("CreateChatCompletion error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.OpenRouterMsgs[len(d.OpenRouterMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
	}
	
	return response.Choices[0].Message.Content.Text, nil
}

func (d *AIRouterReq) requestOneToolsCall(ctx context.Context, toolsCall []openrouter.ToolCall, l *LLM) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			return
		}
		
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return
		}
		
		d.OpenRouterMsgs = append(d.OpenRouterMsgs, openrouter.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: openrouter.Content{
				Text: toolsData,
			},
			ToolCallID: tool.ID,
		})
		logger.Info("exec tool", "name", tool.Function.Name, "toolsData", toolsData)
		l.DirectSendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "send_mcp_info", map[string]interface{}{
			"function_name": tool.Function.Name,
			"request_args":  property,
			"response":      toolsData,
		}))
	}
}

func (d *AIRouterReq) requestToolsCall(ctx context.Context, choice openrouter.ChatCompletionStreamChoice, l *LLM) error {
	
	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})
		
		if toolCall.Function.Name != "" {
			d.ToolCall = append(d.ToolCall, toolCall)
			d.ToolCall[len(d.ToolCall)-1].Function.Name = toolCall.Function.Name
		}
		
		if toolCall.ID != "" {
			d.ToolCall[len(d.ToolCall)-1].ID = toolCall.ID
		}
		
		if toolCall.Type != "" {
			d.ToolCall[len(d.ToolCall)-1].Type = toolCall.Type
		}
		
		if toolCall.Function.Arguments != "" && toolCall.Function.Name == "" {
			d.ToolCall[len(d.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(d.ToolCall[len(d.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		tool := d.ToolCall[len(d.ToolCall)-1]
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "function", tool.Function.Name,
				"toolCall", tool.ID, "argument", tool.Function.Arguments)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "function", tool.Function.Name,
				"toolCall", tool.ID, "argument", tool.Function.Arguments)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, openrouter.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: openrouter.Content{
				Text: toolsData,
			},
			ToolCallID: tool.ID,
		})
		
		logger.Info("send tool request", "function", tool.Function.Name,
			"toolCall", tool.ID, "argument", tool.Function.Arguments,
			"res", toolsData)
		
		l.DirectSendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "send_mcp_info", map[string]interface{}{
			"function_name": tool.Function.Name,
			"request_args":  property,
			"response":      toolsData,
		}))
	}
	
	return nil
}

func GenerateMixImg(prompt string, imageContent []byte) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	messages := openrouter.ChatCompletionMessage{
		Role: constants.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: prompt,
				},
			},
		},
	}
	
	if len(imageContent) != 0 {
		messages.Content.Multi = append(messages.Content.Multi, openrouter.ChatMessagePart{
			Type: openrouter.ChatMessagePartTypeImageURL,
			ImageURL: &openrouter.ChatMessageImageURL{
				URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
			},
		})
	}
	
	client := GetMixClient()
	request := openrouter.ChatCompletionRequest{
		Model:    *conf.PhotoConfInfo.MixImageModel,
		Messages: []openrouter.ChatCompletionMessage{messages},
	}
	
	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("create chat completion fail", "err", err)
		return "", 0, err
	}
	
	if len(response.Choices) != 0 {
		if *conf.BaseConfInfo.MediaType == param.AI302 {
			pngs := pngFetch.FindAllString(response.Choices[0].Message.Content.Text, -1)
			return pngs[len(pngs)-1], response.Usage.TotalTokens, nil
		} else if *conf.BaseConfInfo.MediaType == param.OpenRouter {
			if len(response.Choices[0].Message.Content.Multi) != 0 {
				return response.Choices[0].Message.Content.Multi[0].ImageURL.URL, response.Usage.TotalTokens, nil
			}
		}
	}
	
	return "", 0, errors.New("image is empty")
}

func GetMixClient() *openrouter.Client {
	config := openrouter.DefaultConfig(*conf.BaseConfInfo.MixToken)
	config.HTTPClient = utils.GetLLMProxyClient()
	if conf.BaseConfInfo.SpecialLLMUrl != "" {
		config.BaseURL = conf.BaseConfInfo.SpecialLLMUrl
	}
	if *conf.BaseConfInfo.CustomUrl != "" {
		config.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	return openrouter.NewClientWithConfig(*config)
}

func Generate302AIVideo(prompt string, image []byte) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	
	// Step 1: prepare payload using map -> json
	payloadMap := map[string]interface{}{
		"model":      *conf.VideoConfInfo.AI302VideoModel,
		"prompt":     prompt,
		"duration":   *conf.VideoConfInfo.Duration,
		"resolution": *conf.VideoConfInfo.Resolution,
		"fps":        *conf.VideoConfInfo.FPS,
	}
	
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal payload: %w", err)
	}
	payload := strings.NewReader(string(payloadBytes))
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.302.ai/302/v2/video/create", payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+*conf.BaseConfInfo.MixToken)
	req.Header.Add("Content-Type", "application/json")
	
	res, err := httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call create API: %w", err)
	}
	defer res.Body.Close()
	
	body, _ := io.ReadAll(res.Body)
	
	var createResp CreateResp
	if err := json.Unmarshal(body, &createResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse create response: %w, body=%s", err, string(body))
	}
	if createResp.TaskID == "" {
		return "", 0, fmt.Errorf("no task_id returned from create API, body=%s", string(body))
	}
	
	// Step 2: Poll fetch API (保持原逻辑)
	fetchURL := "https://api.302.ai/302/v2/video/fetch/" + createResp.TaskID
	for {
		select {
		case <-ctx.Done():
			return "", 0, fmt.Errorf("context canceled or timeout: %w", ctx.Err())
		default:
		}
		
		req, _ := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
		req.Header.Add("Authorization", "Bearer "+*conf.BaseConfInfo.MixToken)
		
		res, err := httpClient.Do(req)
		if err != nil {
			logger.Error("failed to fetch result:", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		
		var fetchResp AI302FetchResp
		if err := json.Unmarshal(body, &fetchResp); err != nil {
			logger.Error("failed to parse fetch response:", "err", err, "body", string(body))
			time.Sleep(5 * time.Second)
			continue
		}
		
		if fetchResp.Status == "completed" {
			if fetchResp.VideoURL != "" {
				return fetchResp.VideoURL, 0, nil
			}
			return "", 0, fmt.Errorf("task completed but no video url found, body=%s", string(body))
		} else if fetchResp.Status == "failed" {
			return "", 0, fmt.Errorf("video generation failed: body=%s", string(body))
		} else {
			logger.Info("task is still running, polling again...")
		}
		
		time.Sleep(5 * time.Second)
	}
}

func GetMixImageContent(imageContent []byte, content string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	contentPrompt := content
	if content == "" {
		contentPrompt = i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_handle_prompt", nil)
	}
	
	messages := openrouter.ChatCompletionMessage{
		Role: constants.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: contentPrompt,
				},
				{
					Type: openrouter.ChatMessagePartTypeImageURL,
					ImageURL: &openrouter.ChatMessageImageURL{
						URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
					},
				},
			},
		},
	}
	
	client := GetMixClient()
	
	request := openrouter.ChatCompletionRequest{
		Model:    *conf.PhotoConfInfo.MixRecModel,
		Messages: []openrouter.ChatCompletionMessage{messages},
	}
	
	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("create chat completion fail", "err", err)
		return "", 0, err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", 0, errors.New("response is empty")
	}
	
	return response.Choices[0].Message.Content.Text, response.Usage.TotalTokens, nil
}
