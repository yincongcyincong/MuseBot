package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type DiscordRobot struct {
	Session *discordgo.Session
	Msg     *discordgo.MessageCreate
	Inter   *discordgo.InteractionCreate
	
	Robot  *RobotInfo
	Prompt string
}

func StartDiscordRobot(ctx context.Context) {
	dg, err := discordgo.New("Bot " + *conf.BaseConfInfo.DiscordBotToken)
	if err != nil {
		logger.Fatal("create discord bot", "err", err)
	}
	dg.Client = utils.GetRobotProxyClient()
	
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
	
	select {
	case <-ctx.Done():
		dg.Close()
	}
}

func NewDiscordRobot(s *discordgo.Session, msg *discordgo.MessageCreate, i *discordgo.InteractionCreate) *DiscordRobot {
	return &DiscordRobot{
		Session: s,
		Msg:     msg,
		Inter:   i,
	}
}

func (d *DiscordRobot) checkValid() bool {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	// check whether you have new message
	if d.Msg != nil {
		if d.skipThisMsg() {
			logger.Warn("skip this msg", "msgId", msgId, "chat", chatId, "content", d.Msg.Content)
			return false
		}
		
		return true
	}
	
	return false
}

func (d *DiscordRobot) getMsgContent() string {
	return d.Msg.Content
}

func (d *DiscordRobot) requestLLMAndResp(content string) {
	d.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			d.executeChain(content)
		} else {
			d.executeLLM(content)
		}
	})
	
}

func (d *DiscordRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	
	go d.Robot.ExecChain(content, messageChan, nil)
	// send response message
	go d.handleUpdate(&MsgChan{
		NormalMessageChan: messageChan,
	})
}

func (d *DiscordRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	
	// request DeepSeek API
	go d.Robot.ExecLLM(content, messageChan, nil)
	
	// send response message
	go d.handleUpdate(&MsgChan{
		NormalMessageChan: messageChan,
	})
}

func (d *DiscordRobot) handleUpdate(messageChan *MsgChan) {
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
	for msg = range messageChan.NormalMessageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from llm!"
		}
		
		if msg.MsgId == "" && originalMsgID != "" {
			msg.MsgId = originalMsgID
		}
		
		if d.Msg != nil {
			if msg.MsgId == "" && originalMsgID == "" {
				_, err = d.Session.ChannelMessageSend(channelID, msg.Content)
				if err != nil {
					logger.Warn("Sending message failed", "err", err)
				}
			} else {
				_, err = d.Session.ChannelMessageEdit(channelID, msg.MsgId, msg.Content)
				if err != nil {
					logger.Warn("Editing message failed", "msgID", msg.MsgId, "err", err)
				}
				originalMsgID = ""
			}
		} else if d.Inter != nil {
			if msg.MsgId == "" && originalMsgID == "" {
				_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
					Content: &msg.Content,
				})
				if err != nil {
					logger.Warn("Sending interaction response failed", "err", err)
				}
			} else {
				_, err = d.Session.FollowupMessageCreate(d.Inter.Interaction, true, &discordgo.WebhookParams{
					Content: msg.Content,
				})
				if err != nil {
					logger.Warn("Editing followup interaction message failed", "err", err)
				}
				originalMsgID = ""
			}
		}
	}
}

func (d *DiscordRobot) GetContent(defaultText string) (string, error) {
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
			{Type: discordgo.ApplicationCommandOptionAttachment, Name: "image", Description: "upload a image", Required: false},
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
	}
	
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			logger.Error("Cannot create command", "cmd", cmd.Name, "err", err)
		}
	}
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("onInteractionCreate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	d := NewDiscordRobot(s, nil, i)
	d.Robot = NewRobot(WithRobot(d))
	d.Robot.Exec()
	
	cmd := ""
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		cmd = i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		d.changeMode(i.MessageComponentData().CustomID)
	}
	
	d.Robot.ExecCmd(cmd, d.sendChatMessage)
}

