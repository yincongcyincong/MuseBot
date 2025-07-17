package robot

import (
	"errors"
	"runtime/debug"
	"strconv"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type DiscordRobot struct {
	Session *discordgo.Session
	Msg     *discordgo.MessageCreate
	Inter   *discordgo.InteractionCreate
	
	Robot *Robot
}

func StartDiscordRobot() {
	dg, err := discordgo.New("Bot " + *conf.BaseConfInfo.DiscordBotToken)
	if err != nil {
		logger.Fatal("create discord bot", "err", err)
	}
	
	// 添加消息处理函数
	dg.AddHandler(messageCreate)
	dg.AddHandler(onInteractionCreate)
	
	// 打开连接
	err = dg.Open()
	if err != nil {
		logger.Fatal("connect fail", "err", err)
	}
	
	logger.Info("discordBot Info", dg.State.User.Username)
	
	registerSlashCommands(dg)
}

func NewDiscordRobot(s *discordgo.Session, msg *discordgo.MessageCreate, i *discordgo.InteractionCreate) *DiscordRobot {
	return &DiscordRobot{
		Session: s,
		Msg:     msg,
		Inter:   i,
	}
}

func (d *DiscordRobot) Execute() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	//if d.handleCommandAndCallback() {
	//	return
	//}
	// check whether you have new message
	if d.Msg != nil {
		if d.skipThisMsg() {
			logger.Warn("skip this msg", "msgId", msgId, "chat", chatId, "content", d.Msg.Content)
			return
		}
		d.requestDeepseekAndResp(d.Msg.Content)
	}
}

func (d *DiscordRobot) requestDeepseekAndResp(content string) {
	chatId, replyToMessageID, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	if d.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	if conf.RagConfInfo.Store != nil {
		//d.executeChain(content)
	} else {
		d.executeLLM(content)
	}
}

func (d *DiscordRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	
	// request DeepSeek API
	go d.callLLM(content, messageChan)
	
	// send response message
	go d.handleUpdate(messageChan)
}

func (d *DiscordRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdateDiscord panic", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var originalMsgID string
	var channelID string
	var err error
	
	// 获取当前通道ID
	if d.Msg != nil {
		channelID = d.Msg.ChannelID
		
		// 1. 发送一个“thinking...”占位消息
		thinkingMsg, err := d.Session.ChannelMessageSend(channelID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil))
		if err != nil {
			logger.Warn("Sending thinking message failed", "err", err)
		} else {
			originalMsgID = thinkingMsg.ID
		}
		
	} else if d.Inter != nil {
		channelID = d.Inter.ChannelID
		
		// 1. 响应占位符（deferred response）
		err = d.Session.InteractionRespond(d.Inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			logger.Warn("Failed to defer interaction response", "err", err)
		}
	} else {
		logger.Error("Unknown Discord message type")
		return
	}
	
	var msg *param.MsgInfo
	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from deepseek!"
		}
		
		if msg.MsgId == 0 && originalMsgID != "" {
			msg.MsgId = utils.ParseInt(originalMsgID)
		}
		
		if d.Msg != nil {
			// 普通消息：编辑占位，或发送新消息
			if msg.MsgId == 0 {
				_, err = d.Session.ChannelMessageSend(channelID, msg.Content)
				if err != nil {
					logger.Warn("Sending message failed", "err", err)
				}
			} else {
				_, err = d.Session.ChannelMessageEdit(channelID, strconv.Itoa(msg.MsgId), msg.Content)
				if err != nil {
					logger.Warn("Editing message failed", "msgID", msg.MsgId, "err", err)
				}
				originalMsgID = "" // 避免后续再编辑
			}
		} else if d.Inter != nil {
			// 如果是 Interaction，使用 follow-up message 或 edit original
			if msg.MsgId == 0 {
				// 编辑原始响应
				_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
					Content: &msg.Content,
				})
				if err != nil {
					logger.Warn("Editing interaction response failed", "err", err)
				}
			} else {
				// 发送新的 follow-up 消息（如果支持的话）
				_, err = d.Session.FollowupMessageCreate(d.Inter.Interaction, true, &discordgo.WebhookParams{
					Content: msg.Content,
				})
				if err != nil {
					logger.Warn("Sending followup interaction message failed", "err", err)
				}
			}
		}
	}
}

