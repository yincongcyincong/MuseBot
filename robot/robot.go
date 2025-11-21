package robot

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work/message/request"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	cron "github.com/robfig/cron/v3"
	"github.com/sashabaranov/go-openai"
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

var (
	smartModeReg = regexp.MustCompile(`\{\s*(?:"cron"\s*:\s*".*?"\s*,\s*"prompt"\s*:\s*".*?"\s*,\s*)?"command"\s*:\s*".*?"\s*\}`)
	Cron         *cron.Cron
)

type MsgChan struct {
	NormalMessageChan chan *param.MsgInfo
	StrMessageChan    chan string
}

type RobotController struct {
	Cancel context.CancelFunc
}

type RobotInfo struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Robot  Robot
	
	cs *param.ContextState
}

var (
	RobotControl  = new(RobotController)
	TencentMsgMap sync.Map
)

type TencentWechatMessage struct {
	Msg       string
	Status    int
	StartTime time.Time
}

type Robot interface {
	checkValid() bool
	
	getMsgContent() string
	
	requestLLM(content string)
	
	sendChatMessage()
	
	sendImg()
	
	sendVideo()
	
	getPrompt() string
	
	setPrompt(prompt string)
	
	getPerMsgLen() int
	
	sendVoiceContent(voiceContent []byte, duration int) error
	
	setCommand(command string)
	
	getCommand() string
	
	getUserName() string
	
	executeLLM()
	
	getImage() []byte
	
	setImage(image []byte)
	
	sendMedia(media []byte, contentType, sType string) error
}

type botOption func(r *RobotInfo)

func NewRobot(options ...botOption) *RobotInfo {
	r := new(RobotInfo)
	r.cs = new(param.ContextState)
	r.cs.UseRecord = true
	for _, o := range options {
		o(r)
	}
	
	if r.Ctx == nil {
		r.Ctx = context.Background()
	}
	
	ctx, cancel := context.WithTimeout(r.Ctx, 15*time.Minute)
	
	if ctx.Value("bot_name") == nil {
		ctx = context.WithValue(ctx, "bot_name", *conf.BaseConfInfo.BotName)
	}
	if ctx.Value("log_id") == nil {
		ctx = context.WithValue(ctx, "log_id", uuid.New().String())
	}
	
	r.Ctx = ctx
	r.Cancel = cancel
	return r
}

