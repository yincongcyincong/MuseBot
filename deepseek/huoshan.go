package deepseek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

func GetContentFromHS(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	text := strings.ReplaceAll(content, "@"+bot.Self.UserName, "")
	err := getContentFromHS(text, update, messageChan)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %s\n", err)
	}
	close(messageChan)
}

func getContentFromHS(prompt string, update tgbotapi.Update, messageChan chan *param.MsgInfo) error {

	_, updateMsgID, username := utils.GetChatIdAndMsgIdAndUserName(update)

	messages := make([]*model.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(username)
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
		//通过 os.Getenv 从环境变量中获取 ARK_API_KEY
		*conf.DeepseekToken,
		//深度推理模型耗费时间会较长，请您设置较大的超时时间，避免超时导致任务失败。推荐30分钟以上
		arkruntime.WithTimeout(30*time.Minute),
	)
	// 创建一个上下文，通常用于传递请求的上下文信息，如超时、取消等
	ctx := context.Background()
	// 构建聊天完成请求，设置请求的模型和消息内容
	req := model.ChatCompletionRequest{
		// 需要替换 <Model> 为模型的Model ID
		Model:    *conf.DeepseekType,
		Messages: messages,
	}

	fmt.Printf("[%s]: %s\n", username, prompt)
	// 发送聊天完成请求，并将结果存储在 resp 中，将可能出现的错误存储在 err 中
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		// 若出现错误，打印错误信息并终止程序
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
	return nil
}
