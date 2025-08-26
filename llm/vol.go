package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
)

type VolReq struct {
	ToolCall           []*model.ToolCall
	ToolMessage        []*model.ChatCompletionMessage
	CurrentToolMessage []*model.ChatCompletionMessage
	
	VolMsgs []*model.ChatCompletionMessage
}

func (h *VolReq) GetModel(l *LLM) {
	l.Model = param.ModelDeepSeekR1_528
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.VolModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (h *VolReq) GetMessages(userId string, prompt string) {
	messages := make([]*model.ChatCompletionMessage, 0)
	
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > *conf.BaseConfInfo.MaxQAPair {
			aqs = aqs[len(aqs)-*conf.BaseConfInfo.MaxQAPair:]
		}
		for i, record := range aqs {
			
			logger.Info("context content", "dialog", i, "question:", record.Question,
				"toolContent", record.Content, "answer:", record.Answer)
			
			if record.Question != "" {
				messages = append(messages, &model.ChatCompletionMessage{
					Role: constants.ChatMessageRoleUser,
					Content: &model.ChatCompletionMessageContent{
						StringValue: &record.Question,
					},
				})
			}
			
			if record.Content != "" {
				toolsMsgs := make([]*model.ChatCompletionMessage, 0)
				err := json.Unmarshal([]byte(record.Content), &toolsMsgs)
				if err != nil {
					logger.Error("Error unmarshalling tools json", "err", err)
				} else {
					messages = append(messages, toolsMsgs...)
				}
			}
			
			if record.Answer != "" {
				messages = append(messages, &model.ChatCompletionMessage{
					Role: constants.ChatMessageRoleAssistant,
					Content: &model.ChatCompletionMessageContent{
						StringValue: &record.Answer,
					},
				})
			}
		}
	}
	messages = append(messages, &model.ChatCompletionMessage{
		Role: constants.ChatMessageRoleUser,
		Content: &model.ChatCompletionMessageContent{
			StringValue: &prompt,
		},
	})
	
	h.VolMsgs = messages
}