func (r *RobotInfo) Exec() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	if !r.cs.SkipCheck && !r.checkUserAllow(userId) && !r.checkGroupAllow(chatId) {
		logger.WarnCtx(r.Ctx, "user/group not allow to use this bot", "userID", userId, "chat", chatId)
		r.SendMsg(chatId, i18n.GetMessage("valid_user_group", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	if r.AddUserInfo() && r.Robot.checkValid() && r.smartMode() {
		r.Robot.requestLLM(r.Robot.getMsgContent())
	}
}

func (r *RobotInfo) AddUserInfo() bool {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "addUserInfo GetUserByID err", "err", err)
		return false
	}
	
	if userInfo == nil || userInfo.ID == 0 {
		_, err = db.InsertUser(userId, utils.GetDefaultLLMConfig())
		if err != nil {
			logger.ErrorCtx(r.Ctx, "insert user fail", "userID", userId, "err", err)
			return false
		}
		
		userInfo, err = db.GetUserByID(userId)
		if err != nil || userInfo == nil {
			logger.ErrorCtx(r.Ctx, "addUserInfo GetUserByID err", "err", err)
			return false
		}
	}
	
	if userInfo.LLMConfigRaw == nil {
		userInfo.LLMConfigRaw = new(param.LLMConfig)
	}
	
	r.Ctx = context.WithValue(r.Ctx, "user_info", userInfo)
	return true
	
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
			msgId = slackRobot.Event.EventTimeStamp
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
	case *PersonalQQRobot:
		personalQQRobot := r.Robot.(*PersonalQQRobot)
		chatId = personalQQRobot.Msg.Raw.PeerUid
		userId = strconv.Itoa(int(personalQQRobot.Msg.UserID))
		msgId = strconv.Itoa(int(personalQQRobot.Msg.MessageID))
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
			logger.WarnCtx(r.Ctx, "send clear message fail", "err", err)
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
				logger.WarnCtx(r.Ctx, "send discord message fail", "err", err)
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
				logger.WarnCtx(r.Ctx, "send discord interaction response fail", "err", err)
			}
			return ""
		}
	
	case *SlackRobot:
		slackRobot := r.Robot.(*SlackRobot)
		_, timestamp, err := slackRobot.Client.PostMessage(chatId, slack.MsgOptionText(msgContent, false))
		if err != nil {
			logger.WarnCtx(r.Ctx, "send message fail", "err", err)
		}
		
		return timestamp
	case *LarkRobot:
		lark := r.Robot.(*LarkRobot)
		
		if replyToMessageID != "" {
			resp, err := lark.Client.Im.Message.Reply(r.Ctx, larkim.NewReplyMessageReqBuilder().
				MessageId(replyToMessageID).
				Body(larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					Content(GetMarkdownContent(msgContent)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.WarnCtx(r.Ctx, "send message fail", "err", err, "resp", resp)
				return ""
			}
			
			return *resp.Data.MessageId
		} else {
			resp, err := lark.Client.Im.Message.Create(r.Ctx, larkim.NewCreateMessageReqBuilder().
				ReceiveIdType(larkim.ReceiveIdTypeChatId).
				Body(larkim.NewCreateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					ReceiveId(chatId).
					Content(GetMarkdownContent(msgContent)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.ErrorCtx(r.Ctx, "send message fail", "err", err, "resp", resp)
				return ""
			}
			
			return *resp.Data.MessageId
		}
	
	case *DingRobot:
		d := r.Robot.(*DingRobot)
		_, err := d.SimpleReplyMarkdown(r.Ctx, []byte(">"+d.OriginPrompt+"\n\n"+msgContent))
		if err != nil {
			logger.WarnCtx(r.Ctx, "send message fail", "err", err)
			return ""
		}
	case *ComWechatRobot:
		c := r.Robot.(*ComWechatRobot)
		_, err := c.App.Message.SendMarkdown(r.Ctx, &request.RequestMessageSendMarkdown{
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
			logger.WarnCtx(r.Ctx, "send message fail", "err", err)
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
			resp, err := q.QQApi.PostC2CMessage(q.Robot.Ctx, q.C2CMessage.Author.ID, qqMsg)
			if err != nil {
				logger.WarnCtx(r.Ctx, "send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
		
		if q.ATMessage != nil {
			resp, err := q.QQApi.PostMessage(r.Ctx, q.ATMessage.GuildID, qqMsg)
			if err != nil {
				logger.WarnCtx(r.Ctx, "send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
		
		if q.GroupAtMessage != nil {
			resp, err := q.QQApi.PostGroupMessage(r.Ctx, q.GroupAtMessage.GroupID, qqMsg)
			if err != nil {
				logger.WarnCtx(r.Ctx, "send message fail", "err", err)
				return ""
			}
			
			return resp.ID
		}
	case *WechatRobot:
		w := r.Robot.(*WechatRobot)
		if *conf.BaseConfInfo.WechatActive {
			resp, err := w.App.CustomerService.Message(r.Ctx, messages.NewText(msgContent)).
				SetTo(w.Event.GetFromUserName()).From(w.Event.GetToUserName()).Send(r.Ctx)
			if err != nil {
				logger.ErrorCtx(r.Ctx, "send image fail", "err", err, "resp", resp)
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
	
	case *PersonalQQRobot:
		personalQQRobot := r.Robot.(*PersonalQQRobot)
		msgId, err := personalQQRobot.SendMsg(msgContent, nil, nil, nil)
		if err != nil {
			logger.ErrorCtx(r.Ctx, "send msg fail", "err", err)
			return ""
		}
		return msgId
	}
	
	return ""
}

func WithRobot(robot Robot) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.Robot = robot
	}
}

func WithContext(ctx context.Context) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.Ctx = ctx
	}
}

func WithSkipCheck(skipCheck bool) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.cs.SkipCheck = skipCheck
	}
}

func WithUseRecord(userRecord bool) func(*RobotInfo) {
	return func(r *RobotInfo) {
		r.cs.UseRecord = userRecord
	}
}

func StartRobot() {
	ctx, cancel := context.WithCancel(context.Background())
	RobotControl.Cancel = cancel
	ctx = context.WithValue(ctx, "bot_name", *conf.BaseConfInfo.BotName)
	
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
	if *conf.BaseConfInfo.TokenPerUser <= 0 {
		return false
	}
	
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.WarnCtx(r.Ctx, "get user info fail", "err", err)
		return false
	}
	
	if userInfo.Token >= userInfo.AvailToken {
		tpl := i18n.GetMessage("token_exceed", nil)
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
	var token = param.AudioTokenUsage
	llmConf := db.GetCtxUserInfo(r.Ctx).LLMConfigRaw
	t := utils.GetRecType(llmConf)
	logger.InfoCtx(r.Ctx, "recognize audio", "type", t, "model", utils.GetUsingRecModel(t, llmConf.RecModel))
	switch t {
	case param.Vol:
		answer, err = utils.FileRecognize(audioContent)
	case param.OpenAi:
		answer, err = llm.GenerateOpenAIText(r.Ctx, audioContent)
	case param.Gemini:
		answer, token, err = llm.GenerateGeminiText(r.Ctx, audioContent)
	case param.Aliyun:
		answer, token, err = llm.GenerateAliyunText(r.Ctx, audioContent)
	}
	
	if err != nil {
		return "", err
	}
	
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	err = db.AddRecordToken(r.Ctx, r.cs.RecordID, userId, token)
	if err != nil {
		logger.WarnCtx(r.Ctx, "addRecordToken err", "err", err)
	}
	err = db.AddRecordContent(r.cs.RecordID, fmt.Sprintf("data:audio/%s;base64,%s", utils.DetectAudioFormat(audioContent), base64.StdEncoding.EncodeToString(audioContent)))
	if err != nil {
		logger.WarnCtx(r.Ctx, "AddRecordContent err", "err", err)
	}
	
	return answer, err
}

func (r *RobotInfo) GetLastImageContent() ([]byte, error) {
	_, _, userID := r.GetChatIdAndMsgIdAndUserID()
	imageInfo, err := db.GetLastImageRecord(userID)
	if err != nil {
		logger.WarnCtx(r.Ctx, "get last image content fail", "err", err)
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
			logger.WarnCtx(r.Ctx, "decode base64 image fail", "err", err)
			return nil, err
		}
		return imageContent, nil
	}
	
	imageContent, err := utils.DownloadFile(answer)
	if err != nil {
		logger.WarnCtx(r.Ctx, "download image fail", "err", err)
	}
	return imageContent, err
}

func (r *RobotInfo) TalkingPreCheck(f func()) {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	if r.checkUserTokenExceed(chatId, msgId, userId) {
		logger.WarnCtx(r.Ctx, "user token exceed", "userID", userId)
		return
	}
	
	defer utils.DecreaseUserChat(userId)
	
	// check user chat exceed max count
	if utils.CheckUserChatExceed(userId) {
		r.SendMsg(chatId, i18n.GetMessage("chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	f()
}

type RobotModel struct {
	TxtType    string
	ImgType    string
	VideoType  string
	TxtModel   string
	ImgModel   string
	VideoModel string
	RecType    string
	RecModel   string
	TTSType    string
	TTSModel   string
}

func (r *RobotInfo) handleModelUpdate(rm *RobotModel) {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	userInfo := db.GetCtxUserInfo(r.Ctx)
	if userInfo != nil && userInfo.ID != 0 {
		llmConf := userInfo.LLMConfigRaw
		if llmConf == nil {
			llmConf = &param.LLMConfig{}
		}
		if rm.TxtType != "" {
			llmConf.TxtType = rm.TxtType
		}
		if rm.ImgType != "" {
			llmConf.ImgType = rm.ImgType
		}
		if rm.VideoType != "" {
			llmConf.VideoType = rm.VideoType
		}
		if rm.TxtModel != "" {
			llmConf.TxtModel = rm.TxtModel
		}
		if rm.ImgModel != "" {
			llmConf.ImgModel = rm.ImgModel
		}
		if rm.VideoModel != "" {
			llmConf.VideoModel = rm.VideoModel
		}
		if rm.RecType != "" {
			llmConf.RecType = rm.RecType
		}
		if rm.RecModel != "" {
			llmConf.RecModel = rm.RecModel
		}
		if rm.TTSType != "" {
			llmConf.TTSType = rm.TTSType
		}
		if rm.TTSModel != "" {
			llmConf.TTSModel = rm.TTSModel
		}
		
		mode, _ := json.Marshal(llmConf)
		
		err := db.UpdateUserLLMConfig(userId, string(mode))
		if err != nil {
			logger.WarnCtx(r.Ctx, "update user fail", "userID", userId, "err", err)
			r.SendMsg(chatId, i18n.GetMessage("set_mode", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	
	totalContent := i18n.GetMessage("mode_choose", nil) + r.Robot.getPrompt()
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

// ParseCommand extracts command and arguments like /photo xxx
func ParseCommand(prompt string) (command string, args string) {
	prompt = strings.TrimSpace(prompt)
	if len(prompt) == 0 || (prompt[0] != '/' && prompt[0] != '$') {
		return "", prompt
	}
	parts := strings.SplitN(prompt, " ", 2)
	command = parts[0]
	if len(parts) > 1 {
		args = parts[1]
	}
	return command, args
}

func (r *RobotInfo) ExecCmd(cmd string, defaultFunc func(), modeFunc func(string), typesFunc func(string)) {
	_, _, userID := r.GetChatIdAndMsgIdAndUserID()
	logger.InfoCtx(r.Ctx, "command info", "userID", userID, "cmd", cmd)
	
	switch cmd {
	case param.State, "/" + param.State, "$" + param.State:
		r.showStateInfo()
	case param.Clear, "/" + param.Clear, "$" + param.Clear:
		r.clearAllRecord()
	case param.Retry, "/" + param.Retry, "$" + param.Retry:
		r.retryLastQuestion()
	case param.Chat, "/" + param.Chat, "$" + param.Chat:
		r.Robot.sendChatMessage()
	case param.TxtType, "/" + param.TxtType, "$" + param.TxtType, param.PhotoType, "/" + param.PhotoType, "$" + param.PhotoType, param.VideoType,
		"/" + param.VideoType, "$" + param.VideoType, param.RecType, "/" + param.RecType, "$" + param.RecType, param.TtsType, "/" + param.TtsType, "$" + param.TtsType:
		if typesFunc != nil {
			typesFunc(cmd)
		} else {
			r.changeType(cmd)
		}
	case param.TxtModel, "/" + param.TxtModel, "$" + param.TxtModel, param.PhotoModel, "/" + param.PhotoModel, "$" + param.PhotoModel,
		param.VideoModel, "/" + param.VideoModel, "$" + param.VideoModel, param.RecModel, "/" + param.RecModel, "$" + param.RecModel,
		param.TtsModel, "/" + param.TtsModel, "$" + param.TtsModel:
		if modeFunc != nil {
			modeFunc(cmd)
		} else {
			r.changeModel(cmd)
		}
	case param.Photo, "/" + param.Photo, "$" + param.Photo, param.EditPhoto, "/" + param.EditPhoto, "$" + param.EditPhoto:
		r.Robot.sendImg()
	case param.Video, "/" + param.Video, "$" + param.Video:
		r.Robot.sendVideo()
	case param.Help, "/" + param.Help, "$" + param.Help:
		r.sendHelpInfo()
	case param.RecPhoto, "/" + param.RecPhoto, "$" + param.RecPhoto:
		r.recPhoto()
	case param.Task, "/" + param.Task, "$" + param.Task:
		var emptyPromptFunc func()
		if t, ok := r.Robot.(*TelegramRobot); ok {
			emptyPromptFunc = t.sendForceReply("task_empty_content")
		}
		r.sendMultiAgent("task_empty_content", emptyPromptFunc)
	case param.Mcp, "/" + param.Mcp, "$" + param.Mcp:
		var emptyPromptFunc func()
		if t, ok := r.Robot.(*TelegramRobot); ok {
			emptyPromptFunc = t.sendForceReply("mcp_empty_content")
		}
		r.sendMultiAgent("mcp_empty_content", emptyPromptFunc)
	case param.Mode, "/" + param.Mode, "$" + param.Mode:
		r.showMode()
	case param.CronList, "/" + param.CronList, "$" + param.CronList:
		r.cronList()
	case param.CronDel, "/" + param.CronDel, "$" + param.CronDel:
		r.cronDel()
	case param.CronClear, "/" + param.CronClear, "$" + param.CronClear:
		r.cronClear()
	default:
		defaultFunc()
	}
}

func (r *RobotInfo) cronList() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	crons, err := db.GetCronsByPage(1, 10, "", userId)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "get crons failed", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	txt := i18n.GetMessage("cron_list_header", nil)
	for _, c := range crons {
		txt += i18n.GetMessage("cron_list_item", map[string]interface{}{
			"id":        c.ID,
			"cron_spec": c.CronSpec,
			"prompt":    c.Prompt,
			"target":    c.TargetID,
			"group_id":  c.GroupID,
		})
	}
	r.SendMsg(chatId, txt, msgId, "", nil)
	
}

func (r *RobotInfo) cronDel() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	err := db.DeleteCronByCreateBy(userId, r.Robot.getPrompt())
	if err != nil {
		logger.ErrorCtx(r.Ctx, "delete cron failed", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	r.SendMsg(chatId, i18n.GetMessage("del_cron_success", nil), msgId, "", nil)
}

func (r *RobotInfo) cronClear() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	err := db.DeleteCronByCreateBy(userId, "")
	if err != nil {
		logger.ErrorCtx(r.Ctx, "clear cron failed", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	r.SendMsg(chatId, i18n.GetMessage("clear_cron_success", nil), msgId, "", nil)
}

func (r *RobotInfo) showMode() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	llmConf := db.GetCtxUserInfo(r.Ctx).LLMConfigRaw
	if llmConf == nil {
		llmConf = new(param.LLMConfig)
	}
	txtType := utils.GetTxtType(llmConf)
	photoType := utils.GetImgType(llmConf)
	videoType := utils.GetVideoType(llmConf)
	recType := utils.GetRecType(llmConf)
	ttsType := utils.GetTTSType(llmConf)
	r.SendMsg(chatId, i18n.GetMessage("mode_info", map[string]interface{}{
		"txt_type":    txtType,
		"photo_type":  photoType,
		"video_type":  videoType,
		"rec_type":    recType,
		"tts_type":    ttsType,
		"txt_model":   utils.GetUsingTxtModel(txtType, llmConf.TxtModel),
		"photo_model": utils.GetUsingImgModel(photoType, llmConf.ImgModel),
		"video_model": utils.GetUsingVideoModel(videoType, llmConf.VideoModel),
		"rec_model":   utils.GetUsingRecModel(recType, llmConf.RecModel),
		"tts_model":   utils.GetUsingTTSModel(ttsType, llmConf.TTSModel),
	}), msgId, "", nil)
}

func (r *RobotInfo) changeType(t string) {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	totalContent := ""
	switch t {
	case "txt_type", "/txt_type", "$txt_type":
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{TxtType: r.Robot.getPrompt()})
			return
		}
		for _, model := range utils.GetAvailTxtType() {
			totalContent += fmt.Sprintf(`%s

`, model)
		}
	
	case "photo_type", "/photo_type", "$photo_type":
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{ImgType: r.Robot.getPrompt()})
			return
		}
		for _, model := range utils.GetAvailImgType() {
			totalContent += fmt.Sprintf(`%s

`, model)
		}
	
	case "video_type", "/video_type", "$video_type":
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{VideoType: r.Robot.getPrompt()})
			return
		}
		
		for _, model := range utils.GetAvailVideoType() {
			totalContent += fmt.Sprintf(`%s

`, model)
		}
	case "rec_type", "/rec_type", "$rec_type":
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{RecType: r.Robot.getPrompt()})
			return
		}
		
		for _, model := range utils.GetAvailRecType() {
			totalContent += fmt.Sprintf(`%s

`, model)
		}
	case "tts_type", "/tts_type", "$tts_type":
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{TTSType: r.Robot.getPrompt()})
			return
		}
		
		for _, model := range utils.GetAvailRecType() {
			totalContent += fmt.Sprintf(`%s

`, model)
		}
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
	
}

func (r *RobotInfo) changeModel(ty string) {
	switch ty {
	case param.TxtModel, "/" + param.TxtModel, "$" + param.TxtModel:
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{TxtModel: r.Robot.getPrompt()})
			return
		}
		r.showTxtModel()
	case param.PhotoModel, "/" + param.PhotoModel, "$" + param.PhotoModel:
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{ImgModel: r.Robot.getPrompt()})
			return
		}
		r.showImageModel()
	case param.VideoModel, "/" + param.VideoModel, "$" + param.VideoModel:
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{VideoModel: r.Robot.getPrompt()})
			return
		}
		r.showVideoModel()
	case param.RecModel, "/" + param.RecModel, "$" + param.RecModel:
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{RecModel: r.Robot.getPrompt()})
			return
		}
		r.showRecModel()
	case param.TtsModel, "/" + param.TtsModel, "$" + param.TtsModel:
		if r.Robot.getPrompt() != "" {
			r.handleModelUpdate(&RobotModel{TTSModel: r.Robot.getPrompt()})
			return
		}
		r.showTTSModel()
	}
	
}