func (d *DiscordRobot) changeMode(mode string) {
	if param.GeminiModels[mode] || param.OpenAIModels[mode] ||
		param.DeepseekModels[mode] || param.DeepseekLocalModels[mode] ||
		param.OpenRouterModels[mode] || param.VolModels[mode] {
		d.Robot.handleModeUpdate(mode)
	}
	
	if param.OpenRouterModelTypes[mode] {
		buttons := make([]discordgo.MessageComponent, 0)
		for k := range param.OpenRouterModels {
			if strings.Contains(k, mode+"/") {
				buttons = append(buttons, discordgo.Button{Label: mode, CustomID: mode, Style: discordgo.SecondaryButton})
			}
		}
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
			logger.Error("Failed to defer interaction response", "err", err)
		}
		
	}
}

func (d *DiscordRobot) sendChatMessage() {
	prompt := ""
	if d.Inter != nil && d.Inter.Type == discordgo.InteractionApplicationCommand && len(d.Inter.ApplicationCommandData().Options) > 0 {
		prompt = d.Inter.ApplicationCommandData().Options[0].StringValue()
	}
	d.Prompt = prompt
	d.requestLLMAndResp(prompt)
}

func (d *DiscordRobot) sendModeConfigurationOptions() {
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

func (d *DiscordRobot) sendImg() {
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := d.Inter.ApplicationCommandData().Options[0].StringValue()
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		
		var lastImageContent []byte
		var err error
		
		if d.Inter.ApplicationCommandData().GetOption("image") != nil {
			if attachment, ok := d.Inter.ApplicationCommandData().GetOption("image").Value.(string); ok {
				lastImageContent, err = utils.DownloadFile(d.Inter.ApplicationCommandData().Resolved.Attachments[attachment].URL)
				if err != nil {
					logger.Warn("download image fail", "err", err)
				}
			}
		}
		
		if len(lastImageContent) == 0 {
			lastImageContent, err = d.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		var imageUrl string
		var imageContent []byte
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			imageUrl, totalToken, err = llm.GenerateVolImg(prompt, lastImageContent)
		case param.OpenAi:
			imageContent, totalToken, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
		case param.Gemini:
			imageContent, totalToken, err = llm.GenerateGeminiImg(prompt, lastImageContent)
		default:
			err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		if len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				return
			}
		}
		
		file := &discordgo.File{
			Name:   "image." + utils.DetectImageFormat(imageContent),
			Reader: bytes.NewReader(imageContent),
		}
		_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{file},
		})
		
		if err != nil {
			logger.Warn("send image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(imageContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

func (d *DiscordRobot) sendVideo() {
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := d.Inter.ApplicationCommandData().Options[0].StringValue()
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		
		var imageContent []byte
		var err error
		if d.Inter.ApplicationCommandData().GetOption("image") != nil {
			if attachment, ok := d.Inter.ApplicationCommandData().GetOption("image").Value.(string); ok {
				imageContent, err = utils.DownloadFile(d.Inter.ApplicationCommandData().Resolved.Attachments[attachment].URL)
				if err != nil {
					logger.Warn("download image fail", "err", err)
				}
			}
		}
		
		var videoUrl string
		var videoContent []byte
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			videoUrl, totalToken, err = llm.GenerateVolVideo(prompt, imageContent)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt, imageContent)
		default:
			err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		if len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				return
			}
		}
		
		file := &discordgo.File{
			Name:   "video." + utils.DetectVideoMimeType(videoContent),
			Reader: bytes.NewReader(videoContent),
		}
		
		_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{file},
		})
		
		if err != nil {
			logger.Warn("send video fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", utils.DetectVideoMimeType(videoContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
}

func (d *DiscordRobot) sendHelpConfigurationOptions() {
	chatId, replyToMessageID, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	d.Robot.SendMsg(chatId, helpText, replyToMessageID, tgbotapi.ModeMarkdown, nil)
}

func (d *DiscordRobot) getPrompt() string {
	return d.Prompt
}
