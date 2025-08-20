package llm

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"google.golang.org/genai"
)

type GeminiReq struct {
	ToolCall           []*genai.FunctionCall
	ToolMessage        []*genai.Content
	CurrentToolMessage []*genai.Content
	
	GeminiMsgs []*genai.Content
}

func (h *GeminiReq) GetMessages(userId string, prompt string) {
	messages := make([]*genai.Content, 0)
	
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > 10 {
			aqs = aqs[len(aqs)-10:]
		}
		for i, record := range aqs {
			if record.Answer != "" && record.Question != "" {
				logger.Info("context content", "dialog", i, "question:", record.Question,
					"toolContent", record.Content, "answer:", record.Answer)
				
				messages = append(messages, &genai.Content{
					Role: genai.RoleUser,
					Parts: []*genai.Part{
						{
							Text: record.Question,
						},
					},
				})
				
				messages = append(messages, &genai.Content{
					Role: genai.RoleModel,
					Parts: []*genai.Part{
						{
							Text: record.Answer,
						},
					},
				})
				
			}
		}
	}
	
	h.GeminiMsgs = messages
}

func (h *GeminiReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	h.GetModel(l)
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("init gemini client fail", "err", err)
		return err
	}
	
	config := &genai.GenerateContentConfig{
		TopP:             genai.Ptr[float32](float32(*conf.LLMConfInfo.TopP)),
		FrequencyPenalty: genai.Ptr[float32](float32(*conf.LLMConfInfo.FrequencyPenalty)),
		PresencePenalty:  genai.Ptr[float32](float32(*conf.LLMConfInfo.PresencePenalty)),
		Temperature:      genai.Ptr[float32](float32(*conf.LLMConfInfo.Temperature)),
		Tools:            l.GeminiTools,
	}
	
	chat, err := client.Chats.Create(ctx, l.Model, config, h.GeminiMsgs)
	if err != nil {
		logger.Error("create chat fail", "err", err)
		return err
	}
	
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	hasTools := false
	for response, err := range chat.SendMessageStream(ctx, *genai.NewPartFromText(l.Content)) {
		if errors.Is(err, io.EOF) {
			logger.Info("stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.Error("stream error:", "updateMsgID", l.MsgId, "err", err)
			break
		}
		
		toolCalls := response.FunctionCalls()
		if len(toolCalls) > 0 {
			hasTools = true
			err = h.RequestToolsCall(ctx, response)
			if err != nil {
				if errors.Is(err, ToolsJsonErr) {
					continue
				} else {
					logger.Error("requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
				}
			}
		}
		
		if len(response.Text()) > 0 {
			msgInfoContent = l.SendMsg(msgInfoContent, response.Text())
		}
		
		if response.UsageMetadata != nil {
			l.Token += int(response.UsageMetadata.TotalTokenCount)
			metrics.TotalTokens.Add(float64(l.Token))
		}
		
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	logger.Info("Stream finished", "updateMsgID", l.MsgId)
	if !hasTools || len(h.CurrentToolMessage) == 0 {
		db.InsertMsgRecord(l.UserId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Token:    l.Token,
			Mode:     param.Gemini,
		}, true)
	} else {
		h.ToolMessage = append(h.ToolMessage, h.CurrentToolMessage...)
		h.GeminiMsgs = append(h.GeminiMsgs, h.CurrentToolMessage...)
		h.CurrentToolMessage = make([]*genai.Content, 0)
		h.ToolCall = make([]*genai.FunctionCall, 0)
		return h.Send(ctx, l)
	}
	
	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (h *GeminiReq) GetUserMessage(msg string) {
	h.GetMessage(genai.RoleUser, msg)
}

func (h *GeminiReq) GetAssistantMessage(msg string) {
	h.GetMessage(genai.RoleModel, msg)
}

func (h *GeminiReq) AppendMessages(client LLMClient) {
	if len(h.GeminiMsgs) == 0 {
		h.GeminiMsgs = make([]*genai.Content, 0)
	}
	
	h.GeminiMsgs = append(h.GeminiMsgs, client.(*GeminiReq).GeminiMsgs...)
}

func (h *GeminiReq) GetMessage(role, msg string) {
	if len(h.GeminiMsgs) == 0 {
		h.GeminiMsgs = []*genai.Content{
			{
				Role: role,
				Parts: []*genai.Part{
					{
						Text: msg,
					},
				},
			},
		}
		return
	}
	
	h.GeminiMsgs = append(h.GeminiMsgs, &genai.Content{
		Role: role,
		Parts: []*genai.Part{
			{
				Text: msg,
			},
		},
	})
}

func (h *GeminiReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	h.GetModel(l)
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("init gemini client fail", "err", err)
		return "", err
	}
	
	config := &genai.GenerateContentConfig{
		TopP:             genai.Ptr[float32](float32(*conf.LLMConfInfo.TopP)),
		FrequencyPenalty: genai.Ptr[float32](float32(*conf.LLMConfInfo.FrequencyPenalty)),
		PresencePenalty:  genai.Ptr[float32](float32(*conf.LLMConfInfo.PresencePenalty)),
		Temperature:      genai.Ptr[float32](float32(*conf.LLMConfInfo.Temperature)),
		Tools:            l.GeminiTools,
	}
	
	chat, err := client.Chats.Create(ctx, l.Model, config, h.GeminiMsgs)
	if err != nil {
		logger.Error("create chat fail", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	response, err := chat.Send(ctx, genai.NewPartFromText(l.Content))
	if err != nil {
		logger.Error("create chat fail", "err", err)
		return "", err
	}
	
	l.Token += int(response.UsageMetadata.TotalTokenCount)
	if len(response.FunctionCalls()) > 0 {
		h.requestOneToolsCall(ctx, response.FunctionCalls())
	}
	
	return response.Text(), nil
}

func (h *GeminiReq) requestOneToolsCall(ctx context.Context, toolsCall []*genai.FunctionCall) {
	for _, tool := range toolsCall {
		
		mc, err := clients.GetMCPClientByToolName(tool.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "name", tool.Name, "args", tool.Args)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Name, tool.Args)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "name", tool.Name, "args", tool.Args)
			return
		}
		
		h.GeminiMsgs = append(h.GeminiMsgs, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionCall: tool,
				},
			},
		})
		
		h.GeminiMsgs = append(h.GeminiMsgs, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						Response: map[string]any{"output": toolsData},
						ID:       tool.ID,
						Name:     tool.Name,
					},
				},
			},
		})
		
		logger.Info("exec tool", "name", tool.Name, "args", tool.Args, "toolsData", toolsData)
	}
}