func (r *RobotInfo) showTxtModel() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	var modelList []string
	
	switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
	case param.DeepSeek:
		for k := range param.DeepseekModels {
			modelList = append(modelList, k)
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			modelList = append(modelList, k)
		}
	case param.Aliyun:
		for k := range param.AliyunModel {
			modelList = append(modelList, k)
		}
	case param.OpenRouter, param.AI302, param.Ollama, param.OpenAi:
		switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
		case param.OpenAi:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.AI302:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://ollama.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		}
		
		return
	case param.Vol:
		for k := range param.VolModels {
			modelList = append(modelList, k)
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (r *RobotInfo) showImageModel() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	var modelList []string
	
	switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiImageModels {
			modelList = append(modelList, k)
		}
	case param.Aliyun:
		for k := range param.AliyunImageModels {
			modelList = append(modelList, k)
		}
	case param.OpenRouter, param.AI302, param.Ollama, param.OpenAi:
		switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
		case param.OpenAi:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.AI302:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://ollama.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		}
		return
	case param.Vol:
		for k := range param.VolImageModels {
			modelList = append(modelList, k)
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (r *RobotInfo) showVideoModel() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	var modelList []string
	
	switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiVideoModels {
			modelList = append(modelList, k)
		}
	case param.Aliyun:
		for k := range param.AliyunVideoModels {
			modelList = append(modelList, k)
		}
	case param.AI302:
		switch utils.GetTxtType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
		case param.AI302:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		case param.Vol:
			for k := range param.VolVideoModels {
				modelList = append(modelList, k)
			}
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (r *RobotInfo) showRecModel() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	var modelList []string
	
	switch utils.GetRecType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiRecModels {
			modelList = append(modelList, k)
		}
	case param.Aliyun:
		for k := range param.AliyunRecModels {
			modelList = append(modelList, k)
		}
	case param.Vol:
		for k := range param.VolRecModels {
			modelList = append(modelList, k)
		}
	case param.AI302, param.OpenAi:
		switch utils.GetRecType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
		case param.AI302:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		case param.OpenAi:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (r *RobotInfo) showTTSModel() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	var modelList []string
	
	switch utils.GetTTSType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiTTSModels {
			modelList = append(modelList, k)
		}
	case param.Aliyun:
		for k := range param.AliyunTTSModels {
			modelList = append(modelList, k)
		}
	case param.Vol:
		for k := range param.VolTTSModels {
			modelList = append(modelList, k)
		}
	case param.OpenAi:
		switch utils.GetTTSType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw) {
		case param.OpenAi:
			r.SendMsg(chatId, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	r.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (r *RobotInfo) ExecChain(msgContent string, msgChan *MsgChan) {
	r.TalkingPreCheck(func() {
		chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
		
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(r.Ctx, "panic", "err", err, "stack", string(debug.Stack()))
			}
			if msgChan.NormalMessageChan != nil {
				close(msgChan.NormalMessageChan)
			}
			
			if msgChan.StrMessageChan != nil {
				close(msgChan.StrMessageChan)
			}
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		content := r.Robot.getPrompt()
		if len(msgContent) == 0 {
			logger.InfoCtx(r.Ctx, "content is empty")
			return
		}
		
		r.InsertRecord()
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
			llm.WithCS(r.cs),
		)
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err := chains.Run(ctx, qaChain, content)
		if err != nil {
			r.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
	})
}

func (r *RobotInfo) ExecLLM(msgContent string, msgChan *MsgChan) {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(r.Ctx, "GetContent panic err", "err", err, "stack", string(debug.Stack()))
		}
		if msgChan.NormalMessageChan != nil {
			close(msgChan.NormalMessageChan)
		}
		
		if msgChan.StrMessageChan != nil {
			close(msgChan.StrMessageChan)
		}
	}()
	
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	content := r.Robot.getPrompt()
	if len(content) == 0 {
		if len(r.Robot.getImage()) > 0 {
			r.saveRecord(r.Robot.getImage(), r.Robot.getImage(), param.ImageRecordType, 0)
		}
		logger.InfoCtx(r.Ctx, "content is empty")
		return
	}
	
	r.InsertRecord()
	perMsgLen := r.Robot.getPerMsgLen()
	if *conf.AudioConfInfo.TTSType != "" {
		perMsgLen = AudioMsgLen
	}
	
	images := make([][]byte, 0)
	if len(r.Robot.getImage()) > 0 {
		images = append(images, r.Robot.getImage())
	}
	
	llmClient := llm.NewLLM(
		llm.WithChatId(chatId),
		llm.WithUserId(userId),
		llm.WithMsgId(msgId),
		llm.WithMessageChan(msgChan.NormalMessageChan),
		llm.WithHTTPMsgChan(msgChan.StrMessageChan),
		llm.WithContent(content),
		llm.WithPerMsgLen(perMsgLen),
		llm.WithCS(r.cs),
		llm.WithContext(r.Ctx),
		llm.WithContentParameter(map[string]string{
			"username":  r.Robot.getUserName(),
			"image_day": strconv.Itoa(conf.BaseConfInfo.ImageDay),
		}),
		llm.WithTaskTools(&conf.AgentInfo{
			DeepseekTool: conf.DeepseekTools,
			VolTool:      conf.VolTools,
			OpenAITools:  conf.OpenAITools,
			GeminiTools:  conf.GeminiTools,
		}),
		llm.WithImages(images),
	)
	
	err := llmClient.CallLLM()
	if err != nil {
		logger.ErrorCtx(r.Ctx, "get content fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, "", nil)
	}
	
}

func (r *RobotInfo) showStateInfo() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.WarnCtx(r.Ctx, "get user info fail", "err", err)
		r.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	// get today token
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	todayTokey, err := db.GetTokenByUserIdAndTime(userId, startOfDay.Unix(), endOfDay.Unix())
	if err != nil {
		logger.WarnCtx(r.Ctx, "get today token fail", "err", err)
	}
	
	// get this week token
	startOf7DaysAgo := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	weekToken, err := db.GetTokenByUserIdAndTime(userId, startOf7DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.WarnCtx(r.Ctx, "get week token fail", "err", err)
	}
	
	startOf30DaysAgo := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	monthToken, err := db.GetTokenByUserIdAndTime(userId, startOf30DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.WarnCtx(r.Ctx, "get week token fail", "err", err)
	}
	
	template := i18n.GetMessage("state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	r.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
}

func (r *RobotInfo) clearAllRecord() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(r.Ctx, userId)
	deleteSuccMsg := i18n.GetMessage("delete_succ", nil)
	r.SendMsg(chatId, deleteSuccMsg,
		msgId, tgbotapi.ModeMarkdown, nil)
	return
	
}

