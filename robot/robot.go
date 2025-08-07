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
	"github.com/slack-go/slack"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
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

func (r *RobotInfo) GetChatIdAndMsgIdAndUserID() (int64, int, string) {
	chatId := int64(0)
	msgId := 0
	userId := ""
	
	switch r.Robot.(type) {
	case *TelegramRobot:
		telegramRobot := r.Robot.(*TelegramRobot)
		if telegramRobot.Update.Message != nil {
			chatId = telegramRobot.Update.Message.Chat.ID
			userId = strconv.FormatInt(telegramRobot.Update.Message.From.ID, 10)
			msgId = telegramRobot.Update.Message.MessageID
		}
		if telegramRobot.Update.CallbackQuery != nil {
			chatId = telegramRobot.Update.CallbackQuery.Message.Chat.ID
			userId = strconv.FormatInt(telegramRobot.Update.CallbackQuery.From.ID, 10)
			msgId = telegramRobot.Update.CallbackQuery.Message.MessageID
		}
	case *DiscordRobot:
		discordRobot := r.Robot.(*DiscordRobot)
		if discordRobot.Msg != nil {
			chatId, _ = strconv.ParseInt(discordRobot.Msg.ChannelID, 10, 64)
			userId = discordRobot.Msg.Author.ID
			msgId = utils.ParseInt(discordRobot.Msg.Message.ID)
		}
		if discordRobot.Inter != nil {
			chatId, _ = strconv.ParseInt(discordRobot.Inter.ChannelID, 10, 64)
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
			chatId, _ = strconv.ParseInt(slackRobot.Event.Channel, 10, 64)
			userId = slackRobot.Event.User
			msgId, _ = strconv.Atoi(slackRobot.Event.ClientMsgID)
		}
	case *Web:
		web := r.Robot.(*Web)
		if web != nil {
			chatId = web.UserId
			msgId = int(web.UserId)
			userId = web.RealUserId
		}
	}
	
	return chatId, msgId, userId
}

func (r *RobotInfo) SendMsg(chatId int64, msgContent string, replyToMessageID int,
	mode string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) int {
	switch r.Robot.(type) {
	case *TelegramRobot:
		telegramRobot := r.Robot.(*TelegramRobot)
		msg := tgbotapi.NewMessage(chatId, msgContent)
		msg.ParseMode = mode
		msg.ReplyMarkup = inlineKeyboard
		msg.ReplyToMessageID = replyToMessageID
		msgInfo, err := telegramRobot.Bot.Send(msg)
		if err != nil {
			logger.Warn("send clear message fail", "err", err)
			return 0
		}
		return msgInfo.MessageID
	case *DiscordRobot:
		discordRobot := r.Robot.(*DiscordRobot)
		if discordRobot.Msg != nil {
			messageSend := &discordgo.MessageSend{
				Content: msgContent,
			}
			
			// 设置引用消息
			chatIdStr := fmt.Sprintf("%d", chatId)
			if replyToMessageID != 0 {
				messageSend.Reference = &discordgo.MessageReference{
					MessageID: strconv.Itoa(replyToMessageID),
					ChannelID: chatIdStr,
				}
			}
			
			sentMsg, err := discordRobot.Session.ChannelMessageSendComplex(chatIdStr, messageSend)
			if err != nil {
				logger.Warn("send discord message fail", "err", err)
				return 0
			}
			return utils.ParseInt(sentMsg.ID)
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
			return 0
		}
	
	case *SlackRobot:
		slackRobot := r.Robot.(*SlackRobot)
		_, timestamp, err := slackRobot.Client.PostMessage(strconv.FormatInt(chatId, 10), slack.MsgOptionText(msgContent, false))
		if err != nil {
			logger.Warn("send message fail", "err", err)
		}
		
		return utils.ParseInt(timestamp)
	case *Web:
		web := r.Robot.(*Web)
		_, err := web.W.Write([]byte(msgContent))
		if err != nil {
			logger.Warn("send message fail", "err", err)
		}
		web.Flusher.Flush()
	}
	
	return 0
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
}

// checkUserAllow check use can use telegram bot or not
func (r *RobotInfo) checkUserAllow(userId string) bool {
	if len(conf.BaseConfInfo.AllowedTelegramUserIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedTelegramUserIds["0"] {
		return false
	}
	
	_, ok := conf.BaseConfInfo.AllowedTelegramUserIds[userId]
	return ok
}

func (r *RobotInfo) checkGroupAllow(chatId int64) bool {
	
	if len(conf.BaseConfInfo.AllowedTelegramGroupIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedTelegramGroupIds[0] {
		return false
	}
	if _, ok := conf.BaseConfInfo.AllowedTelegramGroupIds[chatId]; ok {
		return true
	}
	
	return false
}

// checkUserTokenExceed check use token exceeded
func (r *RobotInfo) checkUserTokenExceed(chatId int64, msgId int, userId string) bool {
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
