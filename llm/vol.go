package llm

import (
	"context"
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
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type VolReq struct {
	ToolCall           []*model.ToolCall
	ToolMessage        []*model.ChatCompletionMessage
	CurrentToolMessage []*model.ChatCompletionMessage
	
	VolMsgs []*model.ChatCompletionMessage
}

func (h *VolReq) CallLLMAPI(ctx context.Context, prompt string, l *LLM) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	
	h.GetMessages(userId, prompt)
	
	logger.Info("msg receive", "userID", userId, "prompt", l.Content)
	return h.Send(ctx, l)
}

func (h *VolReq) GetModel(l *LLM) {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	
	l.Model = param.ModelDeepSeekR1_528
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.VolModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (h *VolReq) GetMessages(userId int64, prompt string) {
	messages := make([]*model.ChatCompletionMessage, 0)
	
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
				
				messages = append(messages, &model.ChatCompletionMessage{
					Role: constants.ChatMessageRoleUser,
					Content: &model.ChatCompletionMessageContent{
						StringValue: &record.Question,
					},
				})
				
				if record.Content != "" {
					toolsMsgs := make([]*model.ChatCompletionMessage, 0)
					err := json.Unmarshal([]byte(record.Content), &toolsMsgs)
					if err != nil {
						logger.Error("Error unmarshalling tools json", "err", err)
					} else {
						messages = append(messages, toolsMsgs...)
					}
				}
				
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
	if l.LoopNum > MostLoop {
		return errors.New("too many loops")
	}
	l.LoopNum++
	
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	h.GetModel(l)
	
	// set deepseek proxy
	httpClient := utils.GetDeepseekProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	req := model.ChatCompletionRequest{
		Model:    l.Model,
		Messages: h.VolMsgs,
		StreamOptions: &model.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
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
			logger.Info("stream finished", "updateMsgID", updateMsgID)
			break
		}
		if err != nil {
			logger.Error("stream error:", "updateMsgID", updateMsgID, "err", err)
			break
		}
		for _, choice := range response.Choices {
			
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = h.requestToolsCall(ctx, choice)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
					}
				}
			}
			
			if len(choice.Delta.Content) > 0 {
				msgInfoContent = l.sendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
		
	}
	
	if len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	if !hasTools || len(h.CurrentToolMessage) == 0 {
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Token:    l.Token,
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

func (h *VolReq) requestToolsCall(ctx context.Context, choice *model.ChatCompletionStreamChoice) error {
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
		
		if toolCall.Function.Arguments != "" {
			h.ToolCall[len(h.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(h.ToolCall[len(h.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		mc, err := clients.GetMCPClientByToolName(h.ToolCall[len(h.ToolCall)-1].Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "function", h.ToolCall[len(h.ToolCall)-1].Function.Name,
				"toolCall", h.ToolCall[len(h.ToolCall)-1].ID, "argument", h.ToolCall[len(h.ToolCall)-1].Function.Arguments)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, h.ToolCall[len(h.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "function", h.ToolCall[len(h.ToolCall)-1].Function.Name,
				"toolCall", h.ToolCall[len(h.ToolCall)-1].ID, "argument", h.ToolCall[len(h.ToolCall)-1].Function.Arguments)
			return err
		}
		h.CurrentToolMessage = append(h.CurrentToolMessage, &model.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: &model.ChatCompletionMessageContent{
				StringValue: &toolsData,
			},
			ToolCallID: h.ToolCall[len(h.ToolCall)-1].ID,
		})
		
		logger.Info("send tool request", "function", h.ToolCall[len(h.ToolCall)-1].Function.Name,
			"toolCall", h.ToolCall[len(h.ToolCall)-1].ID, "argument", h.ToolCall[len(h.ToolCall)-1].Function.Arguments,
			"res", toolsData)
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
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	h.GetModel(l)
	
	httpClient := utils.GetDeepseekProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	req := model.ChatCompletionRequest{
		Model:    l.Model,
		Messages: h.VolMsgs,
		StreamOptions: &model.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Tools:            l.VolTools,
	}
	
	response, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.Error("CreateChatCompletion error", "updateMsgID", updateMsgID, "err", err)
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
		h.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls)
	}
	
	return *response.Choices[0].Message.Content.StringValue, nil
}

func (h *VolReq) requestOneToolsCall(ctx context.Context, toolsCall []*model.ToolCall) {
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
	}
}

// GenerateImg generate image
func GenerateImg(prompt string) (*param.ImgResponse, error) {
	start := time.Now()
	visual.DefaultInstance.Client.SetAccessKey(*conf.VolcAK)
	visual.DefaultInstance.Client.SetSecretKey(*conf.VolcSK)
	
	reqBody := map[string]interface{}{
		"req_key":           *conf.ReqKey,
		"prompt":            prompt,
		"model_version":     *conf.ModelVersion,
		"req_schedule_conf": *conf.ReqScheduleConf,
		"llm_seed":          *conf.Seed,
		"seed":              *conf.Seed,
		"scale":             *conf.Scale,
		"ddim_steps":        *conf.DDIMSteps,
		"width":             *conf.Width,
		"height":            *conf.Height,
		"use_pre_llm":       *conf.UsePreLLM,
		"use_sr":            *conf.UseSr,
		"return_url":        *conf.ReturnUrl,
		"logo_info": map[string]interface{}{
			"add_logo":          *conf.AddLogo,
			"position":          *conf.Position,
			"language":          *conf.Language,
			"opacity":           *conf.Opacity,
			"logo_text_content": *conf.LogoTextContent,
		},
	}
	
	resp, _, err := visual.DefaultInstance.CVProcess(reqBody)
	if err != nil {
		logger.Error("request img api fail", "err", err)
		return nil, err
	}
	
	respByte, _ := json.Marshal(resp)
	data := &param.ImgResponse{}
	err = json.Unmarshal(respByte, data)
	if err != nil {
		logger.Error("unmarshal response fail", "err", err)
		return nil, err
	}
	
	logger.Info("image response", "respByte", respByte)
	
	// generate image time costing
	totalDuration := time.Since(start).Seconds()
	metrics.ImageDuration.Observe(totalDuration)
	return data, nil
}

func GenerateVideo(prompt string) (string, error) {
	if prompt == "" {
		logger.Warn("prompt is empty", "prompt", prompt)
		return "", errors.New("prompt is empty")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	httpClient := utils.GetDeepseekProxyClient()
	
	client := arkruntime.NewClientWithApiKey(
		*conf.VideoToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	
	videoParam := fmt.Sprintf(" --ratio %s --fps %d  --dur %d --resolution %s --watermark %t",
		*conf.Radio, *conf.FPS, *conf.Duration, *conf.Resolution, *conf.Watermark)
	
	text := prompt + videoParam
	resp, err := client.CreateContentGenerationTask(ctx, model.CreateContentGenerationTaskRequest{
		Model: *conf.VideoModel,
		Content: []*model.CreateContentGenerationContentItem{
			{
				Type: model.ContentGenerationContentItemTypeText,
				Text: &text,
			},
		},
	})
	if err != nil {
		logger.Error("request create video api fail", "err", err)
		return "", err
	}
	
	for {
		getResp, err := client.GetContentGenerationTask(ctx, model.GetContentGenerationTaskRequest{
			ID: resp.ID,
		})
		
		if err != nil {
			logger.Error("request get video api fail", "err", err)
			return "", err
		}
		
		if getResp.Status == model.StatusRunning || getResp.Status == model.StatusQueued {
			logger.Info("video is createing...")
			time.Sleep(5 * time.Second)
			continue
		}
		
		if getResp.Error != nil {
			logger.Error("request get video api fail", "err", getResp.Error)
			return "", errors.New(getResp.Error.Message)
		}
		
		if getResp.Status == model.StatusSucceeded {
			return getResp.Content.VideoURL, nil
		} else {
			logger.Error("request get video api fail", "status", getResp.Status)
			return "", errors.New("create video fail")
		}
	}
}