func (r *RobotInfo) retryLastQuestion() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		r.Robot.requestLLM(records.AQs[len(records.AQs)-1].Question)
	} else {
		r.SendMsg(chatId, i18n.GetMessage("last_question_fail", nil),
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
				logger.WarnCtx(r.Ctx, "prompt is empty")
				r.SendMsg(chatId, i18n.GetMessage("photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			}
			return
		}
		
		dpReq := &llm.LLMTaskReq{
			Content:   prompt,
			UserId:    userId,
			ChatId:    chatId,
			MsgId:     msgId,
			PerMsgLen: r.Robot.getPerMsgLen(),
			Ctx:       r.Ctx,
		}
		
		if _, ok := r.Robot.(*QQRobot); ok {
			dpReq.HTTPMsgChan = make(chan string)
		} else {
			dpReq.MessageChan = make(chan *param.MsgInfo)
		}
		
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorCtx(r.Ctx, "multi agent panic", "err", err, "stack", string(debug.Stack()))
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
				logger.WarnCtx(r.Ctx, "execute task fail", "err", err)
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
	llmConf := db.GetCtxUserInfo(r.Ctx).LLMConfigRaw
	mediaType := utils.GetImgType(llmConf)
	logger.InfoCtx(r.Ctx, "create image", "mediaType", mediaType, "mediaModel",
		utils.GetUsingRecModel(mediaType, llmConf.ImgModel), "lastImageContent", len(lastImageContent), "prompt", prompt)
	switch mediaType {
	case param.Vol:
		imageUrl, totalToken, err = llm.GenerateVolImg(r.Ctx, prompt, lastImageContent)
	case param.OpenAi, param.ChatAnyWhere:
		imageContent, totalToken, err = llm.GenerateOpenAIImg(r.Ctx, prompt, lastImageContent)
	case param.Gemini:
		imageContent, totalToken, err = llm.GenerateGeminiImg(r.Ctx, prompt, lastImageContent)
	case param.AI302, param.OpenRouter:
		imageContent, totalToken, err = llm.GenerateMixImg(r.Ctx, prompt, lastImageContent)
	case param.Aliyun:
		imageUrl, totalToken, err = llm.GenerateAliyunImg(r.Ctx, prompt, lastImageContent)
	default:
		err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
	}
	
	if err != nil {
		logger.ErrorCtx(r.Ctx, "generate image fail", "err", err)
		return nil, 0, err
	}
	
	if len(imageContent) == 0 {
		imageContent, err = utils.DownloadFile(imageUrl)
		if err != nil {
			logger.WarnCtx(r.Ctx, "download image fail", "err", err)
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
	llmConf := db.GetCtxUserInfo(r.Ctx).LLMConfigRaw
	mediaType := utils.GetVideoType(llmConf)
	logger.InfoCtx(r.Ctx, "create video", "mediaType", mediaType, "mediaModel",
		utils.GetUsingVideoModel(mediaType, llmConf.VideoModel), "lastImageContent", len(lastImageContent), "prompt", prompt)
	switch mediaType {
	case param.Vol:
		videoUrl, totalToken, err = llm.GenerateVolVideo(r.Ctx, prompt, lastImageContent)
	case param.Gemini:
		videoContent, totalToken, err = llm.GenerateGeminiVideo(r.Ctx, prompt, lastImageContent)
	case param.AI302:
		videoUrl, totalToken, err = llm.Generate302AIVideo(r.Ctx, prompt, lastImageContent)
	case param.Aliyun:
		videoUrl, totalToken, err = llm.GenerateAliyunVideo(r.Ctx, prompt, lastImageContent)
	default:
		err = fmt.Errorf("unsupported type: %s", mediaType)
	}
	if err != nil {
		logger.WarnCtx(r.Ctx, "generate video fail", "err", err)
		return nil, 0, err
	}
	
	if len(videoContent) == 0 {
		videoContent, err = utils.DownloadFile(videoUrl)
		if err != nil {
			logger.WarnCtx(r.Ctx, "download video fail", "err", err)
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
		ttsContent, token, duration, err = llm.VolTTS(r.Ctx, content, userId, encoding)
	case param.Gemini:
		ttsContent, token, duration, err = llm.GeminiTTS(r.Ctx, content, encoding)
	case param.OpenAi:
		ttsContent, token, duration, err = llm.OpenAITTS(r.Ctx, content, encoding)
	case param.Aliyun:
		ttsContent, token, duration, err = llm.AliyunTTS(r.Ctx, content, encoding)
	}
	
	err = db.AddRecordToken(r.Ctx, r.cs.RecordID, userId, token)
	if err != nil {
		logger.WarnCtx(r.Ctx, "addRecordToken err", "err", err)
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
				logger.ErrorCtx(r.Ctx, "tts fail", "err", err)
				r.SendMsg(chatId, msg.Content, messageId, "", nil)
				continue
			}
			err = r.Robot.sendVoiceContent(voiceContent, duration)
			if err != nil {
				logger.ErrorCtx(r.Ctx, "sendVoice fail", "err", err)
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
		logger.ErrorCtx(r.Ctx, "tts fail", "err", err)
		r.SendMsg(chatId, err.Error(), messageId, "", nil)
		return
	}
	err = r.Robot.sendVoiceContent(voiceContent, duration)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "sendVoice fail", "err", err)
		r.SendMsg(chatId, err.Error(), messageId, "", nil)
		return
	}
}

func (r *RobotInfo) HandleUpdate(messageChan *MsgChan, encoding string) {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(r.Ctx, "handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	if *conf.AudioConfInfo.TTSType != "" && encoding != "" {
		r.sendVoice(messageChan, encoding)
	} else {
		r.sendText(messageChan)
	}
	
}

func (r *RobotInfo) InsertRecord() {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	
	content := ""
	if len(r.Robot.getImage()) > 0 {
		content = fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(r.Robot.getImage()),
			base64.StdEncoding.EncodeToString(r.Robot.getImage()))
	}
	
	id, err := db.InsertRecordInfo(r.Ctx, &db.Record{
		UserId:     userId,
		Question:   r.Robot.getCommand() + " " + r.Robot.getPrompt(),
		RecordType: param.TextRecordType,
		Token:      r.cs.Token,
		Content:    content,
	})
	if err != nil {
		logger.ErrorCtx(r.Ctx, "insert record fail", "err", err)
		return
	}
	
	r.cs.RecordID = id
}

func (r *RobotInfo) sendHelpInfo() {
	chatId, msgId, _ := r.GetChatIdAndMsgIdAndUserID()
	helpTxt := i18n.GetMessage("help_text", nil)
	if _, ok := r.Robot.(*TelegramRobot); !ok {
		helpTxt = strings.ReplaceAll(helpTxt, "\\_", "_")
	}
	r.SendMsg(chatId, helpTxt,
		msgId, tgbotapi.ModeMarkdown, nil)
}

type SmartModeResult struct {
	Command string `json:"command"`
	Prompt  string `json:"prompt"`
	Cron    string `json:"cron"`
}

func (r *RobotInfo) smartMode() bool {
	if r.Robot.getCommand() != "" || r.Robot.getPrompt() == "" || !*conf.BaseConfInfo.SmartMode {
		return true
	}
	
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	llmClient := llm.NewLLM(
		llm.WithChatId(chatId),
		llm.WithUserId(userId),
		llm.WithMsgId(msgId),
		llm.WithContent(i18n.GetMessage("smart_mode_prompt", map[string]interface{}{
			"prompt": r.Robot.getPrompt(),
		})),
		llm.WithContext(r.Ctx),
		llm.WithImages([][]byte{
			r.Robot.getImage(),
		}),
	)
	llmClient.LLMClient.GetModel(llmClient)
	llmClient.LLMClient.GetMessage(openai.ChatMessageRoleUser, llmClient.Content)
	content, err := llmClient.LLMClient.SyncSend(r.Ctx, llmClient)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "get content fail", "err", err)
		return true
	}
	
	matches := smartModeReg.FindAllString(content, -1)
	smartResult := new(SmartModeResult)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), smartResult)
		if err != nil {
			logger.ErrorCtx(r.Ctx, "json umarshal fail", "err", err)
		}
	}
	
	logger.InfoCtx(r.Ctx, "smart mode result", "result", smartResult)
	
	switch smartResult.Command {
	case "/cron":
		r.cs.Token = llmClient.Cs.Token
		err = r.InsertCron(smartResult.Cron, smartResult.Prompt)
		if err != nil {
			r.SendMsg(chatId, err.Error(), msgId, "", nil)
		} else {
			r.SendMsg(chatId, i18n.GetMessage("set_cron_success", nil), msgId, "", nil)
		}
		return false
	default:
		if smartResult.Command != "" {
			r.Robot.setCommand(smartResult.Command)
		}
	}
	
	r.cs.Token = llmClient.Cs.Token
	return true
}

