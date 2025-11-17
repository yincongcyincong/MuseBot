package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/devinyf/dashscopego/qwen"
	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type OpenAIReq struct {
	ToolCall           []openai.ToolCall
	ToolMessage        []openai.ChatCompletionMessage
	CurrentToolMessage []openai.ChatCompletionMessage
	
	OpenAIMsgs []openai.ChatCompletionMessage
}

func (d *OpenAIReq) GetModel(l *LLM) {
	userInfo := db.GetCtxUserInfo(l.Ctx)
	model := ""
	if userInfo != nil && userInfo.LLMConfigRaw != nil {
		model = userInfo.LLMConfigRaw.TxtModel
	}
	
	switch utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw) {
	case param.OpenAi, param.ChatAnyWhere:
		l.Model = openai.GPT3Dot5Turbo0125
		if userInfo != nil && model != "" {
			l.Model = model
		}
	case param.Aliyun:
		l.Model = qwen.QwenMax
		if userInfo != nil && model != "" && param.AliyunModel[model] {
			l.Model = model
		}
	case param.DeepSeek:
		l.Model = deepseek.DeepSeekChat
		if userInfo != nil && model != "" && param.DeepseekModels[model] {
			logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", model)
			l.Model = model
		}
	case param.Vol:
		l.Model = param.ModelDoubao15VisionPro328
		if userInfo != nil && model != "" && param.VolModels[model] {
			logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", model)
			l.Model = model
		}
	case param.AI302:
		l.Model = openai.GPT3Dot5Turbo
		if userInfo != nil && model != "" {
			l.Model = model
		}
	case param.OpenRouter:
		l.Model = param.DeepseekDeepseekR1_0528Free
		if userInfo != nil && model != "" {
			l.Model = model
		}
	}
}

func (d *OpenAIReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	
	client := GetOpenAIClient(ctx, "txt")
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
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	hasTools := false
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.InfoCtx(l.Ctx, "Stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.WarnCtx(l.Ctx, "Stream error", "updateMsgID", l.MsgId, "err", err)
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
						logger.ErrorCtx(l.Ctx, "requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
					}
				}
			}
			
			if !hasTools {
				msgInfoContent = l.SendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Cs.Token += response.Usage.TotalTokens
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 || (hasTools && *conf.BaseConfInfo.SendMcpRes) {
		if *conf.BaseConfInfo.Powered != "" {
			msgInfoContent.Content = msgInfoContent.Content + "\n\n" + *conf.BaseConfInfo.Powered
		}
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
	start := time.Now()
	
	client := GetOpenAIClient(ctx, "txt")
	
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
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.ErrorCtx(l.Ctx, "response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Cs.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.OpenAIMsgs[len(d.OpenAIMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
		return d.SyncSend(ctx, l)
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
		
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err, "toolCall", tool)
			return
		}
		
		d.OpenAIMsgs = append(d.OpenAIMsgs, openai.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
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
		
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.ErrorCtx(ctx, "Error executing MCP request", "toolId", toolCall.ID, "err", err)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, openai.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}
	
	return nil
	
}

// GenerateOpenAIImg generate image
func GenerateOpenAIImg(ctx context.Context, prompt string, imageContent []byte) ([]byte, int, error) {
	client := GetOpenAIClient(ctx, "img")
	
	start := time.Now()
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetImgType(llmConfig)
	model := utils.GetUsingImgModel(mediaType, llmConfig.ImgModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	var respUrl openai.ImageResponse
	var err error
	if len(imageContent) != 0 {
		imageFile, err := utils.ConvertToPNGFile(imageContent)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to create temp file:", err)
			return nil, 0, err
		}
		defer os.Remove(imageFile.Name())
		defer imageFile.Close()
		
		respUrl, err = client.CreateEditImage(ctx, openai.ImageEditRequest{
			Image:  imageFile,
			Prompt: prompt,
			Model:  model,
			N:      1,
			Size:   *conf.PhotoConfInfo.OpenAIImageSize,
		})
	} else {
		respUrl, err = client.CreateImage(
			ctx,
			openai.ImageRequest{
				Prompt: prompt,
				Model:  model,
				Size:   *conf.PhotoConfInfo.OpenAIImageSize,
				N:      1,
				Style:  *conf.PhotoConfInfo.OpenAIImageStyle,
			},
		)
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if err != nil {
		logger.ErrorCtx(ctx, "CreateImage error", "err", err)
		return nil, 0, err
	}
	
	if len(respUrl.Data) == 0 {
		logger.ErrorCtx(ctx, "response is emtpy", "response", respUrl)
		return nil, 0, errors.New("response is empty")
	}
	
	imageContentByte, err := base64.StdEncoding.DecodeString(respUrl.Data[0].B64JSON)
	if err != nil {
		logger.ErrorCtx(ctx, "decode image error", "err", err)
		return nil, 0, err
	}
	
	return imageContentByte, respUrl.Usage.TotalTokens, nil
}

func GenerateOpenAIText(ctx context.Context, audioContent []byte) (string, error) {
	
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(openai.Whisper1).Inc()
	
	client := GetOpenAIClient(ctx, "rec")
	
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "voice." + utils.DetectAudioFormat(audioContent),
		Reader:   bytes.NewReader(audioContent),
		Format:   "json",
	}
	
	resp, err := client.CreateTranscription(ctx, req)
	
	metrics.APIRequestDuration.WithLabelValues(openai.Whisper1).Observe(time.Since(start).Seconds())
	
	if err != nil {
		logger.ErrorCtx(ctx, "CreateTranscription error", "err", err)
		return "", err
	}
	
	return resp.Text, nil
}

