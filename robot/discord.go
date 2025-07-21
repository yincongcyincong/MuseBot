package robot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	
	"github.com/bwmarrin/discordgo"
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/rag"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type DiscordRobot struct {
	Session *discordgo.Session
	Msg     *discordgo.MessageCreate
	Inter   *discordgo.InteractionCreate
	
	Robot *RobotInfo
}

func StartDiscordRobot() {
	dg, err := discordgo.New("Bot " + *conf.BaseConfInfo.DiscordBotToken)
	if err != nil {
		logger.Fatal("create discord bot", "err", err)
	}
	dg.Client = utils.GetTelegramProxyClient()
	
	// 添加消息处理函数
	dg.AddHandler(messageCreate)
	dg.AddHandler(onInteractionCreate)
	
	// 打开连接
	err = dg.Open()
	if err != nil {
		logger.Fatal("connect fail", "err", err)
	}
	
	logger.Info("discordBot Info", "username", dg.State.User.Username)
	
	registerSlashCommands(dg)
}

func NewDiscordRobot(s *discordgo.Session, msg *discordgo.MessageCreate, i *discordgo.InteractionCreate) *DiscordRobot {
	return &DiscordRobot{
		Session: s,
		Msg:     msg,
		Inter:   i,
	}
}

func (d *DiscordRobot) Exec() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	
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
		d.executeChain(content)
	} else {
		d.executeLLM(content)
	}
}