func (h *GeminiReq) RequestToolsCall(ctx context.Context, response *genai.GenerateContentResponse) error {
	
	for _, toolCall := range response.FunctionCalls() {
		
		if toolCall.Name != "" {
			h.ToolCall = append(h.ToolCall, toolCall)
			h.ToolCall[len(h.ToolCall)-1].Name = toolCall.Name
		}
		
		if toolCall.ID != "" {
			h.ToolCall[len(h.ToolCall)-1].ID = toolCall.ID
		}
		
		if toolCall.Args != nil {
			h.ToolCall[len(h.ToolCall)-1].Args = toolCall.Args
		}
		
		mc, err := clients.GetMCPClientByToolName(h.ToolCall[len(h.ToolCall)-1].Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, h.ToolCall[len(h.ToolCall)-1].Name, h.ToolCall[len(h.ToolCall)-1].Args)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		h.CurrentToolMessage = append(h.CurrentToolMessage, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionCall: toolCall,
				},
			},
		})
		
		h.CurrentToolMessage = append(h.CurrentToolMessage, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						Response: map[string]any{"output": toolsData},
						ID:       h.ToolCall[len(h.ToolCall)-1].ID,
						Name:     h.ToolCall[len(h.ToolCall)-1].Name,
					},
				},
			},
		})
		logger.Info("send tool request", "function", toolCall.Name,
			"toolCall", toolCall.ID, "argument", toolCall.Args, "toolsData", toolsData)
	}
	
	return nil
}

