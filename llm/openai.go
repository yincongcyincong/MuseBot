package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
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

type OpenAIReq struct {
	ToolCall           []openai.ToolCall
	ToolMessage        []openai.ChatCompletionMessage
	CurrentToolMessage []openai.ChatCompletionMessage
	
	OpenAIMsgs []openai.ChatCompletionMessage
}

func (d *OpenAIReq) GetModel(l *LLM) {
	l.Model = openai.GPT3Dot5Turbo0125
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.OpenAIModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *OpenAIReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	d.GetModel(l)
	
	// set deepseek proxy
	httpClient := utils.GetLLMProxyClient()
	openaiConfig := openai.DefaultConfig(*conf.BaseConfInfo.OpenAIToken)
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
	//openaiConfig.BaseURL = "https://api.chatanywhere.org"
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)
	
	request := openai.ChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: &openai.StreamOptions{
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
		Tools:            l.OpenAITools,
	}
	
	request.Messages = d.OpenAIMsgs
	
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
				err = d.RequestToolsCall(ctx, choice, l)
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
		d.CurrentToolMessage = append([]openai.ChatCompletionMessage{
			{
				Role:      deepseek.ChatMessageRoleAssistant,
				Content:   l.WholeContent,
				ToolCalls: d.ToolCall,
			},
		}, d.CurrentToolMessage...)
		
		d.ToolMessage = append(d.ToolMessage, d.CurrentToolMessage...)
		d.OpenAIMsgs = append(d.OpenAIMsgs, d.CurrentToolMessage...)
		d.CurrentToolMessage = make([]openai.ChatCompletionMessage, 0)
		d.ToolCall = make([]openai.ToolCall, 0)
		return d.Send(ctx, l)
	}
	
	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (d *OpenAIReq) GetUserMessage(msg string) {
	d.GetMessage(openai.ChatMessageRoleUser, msg)
}

func (d *OpenAIReq) GetAssistantMessage(msg string) {
	d.GetMessage(openai.ChatMessageRoleAssistant, msg)
}

func (d *OpenAIReq) AppendMessages(client LLMClient) {
	if len(d.OpenAIMsgs) == 0 {
		d.OpenAIMsgs = make([]openai.ChatCompletionMessage, 0)
	}
	
	d.OpenAIMsgs = append(d.OpenAIMsgs, client.(*OpenAIReq).OpenAIMsgs...)
}

func (d *OpenAIReq) GetMessage(role, msg string) {
	if len(d.OpenAIMsgs) == 0 {
		d.OpenAIMsgs = []openai.ChatCompletionMessage{
			{
				Role:    role,
				Content: msg,
			},
		}
		return
	}
	
	d.OpenAIMsgs = append(d.OpenAIMsgs, openai.ChatCompletionMessage{
		Role:    role,
		Content: msg,
	})
}

func (d *OpenAIReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	// set deepseek proxy
	d.GetModel(l)
	httpClient := utils.GetLLMProxyClient()
	
	openaiConfig := openai.DefaultConfig(*conf.BaseConfInfo.OpenAIToken)
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
	//openaiConfig.BaseURL = "https://api.chatanywhere.org"
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)
	
	request := openai.ChatCompletionRequest{
		Model:            l.Model,
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Tools:            l.OpenAITools,
	}
	
	request.Messages = d.OpenAIMsgs
	
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.OpenAIMsgs[len(d.OpenAIMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (d *OpenAIReq) requestOneToolsCall(ctx context.Context, toolsCall []openai.ToolCall, l *LLM) {
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
		
		d.OpenAIMsgs = append(d.OpenAIMsgs, openai.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
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

func (d *OpenAIReq) RequestToolsCall(ctx context.Context, choice openai.ChatCompletionStreamChoice, l *LLM) error {
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
		d.CurrentToolMessage = append(d.CurrentToolMessage, openai.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
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

// GenerateOpenAIImg generate image
func GenerateOpenAIImg(prompt string, imageContent []byte) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	openaiConfig := openai.DefaultConfig(*conf.BaseConfInfo.OpenAIToken)
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
	//openaiConfig.BaseURL = "https://api.chatanywhere.org"
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)
	
	var respUrl openai.ImageResponse
	var err error
	if len(imageContent) != 0 {
		imageFile, err := utils.ConvertToPNGFile(imageContent)
		if err != nil {
			logger.Error("failed to create temp file:", err)
			return nil, 0, err
		}
		defer os.Remove(imageFile.Name())
		defer imageFile.Close()
		
		respUrl, err = client.CreateEditImage(ctx, openai.ImageEditRequest{
			Image:  imageFile,
			Prompt: prompt,
			Model:  *conf.PhotoConfInfo.OpenAIImageModel,
			N:      1,
			Size:   *conf.PhotoConfInfo.OpenAIImageSize,
		})
	} else {
		respUrl, err = client.CreateImage(
			ctx,
			openai.ImageRequest{
				Prompt: prompt,
				Model:  *conf.PhotoConfInfo.OpenAIImageModel,
				Size:   *conf.PhotoConfInfo.OpenAIImageSize,
				N:      1,
				Style:  *conf.PhotoConfInfo.OpenAIImageStyle,
			},
		)
	}
	
	if err != nil {
		logger.Error("CreateImage error", "err", err)
		return nil, 0, err
	}
	
	if len(respUrl.Data) == 0 {
		logger.Error("response is emtpy", "response", respUrl)
		return nil, 0, errors.New("response is empty")
	}
	
	imageContentByte, err := base64.StdEncoding.DecodeString(respUrl.Data[0].B64JSON)
	if err != nil {
		logger.Error("decode image error", "err", err)
		return nil, 0, err
	}
	
	return imageContentByte, respUrl.Usage.TotalTokens, nil
}

func GenerateOpenAIText(audioContent []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	openaiConfig := openai.DefaultConfig(*conf.BaseConfInfo.OpenAIToken)
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
	//openaiConfig.BaseURL = "https://api.chatanywhere.org"
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)
	
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "voice." + utils.DetectAudioFormat(audioContent),
		Reader:   bytes.NewReader(audioContent),
		Format:   "json",
	}
	
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		logger.Error("CreateTranscription error", "err", err)
		return "", err
	}
	
	return resp.Text, nil
}

func GetOpenAIImageContent(imageContent []byte, content string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	openaiConfig := openai.DefaultConfig(*conf.BaseConfInfo.OpenAIToken)
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
	//openaiConfig.BaseURL = "https://api.chatanywhere.org"
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)
	
	contentPrompt := content
	if content == "" {
		contentPrompt = i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_handle_prompt", nil)
	}
	
	imageDataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageContent)
	req := openai.ChatCompletionRequest{
		Model: *conf.PhotoConfInfo.OpenAIRecModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    imageDataURL,
							Detail: openai.ImageURLDetailHigh,
						},
					},
					{
						Type: openai.ChatMessagePartTypeText,
						Text: contentPrompt,
					},
				},
			},
		},
		MaxTokens: 1000,
	}
	
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.Error("CreateChatCompletion error", "err", err)
		return "", 0, err
	}
	
	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}
