package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"

	"io"
	"log"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
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

func GetContentFromHS(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	text := strings.ReplaceAll(content, "@"+bot.Self.UserName, "")
	err := getContentFromHS(text, update, messageChan)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %s\n", err)
	}
	close(messageChan)
}

func getContentFromHS(prompt string, update tgbotapi.Update, messageChan chan *param.MsgInfo) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(update)

	messages := make([]*model.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		for _, record := range msgRecords.AQs {
			if record.Answer != "" && record.Question != "" {
				log.Println("question:", record.Question, "answer:", record.Answer)
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

	client := arkruntime.NewClientWithApiKey(
		*conf.DeepseekToken,
		arkruntime.WithTimeout(30*time.Minute),
	)
	ctx := context.Background()
	req := model.ChatCompletionRequest{
		Model:    *conf.DeepseekType,
		Messages: messages,
	}

	fmt.Printf("[%d]: %s\n", userId, prompt)
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("standard chat error: %v\n", err)
		return err
	}
	defer stream.Close()

	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\n %d Stream finished", updateMsgID)
			break
		}
		if err != nil {
			fmt.Printf("\n %d Stream error: %v\n", updateMsgID, err)
			break
		}
		for _, choice := range response.Choices {
			// exceed max telegram one message length
			if len(msgInfoContent.Content) > OneMsgLen {
				messageChan <- msgInfoContent
				msgInfoContent = &param.MsgInfo{
					SendLen:     FirstSendLen,
					FullContent: msgInfoContent.FullContent,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			msgInfoContent.FullContent += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				messageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}
	}

	messageChan <- msgInfoContent

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
		"req_key":           "high_aes_general_v21_L",
		"prompt":            prompt,
		"model_version":     "general_v2.1_L",
		"req_schedule_conf": "general_v20_9B_pe",
		"llm_seed":          -1,
		"seed":              -1,
		"scale":             3.5,
		"ddim_steps":        25,
		"width":             512,
		"height":            512,
		"use_pre_llm":       true,
		"use_sr":            true,
		"sr_seed":           -1,
		"sr_strength":       0.4,
		"sr_scale":          3.5,
		"sr_steps":          20,
		"is_only_sr":        false,
		"return_url":        true,
		"logo_info": map[string]interface{}{
			"add_logo":          false,
			"position":          0,
			"language":          0,
			"opacity":           0.3,
			"logo_text_content": "",
		},
	}

	resp, _, err := visual.DefaultInstance.CVProcess(reqBody)
	if err != nil {
		log.Printf("request img api fail: %s\n", err)
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
