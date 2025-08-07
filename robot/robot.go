package robot

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/slack-go/slack"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	helpText = `
Available Commands:

/chat   - Start a normal chat session

/mode   - Set the LLM mode

/balance - Check your current balance (tokens or credits)

/state  - View your current session state and settings

/clear  - Clear all conversation history

/retry  - Retry your last question

/photo  - Create a Image base on your prompt or your Image

/video  - Generate a video based on your prompt

/task   - Let multiple agents collaborate to complete a task

/mcp    - Use Multi-Agent Control Panel for complex task planning

/help   - Show this help message

`
)

type RobotInfo struct {
	Robot Robot
}

type Robot interface {
	Exec()
}

type botOption func(r *RobotInfo)

func NewRobot(options ...botOption) *RobotInfo {
	r := new(RobotInfo)
	for _, o := range options {
		o(r)
	}
	return r
}

func (r *RobotInfo) Exec() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	if !r.checkUserAllow(userId) && !r.checkGroupAllow(chatId) {
		logger.Warn("user/group not allow to use this bot", "userID", userId, "chat", chatId)
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "valid_user_group", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	r.Robot.Exec()
}

func (r *RobotInfo) GetChatIdAndMsgIdAndUserID() (string, string, string) {
	chatId := ""
	msgId := ""
	userId := ""
	
	switch r.Robot.(type) {
	case *TelegramRobot:
		telegramRobot := r.Robot.(*TelegramRobot)
		if telegramRobot.Update.Message != nil {
			chatId = strconv.FormatInt(telegramRobot.Update.Message.Chat.ID, 10)
			userId = strconv.FormatInt(telegramRobot.Update.Message.From.ID, 10)
			msgId = strconv.Itoa(telegramRobot.Update.Message.MessageID)
		}
		if telegramRobot.Update.CallbackQuery != nil {
			chatId = strconv.FormatInt(telegramRobot.Update.CallbackQuery.Message.Chat.ID, 10)
			userId = strconv.FormatInt(telegramRobot.Update.CallbackQuery.From.ID, 10)
			msgId = strconv.Itoa(telegramRobot.Update.CallbackQuery.Message.MessageID)
		}
	case *DiscordRobot:
		discordRobot := r.Robot.(*DiscordRobot)
		if discordRobot.Msg != nil {
			chatId = discordRobot.Msg.ChannelID
			userId = discordRobot.Msg.Author.ID
			msgId = discordRobot.Msg.Message.ID
		}
		if discordRobot.Inter != nil {
			chatId = discordRobot.Inter.ChannelID
			if discordRobot.Inter.User != nil {
				userId = discordRobot.Inter.User.ID
			}
			if discordRobot.Inter.Member != nil {
				userId = discordRobot.Inter.Member.User.ID
			}
		}
	case *SlackRobot:
		slackRobot := r.Robot.(*SlackRobot)
		if slackRobot != nil {
			chatId = slackRobot.Event.Channel
			userId = slackRobot.Event.User
			msgId = slackRobot.Event.ClientMsgID
		}
	case *LarkRobot:
		lark := r.Robot.(*LarkRobot)
		if lark.Message != nil {
			msgId = larkcore.StringValue(lark.Message.Event.Message.MessageId)
			chatId = larkcore.StringValue(lark.Message.Event.Message.ChatId)
			userId = larkcore.StringValue(lark.Message.Event.Sender.SenderId.UserId)
		}
	case *Web:
		web := r.Robot.(*Web)
		if web != nil {
			chatId = web.RealUserId
			msgId = web.RealUserId
			userId = web.RealUserId
		}
	}
	
	return chatId, msgId, userId
}

func (r *RobotInfo) SendMsg(chatId string, msgContent string, replyToMessageID string,
	mode string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) string {
	switch r.Robot.(type) {
	case *TelegramRobot:
		telegramRobot := r.Robot.(*TelegramRobot)
		msg := tgbotapi.NewMessage(int64(utils.ParseInt(chatId)), msgContent)
		msg.ParseMode = mode
		msg.ReplyMarkup = inlineKeyboard
		msg.ReplyToMessageID = utils.ParseInt(replyToMessageID)
		msgInfo, err := telegramRobot.Bot.Send(msg)
		if err != nil {
			logger.Warn("send clear message fail", "err", err)
			return ""
		}
		return utils.ValueToString(msgInfo.MessageID)
	case *DiscordRobot:
		discordRobot := r.Robot.(*DiscordRobot)
		if discordRobot.Msg != nil {
			messageSend := &discordgo.MessageSend{
				Content: msgContent,
			}
			
			// 设置引用消息
			chatIdStr := fmt.Sprintf("%d", chatId)
			if replyToMessageID != "" {
				messageSend.Reference = &discordgo.MessageReference{
					MessageID: replyToMessageID,
					ChannelID: chatIdStr,
				}
			}
			
			sentMsg, err := discordRobot.Session.ChannelMessageSendComplex(chatIdStr, messageSend)
			if err != nil {
				logger.Warn("send discord message fail", "err", err)
				return ""
			}
			return sentMsg.ID
		}
		
		if discordRobot.Inter != nil {
			var err error
			if mode == param.DiscordEditMode {
				_, err = discordRobot.Session.InteractionResponseEdit(discordRobot.Inter.Interaction, &discordgo.WebhookEdit{
					Content: &msgContent,
				})
			} else {
				err = discordRobot.Session.InteractionRespond(discordRobot.Inter.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: msgContent,
					},
				})
			}
			
			if err != nil {
				logger.Warn("send discord interaction response fail", "err", err)
			}
			return ""
		}
	
	case *SlackRobot:
		slackRobot := r.Robot.(*SlackRobot)
		_, timestamp, err := slackRobot.Client.PostMessage(chatId, slack.MsgOptionText(msgContent, false))
		if err != nil {
			logger.Warn("send message fail", "err", err)
		}
		
		return timestamp
	case *LarkRobot:
		lark := r.Robot.(*LarkRobot)
		content := larkim.NewTextMsgBuilder().
			Text(msgContent).
			Build()
		
		if replyToMessageID != "" {
			resp, err := lark.Client.Im.Message.Reply(lark.Ctx, larkim.NewReplyMessageReqBuilder().
				MessageId(replyToMessageID).
				Body(larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeText).
					Content(content).
					Build()).
				Build())
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return *resp.Data.MessageId
		} else {
			resp, err := lark.Client.Im.Message.Create(lark.Ctx, larkim.NewCreateMessageReqBuilder().
				ReceiveIdType(larkim.ReceiveIdTypeChatId).
				Body(larkim.NewCreateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeText).
					ReceiveId(chatId).
					Content(content).
					Build()).
				Build())
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return *resp.Data.MessageId
		}
	
	case *Web:
		web := r.Robot.(*Web)
		_, err := web.W.Write([]byte(msgContent))
		if err != nil {
			logger.Warn("send message fail", "err", err)
		}
		web.Flusher.Flush()
	}
	
	return ""
}

func WithRobot(robot Robot) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.Robot = robot
	}
}

func StartRobot() {
	if *conf.BaseConfInfo.TelegramBotToken != "" {
		go func() {
			StartTelegramRobot()
		}()
	}
	
	if *conf.BaseConfInfo.DiscordBotToken != "" {
		go func() {
			StartDiscordRobot()
		}()
	}
	
	if *conf.BaseConfInfo.LarkAPPID != "" && *conf.BaseConfInfo.LarkAppSecret != "" {
		go func() {
			StartLarkRobot()
		}()
	}
}

// checkUserAllow check use can use telegram bot or not
func (r *RobotInfo) checkUserAllow(userId string) bool {
	if len(conf.BaseConfInfo.AllowedUserIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedUserIds["0"] {
		return false
	}
	
	_, ok := conf.BaseConfInfo.AllowedUserIds[userId]
	return ok
}

func (r *RobotInfo) checkGroupAllow(chatId string) bool {
	
	if len(conf.BaseConfInfo.AllowedGroupIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedGroupIds["0"] {
		return false
	}
	if _, ok := conf.BaseConfInfo.AllowedGroupIds[chatId]; ok {
		return true
	}
	
	return false
}

// checkUserTokenExceed check use token exceeded
func (r *RobotInfo) checkUserTokenExceed(chatId string, msgId string, userId string) bool {
	if *conf.BaseConfInfo.TokenPerUser == 0 {
		return false
	}
	
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return false
	}
	
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		return false
	}
	
	if userInfo.Token >= userInfo.AvailToken {
		tpl := i18n.GetMessage(*conf.BaseConfInfo.Lang, "token_exceed", nil)
		content := fmt.Sprintf(tpl, userInfo.Token, userInfo.AvailToken-userInfo.Token, userInfo.AvailToken)
		r.SendMsg(chatId, content, msgId, tgbotapi.ModeMarkdown, nil)
		return true
	}
	
	return false
}

// checkAdminUser check user is admin
func (r *RobotInfo) checkAdminUser(userId string) bool {
	if len(conf.BaseConfInfo.AdminUserIds) == 0 {
		return false
	}
	
	_, ok := conf.BaseConfInfo.AdminUserIds[userId]
	return ok
}

func (r *RobotInfo) GetAudioContent(audioContent []byte) (string, error) {
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		return utils.FileRecognize(audioContent)
	case param.OpenAi:
		return llm.GenerateOpenAIText(audioContent)
	case param.Gemini:
		return llm.GenerateGeminiText(audioContent)
	}
	
	return "", nil
}

func (r *RobotInfo) GetImageContent(imageContent []byte) (string, error) {
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		return utils.GetImageContent(imageContent)
	case param.Gemini:
		return llm.GetGeminiImageContent(imageContent)
	case param.OpenAi:
		return llm.GetOpenAIImageContent(imageContent)
	}
	
	return "", nil
}

func (r *RobotInfo) GetLastImageContent() ([]byte, error) {
	_, _, userID := r.GetChatIdAndMsgIdAndUserID()
	imageInfo, err := db.GetLastImageRecord(userID)
	if err != nil {
		logger.Warn("get last image content fail", "err", err)
		return nil, err
	}
	if imageInfo == nil {
		return nil, nil
	}
	
	answer := imageInfo.Answer
	const base64Prefix = "data:image/"
	if strings.HasPrefix(answer, base64Prefix) {
		// 去掉前缀，找到 base64 数据起始位置
		idx := strings.Index(answer, "base64,")
		if idx == -1 {
			return nil, errors.New("invalid base64 image data URI")
		}
		base64Data := answer[idx+7:] // "base64," 长度是7
		imageContent, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			logger.Warn("decode base64 image fail", "err", err)
			return nil, err
		}
		return imageContent, nil
	}
	
	// 不是 base64 data URI，尝试下载文件
	imageContent, err := utils.DownloadFile(answer)
	if err != nil {
		logger.Warn("download image fail", "err", err)
	}
	return imageContent, err
}

func (r *RobotInfo) TalkingPreCheck(f func()) {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	if r.checkUserTokenExceed(chatId, msgId, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	defer utils.DecreaseUserChat(userId)
	
	// check user chat exceed max count
	if utils.CheckUserChatExceed(userId) {
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	f()
}

func (r *RobotInfo) handleModeUpdate(mode string) {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user fail", "userID", userId, "err", err)
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(userId, mode)
		if err != nil {
			logger.Warn("update user fail", "userID", userId, "err", err)
			r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	} else {
		_, err = db.InsertUser(userId, mode)
		if err != nil {
			logger.Warn("insert user fail", "userID", userId, "err", err)
			r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	
	totalContent := i18n.GetMessage(*conf.BaseConfInfo.Lang, "mode_choose", nil) + mode
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

// ParseCommand extracts command and arguments like /photo xxx
func ParseCommand(prompt string) (command string, args string) {
	if len(prompt) == 0 || prompt[0] != '/' {
		return "", prompt
	}
	parts := strings.SplitN(prompt, " ", 2)
	command = parts[0]
	if len(parts) > 1 {
		args = parts[1]
	}
	return command, args
}