func (d *DiscordRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	go func() {
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
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		text, err := d.getContent(content)
		if err != nil {
			logger.Error("get content fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		
		dpLLM := rag.NewRag(llm.WithMessageChan(messageChan), llm.WithContent(content),
			llm.WithChatId(chatId), llm.WithMsgId(msgId),
			llm.WithUserId(userId))
		
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err = chains.Run(ctx, qaChain, text)
		if err != nil {
			logger.Warn("execute chain fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
	}()
	
	// send response message
	go d.handleUpdate(messageChan)
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
	
	if d.Msg != nil {
		channelID = d.Msg.ChannelID
		
		thinkingMsg, err := d.Session.ChannelMessageSend(channelID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil))
		if err != nil {
			logger.Warn("Sending thinking message failed", "err", err)
		} else {
			originalMsgID = thinkingMsg.ID
		}
		
	} else if d.Inter != nil {
		channelID = d.Inter.ChannelID
		
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
				originalMsgID = ""
			}
		} else if d.Inter != nil {
			if msg.MsgId == 0 {
				_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
					Content: &msg.Content,
				})
				if err != nil {
					logger.Warn("Editing interaction response failed", "err", err)
				}
			} else {
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
	
	if content == "" {
		content = strings.TrimSpace(defaultText)
	}
	
	if content == "" && len(attachments) > 0 && *conf.AudioConfInfo.AudioAppID != "" {
		for _, att := range attachments {
			if strings.HasPrefix(att.ContentType, "audio/") {
				audioContent, err := utils.DownloadFile(att.URL)
				if audioContent == nil || err != nil {
					logger.Warn("audio url empty", "url", att.URL, "err", err)
					return "", errors.New("audio url empty")
				}
				content, err = d.Robot.GetAudioContent(audioContent)
				if err != nil {
					logger.Warn("get audio content err", "err", err)
					return "", err
				}
				break
			}
		}
	}
	
	if content == "" && len(attachments) > 0 {
		for _, att := range attachments {
			if strings.HasPrefix(att.ContentType, "image/") {
				image, err := utils.DownloadFile(att.URL)
				if image == nil || err != nil {
					logger.Warn("image url empty", "url", att.URL, "err", err)
					return "", errors.New("image url empty")
				}
				content, err = d.Robot.GetImageContent(image)
				if err != nil {
					logger.Warn("get image content err", "err", err)
					return "", err
				}
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
	if d.Msg.Author.ID == d.Session.State.User.ID {
		return true
	}
	
	if d.Msg.GuildID == "" {
		if strings.TrimSpace(d.Msg.Content) == "" && len(d.Msg.Attachments) == 0 {
			return true
		}
		return false
	}
	
	mentionedBot := false
	for _, user := range d.Msg.Mentions {
		if user.ID == d.Session.State.User.ID {
			mentionedBot = true
			break
		}
	}
	
	if !mentionedBot {
		return true
	}
	
	contentWithoutMention := strings.TrimSpace(strings.ReplaceAll(d.Msg.Content, "<@"+d.Session.State.User.ID+">", ""))
	if contentWithoutMention == "" && len(d.Msg.Attachments) == 0 {
		return true
	}
	
	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d := NewDiscordRobot(s, m, nil)
	d.Robot = NewRobot(WithRobot(d))
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
		
		//	{Name: "addtoken", Description: "admin add token", Options: []*discordgo.ApplicationCommandOption{
		//		{Type: discordgo.ApplicationCommandOptionString, Name: "userId", Description: "user id", Required: true},
		//		{Type: discordgo.ApplicationCommandOptionString, Name: "token", Description: "token", Required: true},
		//	}},
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
	d.Robot = NewRobot(WithRobot(d))
	d.Robot.Exec()
	_, _, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	cmd := ""
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		cmd = i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		cmd = i.MessageComponentData().CustomID
	}
	
	switch cmd {
	case "chat":
		prompt := i.ApplicationCommandData().Options[0].StringValue()
		d.sendChatMessage(prompt)
	case "mode":
		d.sendModeOptions()
	case "balance":
		d.showBalanceInfo()
	case "state":
		d.showStateInfo()
	case "clear":
		d.clearAllRecord()
	case "retry":
		d.retryLastQuestion()
	case "photo":
		d.sendImage()
	case "video":
		d.sendVideo()
	case "help":
		d.sendHelp()
	case "task":
		d.sendMultiAgent("task_empty_content")
	case "mcp":
		d.sendMultiAgent("mcp_empty_content")
	case "addtoken":
		if d.Robot.checkAdminUser(userId) {
			d.addToken()
		}
	}
}

func (d *DiscordRobot) sendChatMessage(prompt string) {
	d.requestDeepseekAndResp(prompt)
}

func (d *DiscordRobot) sendModeOptions() {
	var buttons []discordgo.MessageComponent
	switch *conf.BaseConfInfo.Type {
	case param.DeepSeek:
		if *conf.BaseConfInfo.CustomUrl == "" || *conf.BaseConfInfo.CustomUrl == "https://api.deepseek.com/" {
			for k := range param.DeepseekModels {
				buttons = append(buttons, discordgo.Button{Label: k, Style: discordgo.PrimaryButton, CustomID: k})
			}
		} else {
			buttons = append(buttons,
				discordgo.Button{Label: godeepseek.AzureDeepSeekR1, CustomID: godeepseek.AzureDeepSeekR1, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1, CustomID: godeepseek.OpenRouterDeepSeekR1, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1DistillLlama70B, CustomID: godeepseek.OpenRouterDeepSeekR1DistillLlama70B, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1DistillLlama8B, CustomID: godeepseek.OpenRouterDeepSeekR1DistillLlama8B, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1DistillQwen14B, CustomID: godeepseek.OpenRouterDeepSeekR1DistillQwen14B, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B, CustomID: godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: godeepseek.OpenRouterDeepSeekR1DistillQwen32B, CustomID: godeepseek.OpenRouterDeepSeekR1DistillQwen32B, Style: discordgo.SecondaryButton},
				discordgo.Button{Label: "llama2", CustomID: param.LLAVA, Style: discordgo.SecondaryButton},
			)
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			buttons = append(buttons, discordgo.Button{Label: k, Style: discordgo.PrimaryButton, CustomID: k})
		}
	case param.OpenAi:
		for k := range param.OpenAIModels {
			buttons = append(buttons, discordgo.Button{Label: k, Style: discordgo.PrimaryButton, CustomID: k})
		}
	case param.LLAVA:
		buttons = append(buttons, discordgo.Button{Label: "llama2", Style: discordgo.PrimaryButton, CustomID: param.LLAVA})
	case param.OpenRouter:
		for k := range param.OpenRouterModelTypes {
			buttons = append(buttons, discordgo.Button{Label: k, Style: discordgo.PrimaryButton, CustomID: k})
		}
	case param.Vol:
		for k := range param.VolModels {
			buttons = append(buttons, discordgo.Button{Label: k, Style: discordgo.PrimaryButton, CustomID: k})
		}
	}
	
	// 每行最多 5 个按钮，进行分组
	var rows []discordgo.MessageComponent
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, discordgo.ActionsRow{Components: buttons[i:end]})
	}
	
	err := d.Session.InteractionRespond(d.Inter.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_mode", nil),
			Components: rows,
			Flags:      1 << 6,
		},
	})
	
	if err != nil {
		logger.Error("send message error", "err", err)
	}
}

