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
	"time"
	
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

type MsgChan struct {
	NormalMessageChan chan *param.MsgInfo
	StrMessageChan    chan string
}

type RobotController struct {
	Cancel context.CancelFunc
}

type RobotInfo struct {
	Robot Robot
}

var (
	robotController = new(RobotController)
)

type Robot interface {
	checkValid() bool
	
	getMsgContent() string
	
	requestLLMAndResp(content string)
	
	sendChatMessage()
	
	sendModeConfigurationOptions()
	
	sendImg()
	
	sendVideo()
	
	sendHelpConfigurationOptions()
	
	handleUpdate(msgChan *MsgChan)
	
	getPrompt() string
	
	GetContent(content string) (string, error)
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
			msgId = ""
			userId = comWechatRobot.Event.GetFromUserName()
		}
	case *QQRobot:
		q := r.Robot.(*QQRobot)
		if q.C2CMessage != nil {
			chatId = q.C2CMessage.Author.ID
			userId = q.C2CMessage.Author.ID
			msgId = q.C2CMessage.ID
		}
		if q.ATMessage != nil {
			chatId = q.C2CMessage.GroupID
			userId = q.C2CMessage.Author.ID
			msgId = q.C2CMessage.ID
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
		if q.C2CMessage != nil {
			resp, err := q.QQApi.PostC2CMessage(q.Ctx, q.C2CMessage.Author.ID, &dto.MessageToCreate{
				MsgType: dto.TextMsg,
				Content: msgContent,
				MsgID:   replyToMessageID,
				MsgSeq:  crc32.ChecksumIEEE([]byte(msgContent)),
				MessageReference: &dto.MessageReference{
					MessageID:             replyToMessageID,
					IgnoreGetMessageError: true,
				},
			})
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
		
		if q.ATMessage != nil {
			resp, err := q.QQApi.PostGroupMessage(q.Ctx, q.ATMessage.GroupID, &dto.MessageToCreate{
				MsgType: dto.TextMsg,
				Content: msgContent,
				MsgID:   replyToMessageID,
				MsgSeq:  crc32.ChecksumIEEE([]byte(msgContent)),
				MessageReference: &dto.MessageReference{
					MessageID:             replyToMessageID,
					IgnoreGetMessageError: true,
				},
			})
			if err != nil {
				logger.Warn("send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
	}
	
	return ""
}

func WithRobot(robot Robot) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.Robot = robot
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
	
	if *conf.BaseConfInfo.ComWechatSecret != "" && *conf.BaseConfInfo.ComWechatAgentID != "" && *conf.BaseConfInfo.ComWechatCorpID != "" {
		go func() {
			StartComWechatRobot()
		}()
	}
	
	if *conf.BaseConfInfo.QQAppID != "" && *conf.BaseConfInfo.QQAppSecret != "" {
		go func() {
			StartQQRobot(ctx)
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
	case "photo", "/photo":
		r.Robot.sendImg()
	case "video", "/video":
		r.Robot.sendVideo()
	case "help", "/help":
		r.Robot.sendHelpConfigurationOptions()
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

func (r *RobotInfo) ExecChain(msgContent string, msgChan chan *param.MsgInfo, httpMsgChan chan string) {
	r.TalkingPreCheck(func() {
		chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
		
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		content, err := r.Robot.GetContent(strings.TrimSpace(msgContent))
		if err != nil {
			logger.Error("get content fail", "err", err)
			r.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		dpLLM := rag.NewRag(
			llm.WithMessageChan(msgChan),
			llm.WithHTTPMsgChan(httpMsgChan),
			llm.WithContent(content),
			llm.WithChatId(chatId),
			llm.WithUserId(userId),
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

func (r *RobotInfo) ExecLLM(msgContent string, msgChan chan *param.MsgInfo, httpMsgChan chan string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
		}
		if msgChan != nil {
			close(msgChan)
		}
		
		if httpMsgChan != nil {
			close(httpMsgChan)
		}
	}()
	
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	content, err := r.Robot.GetContent(strings.TrimSpace(msgContent))
	if err != nil {
		logger.Error("get content fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	llmClient := llm.NewLLM(
		llm.WithChatId(chatId),
		llm.WithUserId(userId),
		llm.WithMsgId(msgId),
		llm.WithMessageChan(msgChan),
		llm.WithHTTPMsgChan(httpMsgChan),
		llm.WithContent(content),
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
			Content: prompt,
			UserId:  userId,
			ChatId:  chatId,
			MsgId:   msgId,
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
		
		go r.Robot.handleUpdate(&MsgChan{
			NormalMessageChan: dpReq.MessageChan,
			StrMessageChan:    dpReq.HTTPMsgChan,
		})
	})
}

func StopAllRobot() {
	robotController.Cancel()
}
