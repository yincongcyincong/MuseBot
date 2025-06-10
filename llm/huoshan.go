package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

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

type HuoshanReq struct {
	ToolCall           []*model.ToolCall
	ToolMessage        []*model.ChatCompletionMessage
	CurrentToolMessage []*model.ChatCompletionMessage
}

func (h *HuoshanReq) CallLLMAPI(ctx context.Context, prompt string, l *LLM) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)

	messages := h.getMessages(userId, prompt)

	logger.Info("msg receive", "userID", userId, "prompt", l.Content)
	return h.send(ctx, messages, l)
}

func (h *HuoshanReq) getMessages(userId int64, prompt string) []*model.ChatCompletionMessage {
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

	return messages
}

func (h *HuoshanReq) send(ctx context.Context, messages []*model.ChatCompletionMessage, l *LLM) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	// set deepseek proxy
	httpClient := utils.GetDeepseekProxyClient()

	client := arkruntime.NewClientWithApiKey(
		*conf.DeepseekToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)

	req := model.ChatCompletionRequest{
		Model:    *conf.Type,
		Messages: messages,
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
		Tools:            conf.VolTools,
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

			if !hasTools {
				l.sendMsg(msgInfoContent, choice.Delta.Content)
			}
		}

		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}

	}

	if !hasTools || len(h.CurrentToolMessage) == 0 {
		l.MessageChan <- msgInfoContent

		data, _ := json.Marshal(h.ToolMessage)
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Content:  string(data),
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
		messages = append(messages, h.CurrentToolMessage...)
		h.CurrentToolMessage = make([]*model.ChatCompletionMessage, 0)
		h.ToolCall = make([]*model.ToolCall, 0)
		return h.send(ctx, messages, l)
	}

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (h *HuoshanReq) requestToolsCall(ctx context.Context, choice *model.ChatCompletionStreamChoice) error {
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
			logger.Warn("get mcp fail", "err", err)
			return err
		}

		toolsData, err := mc.ExecTools(ctx, h.ToolCall[len(h.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		h.CurrentToolMessage = append(h.CurrentToolMessage, &model.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: &model.ChatCompletionMessageContent{
				StringValue: &toolsData,
			},
			ToolCallID: h.ToolCall[len(h.ToolCall)-1].ID,
		})
	}

	logger.Info("send tool request", "function", h.ToolCall[len(h.ToolCall)-1].Function.Name,
		"toolCall", h.ToolCall[len(h.ToolCall)-1].ID, "argument", h.ToolCall[len(h.ToolCall)-1].Function.Arguments)

	return nil
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
	json.Unmarshal(respByte, data)

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