func (d *DiscordRobot) showBalanceInfo() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil),
			msgId, "", nil)
		return
	}
	
	balance := llm.GetBalanceInfo()
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_title", nil), balance.IsAvailable)
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_content", nil)
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance, bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}
	
	d.Robot.SendMsg(chatId, msgContent, msgId, "", nil)
}

func (d *DiscordRobot) showStateInfo() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return
	}
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		userInfo, err = db.GetUserByID(userId)
	}
	
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	
	todayToken, _ := db.GetTokenByUserIdAndTime(userId, startOfDay.Unix(), endOfDay.Unix())
	weekToken, _ := db.GetTokenByUserIdAndTime(userId, now.AddDate(0, 0, -7).Unix(), endOfDay.Unix())
	monthToken, _ := db.GetTokenByUserIdAndTime(userId, now.AddDate(0, 0, -30).Unix(), endOfDay.Unix())
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayToken, weekToken, monthToken)
	
	d.Robot.SendMsg(chatId, msgContent, msgId, "", nil)
}

func (d *DiscordRobot) clearAllRecord() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}
func (d *DiscordRobot) retryLastQuestion() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		d.requestDeepseekAndResp(records.AQs[len(records.AQs)-1].Question)
	} else {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
}

func (d *DiscordRobot) sendImage() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	if utils.CheckUserChatExceed(userId) {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	defer utils.DecreaseUserChat(userId)
	
	if d.Robot.checkUserTokenExceed(chatId, msgId, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	prompt := d.Inter.ApplicationCommandData().Options[0].StringValue()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	lastImageContent, err := d.Robot.GetLastImageContent()
	if err != nil {
		logger.Warn("get last image record fail", "err", err)
	}
	
	msgThinking := d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
	
	var imageUrl string
	var imageContent []byte
	
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		imageUrl, err = llm.GenerateVolImg(prompt, lastImageContent)
	case param.OpenAi:
		imageUrl, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
	case param.Gemini:
		imageContent, err = llm.GenerateGeminiImg(prompt, lastImageContent)
	default:
		err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
	}
	
	if err != nil {
		logger.Warn("generate image fail", "err", err)
		return
	}
	
	if imageUrl != "" {
		_, err = d.Session.FollowupMessageCreate(d.Inter.Interaction, true, &discordgo.WebhookParams{
			Content: imageUrl,
		})
		if err != nil {
			logger.Warn("Sending followup interaction message failed", "err", err)
		}
	} else if len(imageContent) > 0 {
		file := &discordgo.File{
			Name:   "image." + utils.DetectImageFormat(imageContent),
			Reader: bytes.NewReader(imageContent),
		}
		_, err = d.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      strconv.Itoa(msgThinking),
			Channel: strconv.FormatInt(chatId, 10),
			Files:   []*discordgo.File{file},
		})
	}
	
	if err != nil {
		logger.Warn("send image fail", "err", err)
	}
	
	db.InsertRecordInfo(&db.Record{
		UserId:     userId,
		Question:   prompt,
		Answer:     imageUrl,
		Token:      param.ImageTokenUsage,
		IsDeleted:  1,
		RecordType: param.ImageRecordType,
	})
}