func (h *GeminiReq) GetModel(l *LLM) {
	l.Model = param.ModelGemini20Flash
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.GeminiModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func GenerateGeminiImg(prompt string, imageContent []byte) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("create client fail", "err", err)
		return nil, 0, err
	}
	
	geminiContent := genai.Text(prompt)
	if len(imageContent) > 0 {
		geminiContent = append(geminiContent, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{
					InlineData: &genai.Blob{
						Data:     imageContent,
						MIMEType: "image/" + utils.DetectImageFormat(imageContent),
					},
				},
			},
		})
	}
	
	response, err := client.Models.GenerateContent(
		ctx,
		*conf.PhotoConfInfo.GeminiImageModel,
		geminiContent,
		&genai.GenerateContentConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	)
	if err != nil {
		logger.Error("generate image fail", "err", err)
		return nil, 0, err
	}
	
	if len(response.Candidates) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				return part.InlineData.Data, int(response.UsageMetadata.TotalTokenCount), nil
			}
		}
	}
	
	return nil, 0, errors.New("image is empty")
}

func GenerateGeminiVideo(prompt string, image []byte) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("create client fail", "err", err)
		return nil, 0, err
	}
	
	var geminiImage *genai.Image
	if len(image) > 0 {
		geminiImage = &genai.Image{
			ImageBytes: image,
			MIMEType:   "image/" + utils.DetectImageFormat(image),
		}
	}
	
	operation, err := client.Models.GenerateVideos(ctx,
		*conf.VideoConfInfo.GeminiVideoModel, prompt,
		geminiImage,
		&genai.GenerateVideosConfig{
			AspectRatio:      *conf.VideoConfInfo.GeminiVideoAspectRatio,
			PersonGeneration: *conf.VideoConfInfo.GeminiVideoPersonGeneration,
			DurationSeconds:  &conf.VideoConfInfo.GeminiVideoDurationSeconds,
		})
	if err != nil {
		logger.Error("generate video fail", "err", err)
		return nil, 0, err
	}
	
	for !operation.Done {
		logger.Info("video is createing...")
		time.Sleep(5 * time.Second)
		operation, err = client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			logger.Error("get video operation fail", "err", err)
			return nil, 0, err
		}
	}
	
	if len(operation.Response.GeneratedVideos) == 0 {
		logger.Error("generate video fail", "err", "video is empty", "resp", operation.Response)
		return nil, 0, errors.New("video is empty")
	}
	
	var totalToken int
	if operation.Metadata != nil {
		if usageRaw, ok := operation.Metadata["usageMetadata"]; ok {
			if usage, ok := usageRaw.(map[string]interface{}); ok {
				if tokenValue, ok := usage["totalTokenCount"]; ok {
					if tokenFloat, ok := tokenValue.(float64); ok {
						totalToken = int(tokenFloat)
					}
				}
			}
		}
	}
	
	return operation.Response.GeneratedVideos[0].Video.VideoBytes, totalToken, nil
}

func GenerateGeminiText(audioContent []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("create client fail", "err", err)
		return "", err
	}
	
	parts := []*genai.Part{
		genai.NewPartFromText("Get Content from this audio clip"),
		{
			InlineData: &genai.Blob{
				MIMEType: "audio/" + utils.DetectAudioFormat(audioContent),
				Data:     audioContent,
			},
		},
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	
	result, err := client.Models.GenerateContent(
		ctx,
		*conf.PhotoConfInfo.GeminiRecModel,
		contents,
		nil,
	)
	
	if err != nil || result == nil {
		logger.Error("generate text fail", "err", err)
		return "", err
	}
	
	return result.Text(), nil
}

func GetGeminiImageContent(imageContent []byte, content string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.BaseConfInfo.GeminiToken,
	})
	if err != nil {
		logger.Error("create client fail", "err", err)
		return "", err
	}
	
	contentPrompt := content
	if content == "" {
		contentPrompt = i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_handle_prompt", nil)
	}
	
	parts := []*genai.Part{
		genai.NewPartFromBytes(imageContent, "image/jpeg"),
		genai.NewPartFromText(contentPrompt),
	}
	
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	
	result, err := client.Models.GenerateContent(
		ctx,
		*conf.PhotoConfInfo.GeminiRecModel,
		contents,
		nil,
	)
	
	if err != nil || result == nil {
		logger.Error("generate text fail", "err", err)
		return "", err
	}
	
	return result.Text(), nil
	
}
