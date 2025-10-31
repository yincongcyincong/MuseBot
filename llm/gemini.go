package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
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
	"google.golang.org/genai"
)

type GeminiReq struct {
	ToolCall           []*genai.FunctionCall
	ToolMessage        []*genai.Content
	CurrentToolMessage []*genai.Content
	
	GeminiMsgs []*genai.Content
}

func (h *GeminiReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "create client fail", "err", err)
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
		logger.ErrorCtx(l.Ctx, "create chat fail", "err", err)
		return err
	}
	
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	hasTools := false
	for response, err := range chat.SendMessageStream(ctx, *genai.NewPartFromText(l.GetContent(l.Content))) {
		if errors.Is(err, io.EOF) {
			logger.InfoCtx(l.Ctx, "stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.ErrorCtx(l.Ctx, "stream error:", "updateMsgID", l.MsgId, "err", err)
			return err
		}
		
		toolCalls := response.FunctionCalls()
		if len(toolCalls) > 0 {
			hasTools = true
			err = h.RequestToolsCall(ctx, response, l)
			if err != nil {
				if errors.Is(err, ToolsJsonErr) {
					continue
				} else {
					logger.ErrorCtx(l.Ctx, "requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
				}
			}
		}
		
		if !hasTools {
			msgInfoContent = l.SendMsg(msgInfoContent, response.Text())
		}
		
		if response.UsageMetadata != nil {
			l.Token += int(response.UsageMetadata.TotalTokenCount)
		}
		
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	logger.InfoCtx(l.Ctx, "Stream finished", "updateMsgID", l.MsgId)
	if hasTools && len(h.CurrentToolMessage) != 0 {
		h.ToolMessage = append(h.ToolMessage, h.CurrentToolMessage...)
		h.GeminiMsgs = append(h.GeminiMsgs, h.CurrentToolMessage...)
		h.CurrentToolMessage = make([]*genai.Content, 0)
		h.ToolCall = make([]*genai.FunctionCall, 0)
		return h.Send(ctx, l)
	}
	
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
	start := time.Now()
	
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "create client fail", "err", err)
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
		logger.ErrorCtx(l.Ctx, "create chat fail", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	response, err := chat.Send(ctx, genai.NewPartFromText(l.Content))
	if err != nil {
		logger.ErrorCtx(l.Ctx, "create chat fail", "err", err)
		return "", err
	}
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	l.Token += int(response.UsageMetadata.TotalTokenCount)
	if len(response.FunctionCalls()) > 0 {
		h.requestOneToolsCall(ctx, response.FunctionCalls(), l)
	}
	
	return response.Text(), nil
}

func (h *GeminiReq) requestOneToolsCall(ctx context.Context, toolsCall []*genai.FunctionCall, l *LLM) {
	for _, tool := range toolsCall {
		
		toolsData, err := l.ExecMcpReq(ctx, tool.Name, tool.Args)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err, "name", tool.Name, "args", tool.Args)
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
	}
}

func (h *GeminiReq) RequestToolsCall(ctx context.Context, response *genai.GenerateContentResponse, l *LLM) error {
	
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
		
		toolsData, err := l.ExecMcpReq(ctx, toolCall.Name, toolCall.Args)
		if err != nil {
			logger.ErrorCtx(ctx, "Error executing MCP request", "toolId", toolCall.ID, "err", err)
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
	}
	
	return nil
}

func (h *GeminiReq) GetModel(l *LLM) {
	l.Model = param.ModelGemini20Flash
	userInfo := db.GetCtxUserInfo(l.Ctx)
	if userInfo != nil && userInfo.LLMConfigRaw != nil && param.GeminiModels[userInfo.LLMConfigRaw.TxtModel] {
		logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", userInfo.LLMConfigRaw.TxtModel)
		l.Model = userInfo.LLMConfigRaw.TxtModel
	}
}

func GenerateGeminiImg(ctx context.Context, prompt string, imageContent []byte) ([]byte, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, err
	}
	
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(*conf.PhotoConfInfo.GeminiImageModel).Inc()
	
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
	
	metrics.APIRequestDuration.WithLabelValues(*conf.PhotoConfInfo.GeminiImageModel).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(ctx, "generate image fail", "err", err)
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

func GenerateGeminiVideo(ctx context.Context, prompt string, image []byte) ([]byte, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingVideoModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.VideoModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	var geminiImage *genai.Image
	if len(image) > 0 {
		geminiImage = &genai.Image{
			ImageBytes: image,
			MIMEType:   "image/" + utils.DetectImageFormat(image),
		}
	}
	
	duration := int32(*conf.VideoConfInfo.Duration)
	operation, err := client.Models.GenerateVideos(ctx,
		model, prompt,
		geminiImage,
		&genai.GenerateVideosConfig{
			AspectRatio:     *conf.VideoConfInfo.Radio,
			DurationSeconds: &duration,
		})
	if err != nil {
		logger.ErrorCtx(ctx, "generate video fail", "err", err)
		return nil, 0, err
	}
	
	for !operation.Done {
		logger.InfoCtx(ctx, "video is createing...")
		time.Sleep(5 * time.Second)
		operation, err = client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			logger.ErrorCtx(ctx, "get video operation fail", "err", err)
			return nil, 0, err
		}
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if len(operation.Response.GeneratedVideos) == 0 {
		logger.ErrorCtx(ctx, "generate video fail", "err", "video is empty", "resp", operation.Response)
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

func GenerateGeminiText(ctx context.Context, audioContent []byte) (string, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return "", 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingRecModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.RecModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
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
		model,
		contents,
		nil,
	)
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if err != nil || result == nil {
		logger.ErrorCtx(ctx, "generate text fail", "err", err)
		return "", 0, err
	}
	
	return result.Text(), int(result.UsageMetadata.TotalTokenCount), nil
}

func GetGeminiImageContent(ctx context.Context, imageContent []byte, content string) (string, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return "", 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingRecModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.RecModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
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
		model,
		contents,
		nil,
	)
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if err != nil || result == nil {
		logger.ErrorCtx(ctx, "generate text fail", "err", err)
		return "", 0, err
	}
	
	return result.Text(), int(result.UsageMetadata.TotalTokenCount), nil
	
}

func GeminiTTS(ctx context.Context, content, encoding string) ([]byte, int, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingTTSModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.TTSModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	parts := []*genai.Part{
		genai.NewPartFromText(i18n.GetMessage(*conf.BaseConfInfo.Lang, "audio_create_prompt", map[string]interface{}{
			"content": content,
		})),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	response, err := client.Models.GenerateContent(
		ctx,
		model,
		contents,
		&genai.GenerateContentConfig{
			ResponseModalities: []string{
				"AUDIO",
			},
			SpeechConfig: &genai.SpeechConfig{
				VoiceConfig: &genai.VoiceConfig{
					PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
						VoiceName: *conf.AudioConfInfo.GeminiVoiceName,
					},
				},
			},
		},
	)
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if err != nil {
		logger.ErrorCtx(ctx, "generate audio fail", "err", err)
		return nil, 0, 0, err
	}
	
	if len(response.Candidates) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				var data = part.InlineData.Data
				data, err = utils.GetAudioData(encoding, part.InlineData.Data)
				if err != nil {
					logger.ErrorCtx(ctx, "convert audio fail", "err", err)
				}
				return data, int(response.UsageMetadata.TotalTokenCount), utils.PCMDuration(len(part.InlineData.Data), 24000, 1, 16), nil
			}
		}
	}
	
	return nil, 0, 0, errors.New("audio is empty")
}

func GetGeminiClient(ctx context.Context) (*genai.Client, error) {
	httpClient := utils.GetLLMProxyClient()
	httpOption := genai.HTTPOptions{}
	if *conf.BaseConfInfo.CustomUrl != "" {
		httpOption.BaseURL = *conf.BaseConfInfo.CustomUrl
		httpOption.Headers = http.Header{
			"Authorization": []string{"Bearer " + *conf.BaseConfInfo.GeminiToken},
		}
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient:  httpClient,
		APIKey:      *conf.BaseConfInfo.GeminiToken,
		HTTPOptions: httpOption,
	})
}