func (d *DiscordRobot) sendVideo() {
	chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	if utils.CheckUserChatExceed(userId) {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	defer utils.DecreaseUserChat(userId)
	
	if d.Robot.checkUserTokenExceed(chatId, msgId, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := d.Inter.ApplicationCommandData().Options[0].StringValue()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	msgThinking := d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
	
	var videoUrl string
	var videoContent []byte
	var err error
	
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		videoUrl, err = llm.GenerateVolVideo(prompt)
	case param.Gemini:
		videoContent, err = llm.GenerateGeminiVideo(prompt)
	default:
		err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
	}
	
	if err != nil {
		logger.Warn("generate video fail", "err", err)
		return
	}
	
	if videoUrl != "" {
		_, err = d.Session.FollowupMessageCreate(d.Inter.Interaction, true, &discordgo.WebhookParams{
			Content: videoUrl,
		})
	} else if len(videoContent) > 0 {
		file := &discordgo.File{
			Name:   "video.mp4",
			Reader: bytes.NewReader(videoContent),
		}
		_, err = d.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      strconv.Itoa(msgThinking),
			Channel: strconv.FormatInt(chatId, 10),
			Files:   []*discordgo.File{file},
		})
	}
	
	if err != nil {
		logger.Warn("send video fail", "err", err)
	}
	
	db.InsertRecordInfo(&db.Record{
		UserId:     userId,
		Question:   prompt,
		Answer:     videoUrl,
		Token:      param.VideoTokenUsage,
		IsDeleted:  1,
		RecordType: param.VideoRecordType,
	})
}

func (d *DiscordRobot) sendHelp() {
	chatId, _, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "mode", Style: discordgo.PrimaryButton, CustomID: "mode"},
				discordgo.Button{Label: "clear", Style: discordgo.PrimaryButton, CustomID: "clear"},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "balance", Style: discordgo.PrimaryButton, CustomID: "balance"},
				discordgo.Button{Label: "state", Style: discordgo.PrimaryButton, CustomID: "state"},
			},
		},
	}
	
	_, err := d.Session.ChannelMessageSendComplex(strconv.FormatInt(chatId, 10), &discordgo.MessageSend{
		Content:    "👇 chose a command：",
		Components: components,
	})
	if err != nil {
		log.Println("Failed to send help config options:", err)
	}
}
func (d *DiscordRobot) sendMultiAgent(agentType string) {
	chatId, replyToMessageID, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	if utils.CheckUserChatExceed(userId) {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		return
	}
	defer utils.DecreaseUserChat(userId)
	
	if d.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	// 获取 prompt 内容
	prompt := d.Inter.ApplicationCommandData().Options[0].StringValue()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	// 处理异步任务
	messageChan := make(chan *param.MsgInfo)
	
	dpReq := &llm.DeepseekTaskReq{
		Content:     prompt,
		UserId:      userId,
		ChatId:      chatId,
		MsgId:       replyToMessageID,
		MessageChan: messageChan,
	}
	
	go func() {
		var err error
		if agentType == "mcp_empty_content" {
			err = dpReq.ExecuteMcp()
		} else {
			err = dpReq.ExecuteTask()
		}
		if err != nil {
			d.Robot.SendMsg(chatId, err.Error(), replyToMessageID, tgbotapi.ModeMarkdown, nil)
		}
	}()
	
	go d.handleUpdate(messageChan)
}

func (d *DiscordRobot) addToken() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	userId := d.Inter.ApplicationCommandData().Options[0].StringValue()
	token := d.Inter.ApplicationCommandData().Options[1].StringValue()
	
	db.AddAvailToken(userId, utils.ParseInt(token))
	d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "add_token_succ", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}