func GetOpenAIImageContent(ctx context.Context, imageContent []byte, content string) (string, int, error) {
	
	client := GetOpenAIClient(ctx, "rec")
	
	contentPrompt := content
	
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetImgType(llmConfig)
	model := utils.GetUsingImgModel(mediaType, llmConfig.ImgModel)
	
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	imageDataURL := fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(imageContent), base64.StdEncoding.EncodeToString(imageContent))
	req := openai.ChatCompletionRequest{
		Model: model,
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
	}
	
	resp, err := client.CreateChatCompletion(ctx, req)
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(ctx, "CreateChatCompletion error", "err", err)
		return "", 0, err
	}
	
	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}

func OpenAITTS(ctx context.Context, content, encoding string) ([]byte, int, int, error) {
	formatEncoding := encoding
	if encoding != string(openai.SpeechResponseFormatOpus) && encoding != string(openai.SpeechResponseFormatAac) && encoding != string(openai.SpeechResponseFormatFlac) &&
		encoding != string(openai.SpeechResponseFormatWav) && encoding != string(openai.SpeechResponseFormatPcm) {
		formatEncoding = string(openai.SpeechResponseFormatPcm)
	}
	
	start := time.Now()
	model := utils.GetUsingTTSModel(param.OpenAi, db.GetCtxUserInfo(ctx).LLMConfigRaw.TTSModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	client := GetOpenAIClient(ctx, "")
	resp, err := client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model:          openai.SpeechModel(model),
		Input:          content,
		Voice:          openai.SpeechVoice(*conf.AudioConfInfo.OpenAIVoiceName),
		ResponseFormat: openai.SpeechResponseFormat(formatEncoding),
		Speed:          1.0,
	})
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(ctx, "decode image error", "err", err)
		return nil, 0, 0, err
	}
	data, err := io.ReadAll(resp.ReadCloser)
	if err != nil {
		logger.ErrorCtx(ctx, "read response error", "err", err)
		return nil, 0, 0, err
	}
	
	if formatEncoding == string(openai.SpeechResponseFormatPcm) {
		data, err = utils.GetAudioData(encoding, data)
		if err != nil {
			logger.ErrorCtx(ctx, "GetAudioData error", "err", err)
			return nil, 0, 0, err
		}
	}
	
	return data, db.EstimateTokens(content), utils.PCMDuration(len(data), 24000, 1, 16), nil
}

func GetOpenAIClient(ctx context.Context, clientType string) *openai.Client {
	httpClient := utils.GetLLMProxyClient()
	t := param.OpenAi
	switch clientType {
	case "txt":
		t = utils.GetTxtType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "img":
		t = utils.GetImgType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "video":
		t = utils.GetVideoType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "rec":
		t = utils.GetRecType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	}
	
	var token string
	var specialLLMUrl string
	switch t {
	case param.OpenAi:
		token = *conf.BaseConfInfo.OpenAIToken
	case param.Aliyun:
		token = *conf.BaseConfInfo.AliyunToken
		specialLLMUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	case param.ChatAnyWhere:
		token = *conf.BaseConfInfo.ChatAnyWhereToken
		specialLLMUrl = "https://api.chatanywhere.tech"
	case param.DeepSeek:
		token = *conf.BaseConfInfo.DeepseekToken
		specialLLMUrl = "https://api.deepseek.com/v1"
	case param.Vol:
		token = *conf.BaseConfInfo.VolToken
		specialLLMUrl = "https://ark.cn-beijing.volces.com/api/v3"
	case param.OpenRouter:
		token = *conf.BaseConfInfo.OpenRouterToken
		specialLLMUrl = "https://openrouter.ai/api/v1"
	case param.AI302:
		token = *conf.BaseConfInfo.AI302Token
		specialLLMUrl = "https://api.302.ai/v1"
	}
	
	openaiConfig := openai.DefaultConfig(token)
	if specialLLMUrl != "" {
		openaiConfig.BaseURL = specialLLMUrl
	}
	
	if *conf.BaseConfInfo.CustomUrl != "" {
		openaiConfig.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	openaiConfig.HTTPClient = httpClient
	return openai.NewClientWithConfig(openaiConfig)
}
