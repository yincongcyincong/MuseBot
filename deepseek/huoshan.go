package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type ImgResponse struct {
	Code    int              `json:"code"`
	Data    *ImgResponseData `json:"data"`
	Message string           `json:"message"`
	Status  string           `json:"status"`
}

type ImgResponseData struct {
	AlgorithmBaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_message"`
	} `json:"algorithm_base_resp"`
	ImageUrls        []string `json:"image_urls"`
	PeResult         string   `json:"pe_result"`
	PredictTagResult string   `json:"predict_tag_result"`
	RephraserResult  string   `json:"rephraser_result"`
}

type HuoshanReq struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
}

func (h *HuoshanReq) GetContent() {
	// check user chat exceed max count
	if utils.CheckUserChatExceed(h.Update, h.Bot) {
		return
	}

	defer func() {
		utils.DecreaseUserChat(h.Update)
		close(h.MessageChan)
	}()

	text := strings.ReplaceAll(h.Content, "@"+h.Bot.Self.UserName, "")
	err := h.getContentFromHS(text)
	if err != nil {
		logger.Error("Error calling DeepSeek API", "err", err)
	}

}

func (h *HuoshanReq) getContentFromHS(prompt string) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(h.Update)

	messages := make([]*model.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > 10 {
			aqs = aqs[len(aqs)-10:]
		}
		for i, record := range aqs {
			if record.Answer != "" && record.Question != "" {
				logger.Info("context content", "dialog", i, "question:", record.Question, "answer:", record.Answer)
				messages = append(messages, &model.ChatCompletionMessage{
					Role: constants.ChatMessageRoleAssistant,
					Content: &model.ChatCompletionMessageContent{
						StringValue: &record.Answer,
					},
				})
				messages = append(messages, &model.ChatCompletionMessage{
					Role: constants.ChatMessageRoleUser,
					Content: &model.ChatCompletionMessageContent{
						StringValue: &record.Question,
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

	// set deepseek proxy
	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if *conf.DeepseekProxy != "" {
		proxy, err := url.Parse(*conf.DeepseekProxy)
		if err != nil {
			logger.Error("parse deepseek proxy error", "err", err)
		} else {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	client := arkruntime.NewClientWithApiKey(
		*conf.DeepseekToken,
		arkruntime.WithTimeout(30*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
	ctx := context.Background()
	req := model.ChatCompletionRequest{
		Model:    *conf.DeepseekType,
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
	}

	logger.Info("msg receive", "userID", userId, "prompt", prompt)
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.Error("standard chat error", "err", err)
		return err
	}
	defer stream.Close()

	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

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
			// exceed max telegram one message length
			if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
				h.MessageChan <- msgInfoContent
				msgInfoContent = &param.MsgInfo{
					SendLen:     NonFirstSendLen,
					FullContent: msgInfoContent.FullContent,
					Token:       msgInfoContent.Token,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			msgInfoContent.FullContent += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				h.MessageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}

		if response.Usage != nil {
			msgInfoContent.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(msgInfoContent.Token))
		}

	}

	h.MessageChan <- msgInfoContent

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

// GenerateImg generate image
func GenerateImg(prompt string) (*ImgResponse, error) {
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
	data := &ImgResponse{}
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

	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if *conf.DeepseekProxy != "" {
		proxy, err := url.Parse(*conf.DeepseekProxy)
		if err != nil {
			logger.Error("parse deepseek proxy error", "err", err)
		} else {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	client := arkruntime.NewClientWithApiKey(
		*conf.VideoToken,
		arkruntime.WithTimeout(30*time.Minute),
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