func (r *RobotInfo) saveRecord(content, imageContent []byte, recordType, totalToken int) {
	_, _, userId := r.GetChatIdAndMsgIdAndUserID()
	
	base64Content := base64.StdEncoding.EncodeToString(content)
	format := utils.DetectImageFormat(content)
	dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
	
	originImageURI := ""
	if len(imageContent) > 0 {
		base64Content = base64.StdEncoding.EncodeToString(imageContent)
		format = utils.DetectImageFormat(imageContent)
		originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
	}
	
	mode := ""
	if recordType == param.ImageRecordType {
		mode = utils.GetImgType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw)
	} else {
		mode = utils.GetVideoType(db.GetCtxUserInfo(r.Ctx).LLMConfigRaw)
	}
	
	// save data record
	_, err := db.InsertRecordInfo(r.Ctx, &db.Record{
		UserId:     userId,
		Question:   r.Robot.getCommand() + " " + r.Robot.getPrompt(),
		Answer:     dataURI,
		Content:    originImageURI,
		Token:      totalToken + r.cs.Token,
		IsDeleted:  0,
		RecordType: recordType,
		Mode:       mode,
	})
	if err != nil {
		logger.ErrorCtx(r.Ctx, "insert record fail", "err", err)
	}
}

func (r *RobotInfo) InsertCron(cron, prompt string) error {
	var err error
	var id int64
	chatId, _, userId := r.GetChatIdAndMsgIdAndUserID()
	targetId, groupId, t := "", "", ""
	switch r.Robot.(type) {
	case *PersonalQQRobot:
		q := r.Robot.(*PersonalQQRobot)
		q.Robot.GetChatIdAndMsgIdAndUserID()
		if q.Msg.MessageType != "group" {
			targetId = userId
		} else {
			groupId = strconv.Itoa(int(q.Msg.GroupId))
		}
		t = param.PersonalQQ
	case *TelegramRobot:
		te := r.Robot.(*TelegramRobot)
		if te.Update.Message.Chat.Type == "private" {
			targetId = userId
		} else {
			groupId = strconv.Itoa(int(te.Update.Message.Chat.ID))
		}
		t = param.Telegram
	case *LarkRobot:
		targetId = chatId
		t = param.Lark
	case *SlackRobot:
		targetId = chatId
		t = param.Slack
	case *WechatRobot:
		targetId = chatId
		t = param.Wechat
	case *ComWechatRobot:
		targetId = chatId
		t = param.ComWechat
	}
	
	if t == "" {
		return errors.New("unsupported robot type")
	}
	
	id, err = db.InsertCron(fmt.Sprintf("%s cron task", r.Robot.getUserName()), cron, targetId,
		groupId, "", prompt, t, userId)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "insert cron fail", "err", err)
		return err
	}
	
	cronInfo, err := db.GetCronByID(id)
	if err != nil {
		logger.ErrorCtx(r.Ctx, "get cron by id fail", "err", err)
		return err
	}
	
	return AddCron(cronInfo)
}