func (h *VolReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	h.GetModel(l)
	
	// set deepseek proxy
	httpClient := utils.GetLLMProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.BaseConfInfo.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	req := model.ChatCompletionRequest{
		Model:    l.Model,
		Messages: h.VolMsgs,
		StreamOptions: &model.StreamOptions{
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
		Tools:            l.VolTools,
	}
	
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.Error("standard chat error", "err", err)
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
			logger.Info("stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.Error("stream error:", "updateMsgID", l.MsgId, "err", err)
			return err
		}
		for _, choice := range response.Choices {
			
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = h.requestToolsCall(ctx, choice, l)
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
	
	if !hasTools || len(h.CurrentToolMessage) == 0 {
		db.InsertMsgRecord(l.UserId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Token:    l.Token,
			Mode:     param.Vol,
		}, true)
	} else {
		h.CurrentToolMessage = append([]*model.ChatCompletionMessage{
			{
				Role: deepseek.ChatMessageRoleAssistant,
				Content: &model.ChatCompletionMessageContent{
					StringValue: &l.WholeContent,
				},
				ToolCalls: h.ToolCall,
			},
		}, h.CurrentToolMessage...)
		
		h.ToolMessage = append(h.ToolMessage, h.CurrentToolMessage...)
		h.VolMsgs = append(h.VolMsgs, h.CurrentToolMessage...)
		h.CurrentToolMessage = make([]*model.ChatCompletionMessage, 0)
		h.ToolCall = make([]*model.ToolCall, 0)
		return h.Send(ctx, l)
	}
	
	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (h *VolReq) requestToolsCall(ctx context.Context, choice *model.ChatCompletionStreamChoice, l *LLM) error {
	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})
		
		if toolCall.Function.Name != "" {
			h.ToolCall = append(h.ToolCall, toolCall)
			h.ToolCall[len(h.ToolCall)-1].Function.Name = toolCall.Function.Name
		}
		
		if toolCall.ID != "" {
			h.ToolCall[len(h.ToolCall)-1].ID = toolCall.ID
		}
		
		if toolCall.Type != "" {
			h.ToolCall[len(h.ToolCall)-1].Type = toolCall.Type
		}
		
		if toolCall.Function.Arguments != "" && toolCall.Function.Name == "" {
			h.ToolCall[len(h.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(h.ToolCall[len(h.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		tool := h.ToolCall[len(h.ToolCall)-1]
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
		h.CurrentToolMessage = append(h.CurrentToolMessage, &model.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: &model.ChatCompletionMessageContent{
				StringValue: &toolsData,
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

func (h *VolReq) GetUserMessage(msg string) {
	h.GetMessage(constants.ChatMessageRoleUser, msg)
}

func (h *VolReq) GetAssistantMessage(msg string) {
	h.GetMessage(constants.ChatMessageRoleAssistant, msg)
}

func (h *VolReq) AppendMessages(client LLMClient) {
	if len(h.VolMsgs) == 0 {
		h.VolMsgs = make([]*model.ChatCompletionMessage, 0)
	}
	
	h.VolMsgs = append(h.VolMsgs, client.(*VolReq).VolMsgs...)
}

func (h *VolReq) GetMessage(role, msg string) {
	if len(h.VolMsgs) == 0 {
		h.VolMsgs = []*model.ChatCompletionMessage{
			{
				Role: role,
				Content: &model.ChatCompletionMessageContent{
					StringValue: &msg,
				},
			},
		}
		return
	}
	
	h.VolMsgs = append(h.VolMsgs, &model.ChatCompletionMessage{
		Role: role,
		Content: &model.ChatCompletionMessageContent{
			StringValue: &msg,
		},
	})
}

func (h *VolReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	h.GetModel(l)
	
	httpClient := utils.GetLLMProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.BaseConfInfo.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	req := model.ChatCompletionRequest{
		Model:    l.Model,
		Messages: h.VolMsgs,
		StreamOptions: &model.StreamOptions{
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
		Tools:            l.VolTools,
	}
	
	response, err := client.CreateChatCompletion(ctx, req)
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
		h.GetAssistantMessage("")
		h.VolMsgs[len(h.VolMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		h.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
	}
	
	return *response.Choices[0].Message.Content.StringValue, nil
}

func (h *VolReq) requestOneToolsCall(ctx context.Context, toolsCall []*model.ToolCall, l *LLM) {
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
		
		h.VolMsgs = append(h.VolMsgs, &model.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: &model.ChatCompletionMessageContent{
				StringValue: &toolsData,
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

// GenerateVolImg generate image
func GenerateVolImg(prompt string, imageContent []byte) (string, int, error) {
	start := time.Now()
	visual.DefaultInstance.Client.SetAccessKey(*conf.BaseConfInfo.VolcAK)
	visual.DefaultInstance.Client.SetSecretKey(*conf.BaseConfInfo.VolcSK)
	
	reqBody := map[string]interface{}{
		"req_key":           *conf.PhotoConfInfo.ReqKey,
		"prompt":            prompt,
		"model_version":     *conf.PhotoConfInfo.ModelVersion,
		"req_schedule_conf": *conf.PhotoConfInfo.ReqScheduleConf,
		"llm_seed":          *conf.PhotoConfInfo.Seed,
		"seed":              *conf.PhotoConfInfo.Seed,
		"scale":             *conf.PhotoConfInfo.Scale,
		"ddim_steps":        *conf.PhotoConfInfo.DDIMSteps,
		"width":             *conf.PhotoConfInfo.Width,
		"height":            *conf.PhotoConfInfo.Height,
		"use_pre_llm":       *conf.PhotoConfInfo.UsePreLLM,
		"use_sr":            *conf.PhotoConfInfo.UseSr,
		"return_url":        *conf.PhotoConfInfo.ReturnUrl,
		"logo_info": map[string]interface{}{
			"add_logo":          *conf.PhotoConfInfo.AddLogo,
			"position":          *conf.PhotoConfInfo.Position,
			"language":          *conf.PhotoConfInfo.Language,
			"opacity":           *conf.PhotoConfInfo.Opacity,
			"logo_text_content": *conf.PhotoConfInfo.LogoTextContent,
		},
	}
	
	if len(imageContent) != 0 {
		reqBody["binary_data_base64"] = []string{base64.StdEncoding.EncodeToString(imageContent)}
	}
	
	resp, _, err := visual.DefaultInstance.CVProcess(reqBody)
	if err != nil {
		logger.Error("request img api fail", "err", err)
		return "", 0, err
	}
	
	respByte, _ := json.Marshal(resp)
	data := &param.ImgResponse{}
	err = json.Unmarshal(respByte, data)
	if err != nil {
		logger.Error("unmarshal response fail", "err", err)
		return "", 0, err
	}
	
	logger.Info("image response", "respByte", respByte)
	
	// generate image time costing
	totalDuration := time.Since(start).Seconds()
	metrics.ImageDuration.Observe(totalDuration)
	
	if data.Data == nil || len(data.Data.ImageUrls) == 0 {
		logger.Warn("no image generated")
		return "", 0, errors.New("no image generated")
	}
	
	return data.Data.ImageUrls[0], param.ImageTokenUsage, nil
}

// GenerateVolVideo generate video
func GenerateVolVideo(prompt string, imageContent []byte) (string, int, error) {
	if prompt == "" {
		logger.Warn("prompt is empty", "prompt", prompt)
		return "", 0, errors.New("prompt is empty")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.BaseConfInfo.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	videoParam := fmt.Sprintf(" --ratio %s --fps %d  --dur %d --resolution %s --watermark %t",
		*conf.VideoConfInfo.Radio, *conf.VideoConfInfo.FPS, *conf.VideoConfInfo.Duration, *conf.VideoConfInfo.Resolution, *conf.VideoConfInfo.Watermark)
	
	text := prompt + videoParam
	contents := make([]*model.CreateContentGenerationContentItem, 0)
	contents = append(contents, &model.CreateContentGenerationContentItem{
		Type: model.ContentGenerationContentItemTypeText,
		Text: &text,
	})
	
	if len(imageContent) > 0 {
		frame := "first_frame"
		contents = append(contents, &model.CreateContentGenerationContentItem{
			Type: model.ContentGenerationContentItemTypeImage,
			ImageURL: &model.ImageURL{
				URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
			},
			Role: &frame,
		})
	}
	
	resp, err := client.CreateContentGenerationTask(ctx, model.CreateContentGenerationTaskRequest{
		Model:   *conf.VideoConfInfo.VideoModel,
		Content: contents,
	})
	if err != nil {
		logger.Error("request create video api fail", "err", err)
		return "", 0, err
	}
	
	for {
		getResp, err := client.GetContentGenerationTask(ctx, model.GetContentGenerationTaskRequest{
			ID: resp.ID,
		})
		
		if err != nil {
			logger.Error("request get video api fail", "err", err)
			return "", 0, err
		}
		
		if getResp.Status == model.StatusRunning || getResp.Status == model.StatusQueued {
			logger.Info("video is createing...")
			time.Sleep(5 * time.Second)
			continue
		}
		
		if getResp.Error != nil {
			logger.Error("request get video api fail", "err", getResp.Error)
			return "", 0, errors.New(getResp.Error.Message)
		}
		
		if getResp.Status == model.StatusSucceeded {
			return getResp.Content.VideoURL, getResp.Usage.TotalTokens, nil
		} else {
			logger.Error("request get video api fail", "status", getResp.Status)
			return "", 0, errors.New("create video fail")
		}
	}
}

func GetVolImageContent(imageContent []byte, content string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetLLMProxyClient()
	client := arkruntime.NewClientWithApiKey(
		*conf.BaseConfInfo.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	contentPrompt := content
	if content == "" {
		contentPrompt = i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_handle_prompt", nil)
	}
	
	req := model.ChatCompletionRequest{
		Model: *conf.PhotoConfInfo.VolImageModel,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					ListValue: []*model.ChatCompletionMessageContentPart{
						{
							Type: model.ChatCompletionMessageContentPartTypeImageURL,
							ImageURL: &model.ChatMessageImageURL{
								URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
							},
						},
						{
							Type: model.ChatCompletionMessageContentPartTypeText,
							Text: contentPrompt,
						},
					},
				},
			},
		},
	}
	
	response, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.Error("CreateChatCompletion error", "err", err)
		return "", 0, err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", 0, errors.New("response is empty")
	}
	
	return *response.Choices[0].Message.Content.StringValue, response.Usage.TotalTokens, nil
}