func (d *DiscordRobot) callLLM(content string, messageChan chan *param.MsgInfo) {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
		}
		utils.DecreaseUserChat(userId)
		close(messageChan)
	}()
	// check user chat exceed max count
	if utils.CheckUserChatExceed(userId) {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	text, err := d.getContent(content)
	if err != nil {
		logger.Error("get content fail", "err", err)
		d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	l := llm.NewLLM(llm.WithMessageChan(messageChan), llm.WithContent(text),
		llm.WithChatId(chatId), llm.WithMsgId(msgId),
		llm.WithUserId(userId),
		llm.WithTaskTools(&conf.AgentInfo{
			DeepseekTool:    conf.DeepseekTools,
			VolTool:         conf.VolTools,
			OpenAITools:     conf.OpenAITools,
			GeminiTools:     conf.GeminiTools,
			OpenRouterTools: conf.OpenRouterTools,
		}))
	
	err = l.CallLLM()
	if err != nil {
		logger.Error("get content fail", "err", err)
		d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
	}
}

func (d *DiscordRobot) getContent(defaultText string) (string, error) {
	var content string
	var attachments []*discordgo.MessageAttachment
	
	if d.Msg != nil {
		content = strings.TrimSpace(d.Msg.Content)
		attachments = d.Msg.Attachments
	} else if d.Inter != nil {
		if d.Inter.Type == discordgo.InteractionApplicationCommand {
			if len(d.Inter.ApplicationCommandData().Options) > 0 {
				content = strings.TrimSpace(d.Inter.ApplicationCommandData().Options[0].StringValue())
			}
		}
	}
	
	// 优先使用传入默认文本（外部可指定）
	if content == "" {
		content = strings.TrimSpace(defaultText)
	}
	
	// 如果没有文本，尝试从语音附件中获取
	if content == "" && len(attachments) > 0 && *conf.AudioConfInfo.AudioAppID != "" {
		for _, att := range attachments {
			if strings.HasPrefix(att.ContentType, "audio/") || strings.HasSuffix(att.Filename, ".ogg") || strings.HasSuffix(att.Filename, ".mp3") {
				audioContent, err := utils.DownloadFile(att.URL)
				if audioContent == nil || err != nil {
					logger.Warn("audio url empty", "url", att.URL, "err", err)
					return "", errors.New("audio url empty")
				}
				content = utils.FileRecognize(audioContent)
				break
			}
		}
	}
	
	// 如果仍然没有内容，尝试从图片附件中获取内容
	if content == "" && len(attachments) > 0 {
		for _, att := range attachments {
			if strings.HasPrefix(att.ContentType, "image/") {
				image, err := utils.DownloadFile(att.URL)
				if image == nil || err != nil {
					logger.Warn("image url empty", "url", att.URL, "err", err)
					return "", errors.New("image url empty")
				}
				imageContent, err := utils.GetImageContent(image)
				if err != nil {
					logger.Warn("get image content err", "err", err)
					return "", err
				}
				content = imageContent
				break
			}
		}
	}
	
	if content == "" {
		logger.Warn("content empty")
		return "", errors.New("content empty")
	}
	
	// 去除 @bot 提及
	if d.Session != nil && d.Session.State != nil && d.Session.State.User != nil {
		content = strings.ReplaceAll(content, "<@"+d.Session.State.User.ID+">", "")
	}
	
	return content, nil
}

func (d *DiscordRobot) skipThisMsg() bool {
	// 忽略自己发的消息
	if d.Msg.Author.ID == d.Session.State.User.ID {
		return true
	}
	
	// 私聊频道
	if d.Msg.GuildID == "" {
		// 如果内容为空，且没有附件（比如语音、图片）
		if strings.TrimSpace(d.Msg.Content) == "" && len(d.Msg.Attachments) == 0 {
			return true
		}
		return false
	}
	
	// 公共频道（Guild Channel）
	mentionedBot := false
	for _, user := range d.Msg.Mentions {
		if user.ID == d.Session.State.User.ID {
			mentionedBot = true
			break
		}
	}
	
	// 没有@机器人
	if !mentionedBot {
		return true
	}
	
	// 如果内容是@机器人以外的空内容
	contentWithoutMention := strings.TrimSpace(strings.ReplaceAll(d.Msg.Content, "<@"+d.Session.State.User.ID+">", ""))
	if contentWithoutMention == "" && len(d.Msg.Attachments) == 0 {
		return true
	}
	
	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d := NewDiscordRobot(s, m, nil)
	d.Robot = NewRobot(WithDiscordRobot(d))
	d.Robot.Exec()
}

func registerSlashCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{Name: "chat", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.chat.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: "mode", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.mode.description", nil)},
		{Name: "balance", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.balance.description", nil)},
		{Name: "state", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.state.description", nil)},
		{Name: "clear", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.clear.description", nil)},
		{Name: "retry", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.retry.description", nil)},
		{Name: "photo", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.photo.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: "video", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.video.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: "help", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.help.description", nil)},
		{Name: "task", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.task.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: "mcp", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.mcp.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		
		//{Name: "addtoken", Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.addtoken.description", nil)},
	}
	
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			logger.Error("Cannot create command", "cmd", cmd.Name, "err", err)
		}
	}
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	
	d := NewDiscordRobot(s, nil, i)
	d.Robot = NewRobot(WithDiscordRobot(d))
	d.Robot.Exec()
	
	cmd := i.ApplicationCommandData().Name
	switch cmd {
	case "chat":
		prompt := i.ApplicationCommandData().Options[0].StringValue()
		d.sendChatMessage(prompt)
	case "mode":
		sendModeOptions(s, i)
	case "balance":
		showBalanceInfo(s, i)
	case "state":
		showStateInfo(s, i)
	case "clear":
		d.clearAllRecord()
	case "retry":
		retryLastQuestion(s, i)
	case "photo":
		sendImage(s, i)
	case "video":
		sendVideo(s, i)
	case "help":
		sendHelp(s, i)
	case "task":
		sendMultiAgent(s, i, "task_empty_content")
	case "mcp":
		sendMultiAgent(s, i, "mcp_empty_content")
	case "addtoken":
		//if adminUserIDs[i.Member.User.ID] {
		//	addToken(s, i)
		//}
	}
}

func (d *DiscordRobot) sendChatMessage(prompt string) {
	d.requestDeepseekAndResp(prompt)
}
func sendModeOptions(s *discordgo.Session, i *discordgo.InteractionCreate) {}
func showBalanceInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {}
func showStateInfo(s *discordgo.Session, i *discordgo.InteractionCreate)   {}
func (d *DiscordRobot) clearAllRecord() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}
func retryLastQuestion(s *discordgo.Session, i *discordgo.InteractionCreate)          {}
func sendImage(s *discordgo.Session, i *discordgo.InteractionCreate)                  {}
func sendVideo(s *discordgo.Session, i *discordgo.InteractionCreate)                  {}
func sendHelp(s *discordgo.Session, i *discordgo.InteractionCreate)                   {}
func sendMultiAgent(s *discordgo.Session, i *discordgo.InteractionCreate, tag string) {}
func addToken(s *discordgo.Session, i *discordgo.InteractionCreate)                   {}

func sendChatMessageFromReply(s *discordgo.Session, m *discordgo.MessageCreate)            {}
func sendImageFromReply(s *discordgo.Session, m *discordgo.MessageCreate)                  {}
func sendVideoFromReply(s *discordgo.Session, m *discordgo.MessageCreate)                  {}
func sendMultiAgentFromReply(s *discordgo.Session, m *discordgo.MessageCreate, tag string) {}
