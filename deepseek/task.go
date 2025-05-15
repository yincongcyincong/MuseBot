package deepseek

import (
	"context"
	"encoding/json"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
	"net/http"
	"net/url"
	"time"
)

type DeepseekTaskReq struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
	Model       string
	Token       int
}

type TaskInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (d *DeepseekTaskReq) ExecuteTask() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	d.Model = deepseek.DeepSeekChat
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)

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

	client, err := deepseek.NewClientWithOptions(*conf.DeepseekToken,
		deepseek.WithBaseURL(*conf.CustomUrl), deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Error creating deepseek client", "err", err)
		return
	}

	taskParam := make(map[string]interface{})
	taskParam["assign_param"] = make([]map[string]string, 0)
	taskParam["user_task"] = d.Content
	for _, tool := range conf.TaskTools {
		taskParam["assign_param"] = append(taskParam["assign_param"].([]map[string]string), map[string]string{
			"tool_name": tool.Name,
			"tool_desc": tool.Description,
		})
	}

	messages := []deepseek.ChatCompletionMessage{
		{
			Role:    constants.ChatMessageRoleUser,
			Content: i18n.GetMessage(*conf.Lang, "assign_task_prompt", taskParam),
		},
	}

	request := &deepseek.ChatCompletionRequest{
		Model:            d.Model,
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Messages:         messages,
	}

	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return
	}

	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return
	}

	plans := make([]*TaskInfo, 0)
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &plans)
	if err != nil {
		logger.Error("parse content fail", "err", err)
		return
	}

}