func (r *RobotInfo) recPhoto() {
	if len(r.Robot.getImage()) == 0 {
		lastImageContent, err := r.GetLastImageContent()
		if err != nil {
			logger.ErrorCtx(r.Ctx, "get last image record fail", "err", err)
			return
		}
		r.Robot.setImage(lastImageContent)
	}
	
	r.Robot.executeLLM()
}

func (r *RobotInfo) SendMarkdownMsg(msg *param.MsgInfo) {
	chatId, _, _ := r.GetChatIdAndMsgIdAndUserID()
	blocks := utils.ExtractContentBlocks(msg.Content)
	for _, b := range blocks {
		switch b.Type {
		case "text":
			r.SendMsg(chatId, strings.TrimSpace(b.Content), "", tgbotapi.ModeMarkdown, nil)
		case "video", "image":
			content, err := utils.DownloadFile(b.Media.URL)
			if err != nil {
				logger.ErrorCtx(r.Ctx, "download file fail", "err", err)
				continue
			}
			contentType := ""
			if b.Type == "video" {
				contentType = utils.DetectVideoMimeType(content)
			} else {
				contentType = utils.DetectImageFormat(content)
			}
			
			err = r.Robot.sendMedia(content, contentType, b.Type)
			if err != nil {
				logger.ErrorCtx(r.Ctx, "send media fail", "err", err)
			}
		}
	}
}

func (r *RobotInfo) sendText(messageChan *MsgChan) {
	if messageChan.NormalMessageChan != nil {
		var msg *param.MsgInfo
		for msg = range messageChan.NormalMessageChan {
			if msg.Finished {
				r.SendMarkdownMsg(msg)
			}
		}
		
		if msg != nil {
			r.SendMarkdownMsg(msg)
		}
	}
}
