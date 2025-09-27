package robot

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work/message/request"
	"github.com/bwmarrin/discordgo"
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/slack-go/slack"
	"github.com/tencent-connect/botgo/dto"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/rag"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

const (
	AudioMsgLen = 600
)

type MsgChan struct {
	NormalMessageChan chan *param.MsgInfo
	StrMessageChan    chan string
}

type RobotController struct {
	Cancel context.CancelFunc
}

type RobotInfo struct {
	Robot        Robot
	TencentRobot TencentRobot
	
	Token    int
	RecordID int64
}

var (
	robotController = new(RobotController)
	TencentMsgMap   sync.Map
)

type TencentWechatMessage struct {
	Msg       string
	Status    int
	StartTime time.Time
}

type Robot interface {
	checkValid() bool
	
	getMsgContent() string
	
	requestLLMAndResp(content string)
	
	sendChatMessage()
	
	sendModeConfigurationOptions()
	
	sendImg()
	
	sendVideo()
	
	sendHelpConfigurationOptions()
	
	getPrompt() string
	
	getContent(content string) (string, error)
	
	getPerMsgLen() int
	
	sendVoiceContent(voiceContent []byte, duration int) error
	
	sendText(msgChan *MsgChan)
}

type TencentRobot interface {
	passiveExecCmd()
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
	
	if r.Robot.checkValid() {
		r.Robot.requestLLMAndResp(r.Robot.getMsgContent())
	}
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
		if slackRobot.Event != nil {
			chatId = slackRobot.Event.Channel
			userId = slackRobot.Event.User
			msgId = slackRobot.Event.ClientMsgID
		}
		if slackRobot.Callback != nil {
			chatId = slackRobot.Callback.Channel.ID
			userId = slackRobot.Callback.User.ID
			msgId = slackRobot.Callback.MessageTs
		}
		if slackRobot.CmdEvent != nil {
			chatId = slackRobot.CmdEvent.ChannelID
			userId = slackRobot.CmdEvent.UserID
		}
	case *LarkRobot:
		lark := r.Robot.(*LarkRobot)
		if lark.Message != nil {
			msgId = larkcore.StringValue(lark.Message.Event.Message.MessageId)
			chatId = larkcore.StringValue(lark.Message.Event.Message.ChatId)
			userId = larkcore.StringValue(lark.Message.Event.Sender.SenderId.UserId)
		}
	case *DingRobot:
		dingRobot := r.Robot.(*DingRobot)
		if dingRobot.Message != nil {
			chatId = dingRobot.Message.ConversationId
			msgId = dingRobot.Message.MsgId
			userId = dingRobot.Message.SenderId
		}
	case *ComWechatRobot:
		comWechatRobot := r.Robot.(*ComWechatRobot)
		if comWechatRobot.Event != nil {
			chatId = comWechatRobot.Event.GetFromUserName()
			userId = comWechatRobot.Event.GetFromUserName()
		}
		
		if comWechatRobot.TextMsg != nil {
			msgId = comWechatRobot.TextMsg.MsgID
		}
		if comWechatRobot.ImageMsg != nil {
			msgId = comWechatRobot.ImageMsg.MsgID
		}
		if comWechatRobot.VoiceMsg != nil {
			msgId = comWechatRobot.VoiceMsg.MsgID
		}
	
	case *QQRobot:
		q := r.Robot.(*QQRobot)
		if q.C2CMessage != nil {
			chatId = q.C2CMessage.Author.ID
			userId = q.C2CMessage.Author.ID
			msgId = q.C2CMessage.ID
		}
		if q.GroupAtMessage != nil {
			chatId = q.GroupAtMessage.GroupID
			userId = q.GroupAtMessage.Author.ID
			msgId = q.GroupAtMessage.ID
		}
		if q.ATMessage != nil {
			chatId = q.ATMessage.GuildID
			userId = q.ATMessage.Author.ID
			msgId = q.ATMessage.ID
		}
	case *WechatRobot:
		wechatRobot := r.Robot.(*WechatRobot)
		if wechatRobot.Event != nil {
			chatId = wechatRobot.Event.GetFromUserName()
			userId = wechatRobot.Event.GetFromUserName()
		}
		
		if wechatRobot.TextMsg != nil {
			msgId = wechatRobot.TextMsg.MsgID
		}
		if wechatRobot.ImageMsg != nil {
			msgId = wechatRobot.ImageMsg.MsgID
		}
		if wechatRobot.VoiceMsg != nil {
			msgId = wechatRobot.VoiceMsg.MsgID
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
			
			if replyToMessageID != "" {
				messageSend.Reference = &discordgo.MessageReference{
					MessageID: replyToMessageID,
					ChannelID: chatId,
				}
			}
			
			sentMsg, err := discordRobot.Session.ChannelMessageSendComplex(chatId, messageSend)
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
		
		if replyToMessageID != "" {
			resp, err := lark.Client.Im.Message.Reply(lark.Ctx, larkim.NewReplyMessageReqBuilder().
				MessageId(replyToMessageID).
				Body(larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					Content(GetMarkdownContent(msgContent)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.Warn("send message fail", "err", err, "resp", resp)
				return ""
			}
			
			return *resp.Data.MessageId
		} else {
			resp, err := lark.Client.Im.Message.Create(lark.Ctx, larkim.NewCreateMessageReqBuilder().
				ReceiveIdType(larkim.ReceiveIdTypeChatId).
				Body(larkim.NewCreateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					ReceiveId(chatId).
					Content(GetMarkdownContent(msgContent)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return *resp.Data.MessageId
		}
	
	case *DingRobot:
		d := r.Robot.(*DingRobot)
		_, err := d.SimpleReplyMarkdown(d.Ctx, []byte(">"+d.OriginPrompt+"\n\n"+msgContent))
		if err != nil {
			logger.Warn("send message fail", "err", err)
			return ""
		}
	case *ComWechatRobot:
		c := r.Robot.(*ComWechatRobot)
		_, err := c.App.Message.SendMarkdown(c.Ctx, &request.RequestMessageSendMarkdown{
			RequestMessageSend: request.RequestMessageSend{
				ToUser:                 chatId,
				MsgType:                "markdown",
				AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
				DuplicateCheckInterval: 1800,
			},
			Markdown: &request.RequestMarkdown{
				Content: ">" + c.OriginPrompt + "\n\n" + msgContent,
			},
		})
		if err != nil {
			logger.Warn("send message fail", "err", err)
		}
	case *QQRobot:
		q := r.Robot.(*QQRobot)
		qqMsg := &dto.MessageToCreate{
			MsgType: dto.TextMsg,
			Content: strings.ReplaceAll(strings.ReplaceAll(msgContent, "http", ""), "https", ""),
			MsgID:   replyToMessageID,
			MsgSeq:  crc32.ChecksumIEEE([]byte(msgContent)),
			MessageReference: &dto.MessageReference{
				MessageID:             replyToMessageID,
				IgnoreGetMessageError: true,
			},
		}
		if q.C2CMessage != nil {
			resp, err := q.QQApi.PostC2CMessage(q.Ctx, q.C2CMessage.Author.ID, qqMsg)
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
		
		if q.ATMessage != nil {
			resp, err := q.QQApi.PostMessage(q.Ctx, q.ATMessage.GuildID, qqMsg)
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
		
		if q.GroupAtMessage != nil {
			resp, err := q.QQApi.PostGroupMessage(q.Ctx, q.GroupAtMessage.GroupID, qqMsg)
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
	case *WechatRobot:
		w := r.Robot.(*WechatRobot)
		if *conf.BaseConfInfo.WechatActive {
			resp, err := w.App.CustomerService.Message(w.Ctx, messages.NewText(msgContent)).
				SetTo(w.Event.GetFromUserName()).From(w.Event.GetToUserName()).Send(w.Ctx)
			if err != nil {
				logger.Error("send image fail", "err", err, "resp", resp)
				return ""
			}
		} else {
			_, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
			_, ok := TencentMsgMap.Load(msgId)
			if msgId != "" && !ok {
				msgContent = strings.ReplaceAll(strings.ReplaceAll(msgContent, "http", ""), "https", "")
				TencentMsgMap.Store(msgId, &TencentWechatMessage{
					Msg:       msgContent,
					Status:    msgFinished,
					StartTime: time.Now(),
				})
			}
		}
		
	}
	
	return ""
}

func WithRobot(robot Robot) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.Robot = robot
	}
}

func WithTencentRobot(robot TencentRobot) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.TencentRobot = robot
	}
}

func StartRobot() {
	ctx, cancel := context.WithCancel(context.Background())
	robotController.Cancel = cancel
	
	if *conf.BaseConfInfo.TelegramBotToken != "" {
		go func() {
			StartTelegramRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.DiscordBotToken != "" {
		go func() {
			StartDiscordRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.LarkAPPID != "" && *conf.BaseConfInfo.LarkAppSecret != "" {
		go func() {
			StartLarkRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.SlackBotToken != "" && *conf.BaseConfInfo.SlackAppToken != "" {
		go func() {
			StartSlackRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.DingClientId != "" && *conf.BaseConfInfo.DingClientSecret != "" {
		go func() {
			StartDingRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.ComWechatSecret != "" && *conf.BaseConfInfo.ComWechatAgentID != "" && *conf.BaseConfInfo.ComWechatEncodingAESKey != "" {
		go func() {
			StartComWechatRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.QQAppID != "" && *conf.BaseConfInfo.QQAppSecret != "" {
		go func() {
			StartQQRobot(ctx)
		}()
	}
	
	if *conf.BaseConfInfo.WechatAppID != "" && *conf.BaseConfInfo.WechatAppSecret != "" {
		go func() {
			StartWechatRobot()
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
	var answer string
	var err error
	var token int
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		answer, err = utils.FileRecognize(audioContent)
	case param.OpenAi:
		answer, err = llm.GenerateOpenAIText(audioContent)
	case param.Gemini:
		answer, token, err = llm.GenerateGeminiText(audioContent)
	}
	
	if err != nil {
		return "", err
	}
	
	err = db.AddRecordToken(r.RecordID, token)
	if err != nil {
		logger.Warn("addRecordToken err", "err", err)
	}
	
	return answer, err
}

func (r *RobotInfo) GetImageContent(imageContent []byte, content string) (string, error) {
	var answer string
	var err error
	var token int
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		answer, token, err = llm.GetVolImageContent(imageContent, content)
	case param.Gemini:
		answer, token, err = llm.GetGeminiImageContent(imageContent, content)
	case param.OpenAi, param.Aliyun:
		answer, token, err = llm.GetOpenAIImageContent(imageContent, content)
	case param.AI302, param.OpenRouter:
		answer, token, err = llm.GetMixImageContent(imageContent, content)
	}
	
	if err != nil {
		return "", err
	}
	
	err = db.AddRecordToken(r.RecordID, token)
	if err != nil {
		logger.Warn("addRecordToken err", "err", err)
	}
	
	if content == "" {
		return answer, nil
	}
	
	param := map[string]interface{}{
		"question": content,
		"answer":   answer,
	}
	prompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_export_prompt", param)
	return prompt, nil
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
	prompt = strings.TrimSpace(prompt)
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

func (r *RobotInfo) ExecCmd(cmd string, defaultFunc func()) {
	switch cmd {
	case "balance", "/balance":
		r.showBalanceInfo()
	case "state", "/state":
		r.showStateInfo()
	case "clear", "/clear":
		r.clearAllRecord()
	case "retry", "/retry":
		r.retryLastQuestion()
	case "chat", "/chat":
		r.Robot.sendChatMessage()
	case "mode", "/mode":
		r.Robot.sendModeConfigurationOptions()
	case "photo", "/photo", "edit_photo", "/edit_photo":
		r.Robot.sendImg()
	case "video", "/video":
		r.Robot.sendVideo()
	case "help", "/help":
		r.Robot.sendHelpConfigurationOptions()
	case "change_photo", "/change_photo", "rec_photo", "/rec_photo", "save_voice", "/save_voice":
		if r.TencentRobot != nil {
			r.TencentRobot.passiveExecCmd()
		} else {
			defaultFunc()
		}
	case "task", "/task":
		var emptyPromptFunc func()
		if t, ok := r.Robot.(*TelegramRobot); ok {
			emptyPromptFunc = t.sendForceReply("task_empty_content")
		}
		r.sendMultiAgent("task_empty_content", emptyPromptFunc)
	case "mcp", "/mcp":
		var emptyPromptFunc func()
		if t, ok := r.Robot.(*TelegramRobot); ok {
			emptyPromptFunc = t.sendForceReply("mcp_empty_content")
		}
		r.sendMultiAgent("mcp_empty_content", emptyPromptFunc)
	default:
		defaultFunc()
	}
}

func (r *RobotInfo) ExecChain(msgContent string, msgChan *MsgChan) {
	r.TalkingPreCheck(func() {
		chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
		
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		r.InsertRecord()
		
		content, err := r.Robot.getContent(strings.TrimSpace(msgContent))
		if err != nil {
			logger.Error("get content fail", "err", err)
			r.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		
		perMsgLen := r.Robot.getPerMsgLen()
		if *conf.AudioConfInfo.TTSType != "" {
			perMsgLen = AudioMsgLen
		}
		
		dpLLM := rag.NewRag(
			llm.WithMessageChan(msgChan.NormalMessageChan),
			llm.WithHTTPMsgChan(msgChan.StrMessageChan),
			llm.WithContent(content),
			llm.WithChatId(chatId),
			llm.WithUserId(userId),
			llm.WithPerMsgLen(perMsgLen),
			llm.WithRecordId(r.RecordID),
		)
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err = chains.Run(ctx, qaChain, content)
		if err != nil {
			r.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
	})
}

func (r *RobotInfo) ExecLLM(msgContent string, msgChan *MsgChan) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
		}
		if msgChan.NormalMessageChan != nil {
			close(msgChan.NormalMessageChan)
		}
		
		if msgChan.StrMessageChan != nil {
			close(msgChan.StrMessageChan)
		}
	}()
	
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	r.InsertRecord()
	content, err := r.Robot.getContent(strings.TrimSpace(msgContent))
	if err != nil {
		logger.Error("get content fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	perMsgLen := r.Robot.getPerMsgLen()
	if *conf.AudioConfInfo.TTSType != "" {
		perMsgLen = AudioMsgLen
	}
	
	llmClient := llm.NewLLM(
		llm.WithChatId(chatId),
		llm.WithUserId(userId),
		llm.WithMsgId(msgId),
		llm.WithMessageChan(msgChan.NormalMessageChan),
		llm.WithHTTPMsgChan(msgChan.StrMessageChan),
		llm.WithContent(content),
		llm.WithPerMsgLen(perMsgLen),
		llm.WithRecordId(r.RecordID),
	)
	
	err = llmClient.CallLLM()
	if err != nil {
		logger.Error("get content fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
	}
	
}

func (r *RobotInfo) showBalanceInfo() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	balance := llm.GetBalanceInfo()
	
	// handle balance info msg
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_title", nil), balance.IsAvailable)
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_content", nil)
	
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}
	
	r.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
}

func (r *RobotInfo) showStateInfo() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		userInfo, err = db.GetUserByID(userId)
	}
	
	// get today token
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	todayTokey, err := db.GetTokenByUserIdAndTime(userId, startOfDay.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get today token fail", "err", err)
	}
	
	// get this week token
	startOf7DaysAgo := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	weekToken, err := db.GetTokenByUserIdAndTime(userId, startOf7DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}
	
	// handle balance info msg
	startOf30DaysAgo := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	monthToken, err := db.GetTokenByUserIdAndTime(userId, startOf30DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	r.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
}

func (r *RobotInfo) clearAllRecord() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	deleteSuccMsg := i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil)
	r.SendMsg(chatId, deleteSuccMsg,
		msgId, tgbotapi.ModeMarkdown, nil)
	return
	
}

func (r *RobotInfo) retryLastQuestion() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		r.Robot.requestLLMAndResp(records.AQs[len(records.AQs)-1].Question)
	} else {
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
	
	return
	
}

func (r *RobotInfo) sendMultiAgent(agentType string, emptyPromptFunc func()) {
	r.TalkingPreCheck(func() {
		chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
		
		prompt := r.Robot.getPrompt()
		prompt = strings.TrimSpace(prompt)
		if len(prompt) == 0 {
			if emptyPromptFunc != nil {
				emptyPromptFunc()
			} else {
				logger.Warn("prompt is empty")
				r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			}
			return
		}
		
		dpReq := &llm.LLMTaskReq{
			Content:   prompt,
			UserId:    userId,
			ChatId:    chatId,
			MsgId:     msgId,
			PerMsgLen: r.Robot.getPerMsgLen(),
		}
		
		if _, ok := r.Robot.(*QQRobot); ok {
			dpReq.HTTPMsgChan = make(chan string)
		} else {
			dpReq.MessageChan = make(chan *param.MsgInfo)
		}
		
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("multi agent panic", "err", err, "stack", string(debug.Stack()))
				}
				if dpReq.HTTPMsgChan != nil {
					close(dpReq.HTTPMsgChan)
				}
				if dpReq.MessageChan != nil {
					close(dpReq.MessageChan)
				}
			}()
			
			var err error
			if agentType == "mcp_empty_content" {
				err = dpReq.ExecuteMcp()
			} else {
				err = dpReq.ExecuteTask()
			}
			if err != nil {
				logger.Warn("execute task fail", "err", err)
				r.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}()
		
		go r.HandleUpdate(&MsgChan{
			NormalMessageChan: dpReq.MessageChan,
			StrMessageChan:    dpReq.HTTPMsgChan,
		}, "")
	})
}

func (r *RobotInfo) CreatePhoto(prompt string, lastImageContent []byte) ([]byte, int, error) {
	
	var imageUrl string
	var imageContent []byte
	var totalToken int
	var err error
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		imageUrl, totalToken, err = llm.GenerateVolImg(prompt, lastImageContent)
	case param.OpenAi:
		imageContent, totalToken, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
	case param.Gemini:
		imageContent, totalToken, err = llm.GenerateGeminiImg(prompt, lastImageContent)
	case param.AI302, param.OpenRouter:
		imageUrl, totalToken, err = llm.GenerateMixImg(prompt, lastImageContent)
	default:
		err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
	}
	
	if err != nil {
		logger.Warn("generate image fail", "err", err)
		return nil, 0, err
	}
	
	if len(imageContent) == 0 {
		imageContent, err = utils.DownloadFile(imageUrl)
		if err != nil {
			logger.Warn("download image fail", "err", err)
			return nil, 0, err
		}
	}
	
	return imageContent, totalToken, nil
}

func (r *RobotInfo) CreateVideo(prompt string, lastImageContent []byte) ([]byte, int, error) {
	var videoUrl string
	var videoContent []byte
	var err error
	var totalToken int
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		videoUrl, totalToken, err = llm.GenerateVolVideo(prompt, lastImageContent)
	case param.Gemini:
		videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt, lastImageContent)
	case param.AI302:
		videoUrl, totalToken, err = llm.Generate302AIVideo(prompt, lastImageContent)
	default:
		err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
	}
	if err != nil {
		logger.Warn("generate video fail", "err", err)
		return nil, 0, err
	}
	
	if len(videoContent) == 0 {
		videoContent, err = utils.DownloadFile(videoUrl)
		if err != nil {
			logger.Warn("download video fail", "err", err)
			return nil, 0, err
		}
	}
	
	return videoContent, totalToken, nil
}

func (r *RobotInfo) GetVoiceBaseTTS(content, encoding string) ([]byte, int, error) {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	var ttsContent []byte
	var err error
	var duration int
	var token int
	switch *conf.AudioConfInfo.TTSType {
	case param.Vol:
		ttsContent, token, duration, err = llm.VolTTS(content, userId, encoding)
	case param.Gemini:
		ttsContent, token, duration, err = llm.GeminiTTS(content, encoding)
	case param.OpenAi:
		ttsContent, token, duration, err = llm.OpenAITTS(content, encoding)
	}
	
	err = db.AddRecordToken(r.RecordID, token)
	if err != nil {
		logger.Warn("addRecordToken err", "err", err)
	}
	
	return ttsContent, duration, err
}

func (r *RobotInfo) sendVoice(messageChan *MsgChan, encoding string) {
	chatId, messageId, _ := r.GetChatIdAndMsgIdAndUserID()
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			voiceContent, duration, err := r.GetVoiceBaseTTS(msg.Content, encoding)
			if err != nil {
				logger.Error("tts fail", "err", err)
				r.SendMsg(chatId, msg.Content, messageId, "", nil)
				continue
			}
			err = r.Robot.sendVoiceContent(voiceContent, duration)
			if err != nil {
				logger.Error("sendVoice fail", "err", err)
				r.SendMsg(chatId, msg.Content, messageId, "", nil)
				continue
			}
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		msg = new(param.MsgInfo)
		return
	}
	
	voiceContent, duration, err := r.GetVoiceBaseTTS(msg.Content, encoding)
	if err != nil {
		logger.Error("tts fail", "err", err)
		r.SendMsg(chatId, err.Error(), messageId, "", nil)
		return
	}
	err = r.Robot.sendVoiceContent(voiceContent, duration)
	if err != nil {
		logger.Error("sendVoice fail", "err", err)
		r.SendMsg(chatId, err.Error(), messageId, "", nil)
		return
	}
}

func (r *RobotInfo) HandleUpdate(messageChan *MsgChan, encoding string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	if *conf.AudioConfInfo.TTSType != "" && encoding != "" {
		r.sendVoice(messageChan, encoding)
	} else {
		r.Robot.sendText(messageChan)
	}
	
}

func (r *RobotInfo) InsertRecord() {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	
	id, err := db.InsertRecordInfo(&db.Record{
		UserId:     userId,
		Question:   r.Robot.getPrompt(),
		RecordType: param.TextRecordType,
	})
	if err != nil {
		logger.Error("insert record fail", "err", err)
		return
	}
	
	r.RecordID = id
}
