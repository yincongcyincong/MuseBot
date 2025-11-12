package robot

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
	
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"layeh.com/gopus"
)

type VolDialog struct {
	VolWsConn *websocket.Conn
	Audio     []byte
	
	CallUserId string
	Token      int
	Ctx        context.Context
	Cancel     context.CancelFunc
}

var (
	volDialog = &VolDialog{
		Audio: make([]byte, 0),
	}
	
	DiscordSession *discordgo.Session
)

type DiscordRobot struct {
	Session *discordgo.Session
	Msg     *discordgo.MessageCreate
	Inter   *discordgo.InteractionCreate
	
	Robot        *RobotInfo
	Prompt       string
	Command      string
	ImageContent []byte
	AudioContent []byte
	UserName     string
}

func StartDiscordRobot(ctx context.Context) {
	var err error
	DiscordSession, err = discordgo.New("Bot " + *conf.BaseConfInfo.DiscordBotToken)
	if err != nil {
		logger.ErrorCtx(ctx, "create discord bot", "err", err)
		return
	}
	DiscordSession.Client = utils.GetRobotProxyClient()
	
	// 添加消息处理函数
	DiscordSession.AddHandler(messageCreate)
	DiscordSession.AddHandler(onInteractionCreate)
	// 监听语音状态更新事件
	DiscordSession.AddHandler(voiceStateUpdate)
	
	// 打开连接
	err = DiscordSession.Open()
	if err != nil {
		logger.ErrorCtx(ctx, "connect fail", "err", err)
		return
	}
	
	logger.InfoCtx(ctx, "discordBot Info", "username", DiscordSession.State.User.Username)
	
	registerSlashCommands(DiscordSession)
	
	select {
	case <-ctx.Done():
		DiscordSession.Close()
	}
}

func NewDiscordRobot(s *discordgo.Session, msg *discordgo.MessageCreate, i *discordgo.InteractionCreate) *DiscordRobot {
	metrics.AppRequestCount.WithLabelValues("discord").Inc()
	dr := &DiscordRobot{
		Session: s,
		Msg:     msg,
		Inter:   i,
	}
	
	if msg != nil {
		dr.UserName = msg.Author.Username
	}
	
	if i != nil {
		dr.UserName = i.User.Username
	}
	
	return dr
}

func (d *DiscordRobot) checkValid() bool {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	if d.Msg != nil {
		if d.skipThisMsg() {
			logger.WarnCtx(d.Robot.Ctx, "skip this msg", "msgId", msgId, "chat", chatId, "content", d.Msg.Content)
			return false
		}
		d.Command, d.Prompt = ParseCommand(d.Msg.Content)
		d.getMessageContent()
		return true
	}
	
	if d.Inter != nil {
		switch d.Inter.Type {
		case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
			d.Command = d.Inter.ApplicationCommandData().Name
		}
		
		if d.Inter != nil && d.Inter.Type == discordgo.InteractionApplicationCommand && len(d.Inter.ApplicationCommandData().Options) > 0 {
			d.Prompt = d.Inter.ApplicationCommandData().Options[0].StringValue()
		}
		return true
	}
	
	return false
}

func (d *DiscordRobot) getMsgContent() string {
	if d.Msg != nil {
		return d.Msg.Content
	}
	return ""
}

func (d *DiscordRobot) getMessageContent() {
	var err error
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	if d.Inter != nil && d.Inter.ApplicationCommandData().GetOption("image") != nil {
		if attachment, ok := d.Inter.ApplicationCommandData().GetOption("image").Value.(string); ok {
			d.ImageContent, err = utils.DownloadFile(d.Inter.ApplicationCommandData().Resolved.Attachments[attachment].URL)
			if err != nil {
				logger.WarnCtx(d.Robot.Ctx, "download image fail", "err", err)
			}
		}
	}
	
	if d.Msg != nil {
		attachments := d.Msg.Attachments
		if len(attachments) > 0 {
			for _, att := range attachments {
				if strings.HasPrefix(att.ContentType, "audio/") {
					d.AudioContent, err = utils.DownloadFile(att.URL)
					if d.AudioContent == nil || err != nil {
						logger.ErrorCtx(d.Robot.Ctx, "audio url empty", "url", att.URL, "err", err)
						d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
						return
					}
					if d.AudioContent != nil {
						d.Prompt, err = d.Robot.GetAudioContent(d.AudioContent)
						if err != nil {
							logger.WarnCtx(d.Robot.Ctx, "get audio content err", "err", err)
							d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
							return
						}
					}
				}
				
				if strings.HasPrefix(att.ContentType, "image/") {
					d.ImageContent, err = utils.DownloadFile(att.URL)
					if d.ImageContent == nil || err != nil {
						logger.ErrorCtx(d.Robot.Ctx, "image url empty", "url", att.URL, "err", err)
						d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
						return
					}
				}
			}
		}
	}
}

func (d *DiscordRobot) requestLLM(content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorCtx(d.Robot.Ctx, "DiscordRobot panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		switch d.Command {
		case "talk":
			d.Talk()
			return
		}
		
		d.Robot.ExecCmd(d.Command, d.sendChatMessage, nil, nil)
	}()
}

func (d *DiscordRobot) executeChain(content string) {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	
	go d.Robot.ExecChain(content, messageChan)
	
	go d.Robot.HandleUpdate(messageChan, "mp3")
}

func (d *DiscordRobot) executeLLM(content string) {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	
	go d.Robot.ExecLLM(content, messageChan)
	
	go d.Robot.HandleUpdate(messageChan, "mp3")
}

func (d *DiscordRobot) sendText(messageChan *MsgChan) {
	
	var originalMsgID string
	var channelID string
	var err error
	
	if d.Msg != nil {
		channelID = d.Msg.ChannelID
		
		thinkingMsg, err := d.Session.ChannelMessageSend(channelID, i18n.GetMessage("thinking", nil))
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "Sending thinking message failed", "err", err)
		} else {
			originalMsgID = thinkingMsg.ID
		}
		
	} else if d.Inter != nil {
		channelID = d.Inter.ChannelID
		
		err = d.Session.InteractionRespond(d.Inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "Failed to defer interaction response", "err", err)
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
					logger.WarnCtx(d.Robot.Ctx, "Sending message failed", "err", err)
				}
			} else {
				_, err = d.Session.ChannelMessageEdit(channelID, msg.MsgId, msg.Content)
				if err != nil {
					logger.WarnCtx(d.Robot.Ctx, "Editing message failed", "msgID", msg.MsgId, "err", err)
				}
				originalMsgID = ""
			}
		} else if d.Inter != nil {
			if msg.MsgId == "" && originalMsgID == "" {
				_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
					Content: &msg.Content,
				})
				if err != nil {
					logger.WarnCtx(d.Robot.Ctx, "Sending interaction response failed", "err", err)
				}
			} else {
				_, err = d.Session.FollowupMessageCreate(d.Inter.Interaction, true, &discordgo.WebhookParams{
					Content: msg.Content,
				})
				if err != nil {
					logger.WarnCtx(d.Robot.Ctx, "Editing followup interaction message failed", "err", err)
				}
				originalMsgID = ""
			}
		}
	}
}

func (d *DiscordRobot) getContent(content string) (string, error) {
	
	var err error
	if d.ImageContent != nil {
		content, err = d.Robot.GetImageContent(d.ImageContent, content)
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "get image content err", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.WarnCtx(d.Robot.Ctx, "content empty")
		return "", errors.New("content empty")
	}
	
	if d.Session != nil && d.Session.State != nil && d.Session.State.User != nil {
		content = strings.ReplaceAll(content, "<@"+d.Session.State.User.ID+">", "")
	}
	
	return content, nil
}

func (d *DiscordRobot) skipThisMsg() bool {
	if d.Msg == nil || d.Msg.Author == nil ||
		d.Session == nil || d.Msg.Author.ID == d.Session.State.User.ID {
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
		{Name: param.Chat, Description: i18n.GetMessage("commands.chat.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
			{Type: discordgo.ApplicationCommandOptionAttachment, Name: "image", Description: "upload a image", Required: false},
		}},
		{Name: param.TxtType, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: param.PhotoType, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: param.VideoType, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: param.TxtModel, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: param.PhotoModel, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: param.VideoModel, Description: i18n.GetMessage("commands.mode.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Type", Required: false},
		}},
		{Name: "talk", Description: i18n.GetMessage("commands.talk.description", nil)},
		{Name: param.State, Description: i18n.GetMessage("commands.state.description", nil)},
		{Name: param.Clear, Description: i18n.GetMessage("commands.clear.description", nil)},
		{Name: param.Retry, Description: i18n.GetMessage("commands.retry.description", nil)},
		{Name: param.Photo, Description: i18n.GetMessage("commands.photo.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: param.EditPhoto, Description: i18n.GetMessage("commands.photo.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
			{Type: discordgo.ApplicationCommandOptionAttachment, Name: "image", Description: "upload a image", Required: false},
		}},
		{Name: param.Video, Description: i18n.GetMessage("commands.video.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: param.Help, Description: i18n.GetMessage("commands.help.description", nil)},
		{Name: param.Task, Description: i18n.GetMessage("commands.task.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: param.Mcp, Description: i18n.GetMessage("commands.mcp.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "prompt", Description: "Prompt", Required: true},
		}},
		{Name: param.CronDel, Description: i18n.GetMessage("commands.cron.description", nil), Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "id", Required: true},
		}},
		{Name: param.CronClear, Description: i18n.GetMessage("commands.cron.description", nil)},
		{Name: param.CronDel, Description: i18n.GetMessage("commands.cron.description", nil)},
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
}

func (d *DiscordRobot) sendChatMessage() {
	d.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			d.executeChain(d.Prompt)
		} else {
			d.executeLLM(d.Prompt)
		}
	})
}

func (d *DiscordRobot) sendImg() {
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(d.getPrompt())
		if prompt == "" {
			d.Robot.SendMsg(chatId, i18n.GetMessage("video_empty_content", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		d.Robot.SendMsg(chatId, i18n.GetMessage("thinking", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		
		var lastImageContent = d.ImageContent
		var err error
		if len(lastImageContent) == 0 && strings.Contains(d.Command, "edit_photo") {
			lastImageContent, err = d.Robot.GetLastImageContent()
			if err != nil {
				logger.WarnCtx(d.Robot.Ctx, "get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := d.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "generate image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		file := &discordgo.File{
			Name:   "image." + utils.DetectImageFormat(imageContent),
			Reader: bytes.NewReader(imageContent),
		}
		
		if d.Inter != nil {
			_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
				Files: []*discordgo.File{file},
			})
		} else {
			messageSend := &discordgo.MessageSend{
				Reference: &discordgo.MessageReference{
					MessageID: msgId,
					ChannelID: chatId,
				},
				Files: []*discordgo.File{file},
			}
			_, err = d.Session.ChannelMessageSendComplex(chatId, messageSend)
			if err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "Error sending message:", "err", err)
				return
			}
		}
		
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "send image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		d.Robot.saveRecord(imageContent, lastImageContent, param.ImageRecordType, totalToken)
	})
}

func (d *DiscordRobot) sendVideo() {
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(d.getPrompt())
		if prompt == "" {
			d.Robot.SendMsg(chatId, i18n.GetMessage("video_empty_content", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		d.Robot.SendMsg(chatId, i18n.GetMessage("thinking", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		
		var imageContent = d.ImageContent
		videoContent, totalToken, err := d.Robot.CreateVideo(prompt, imageContent)
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "generate video fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		file := &discordgo.File{
			Name:   "video." + utils.DetectVideoMimeType(videoContent),
			Reader: bytes.NewReader(videoContent),
		}
		
		if d.Inter != nil {
			_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
				Files: []*discordgo.File{file},
			})
		} else {
			messageSend := &discordgo.MessageSend{
				Reference: &discordgo.MessageReference{
					MessageID: msgId,
					ChannelID: chatId,
				},
				Files: []*discordgo.File{file},
			}
			_, err = d.Session.ChannelMessageSendComplex(chatId, messageSend)
			if err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "Error sending message:", "err", err)
				return
			}
		}
		
		if err != nil {
			logger.WarnCtx(d.Robot.Ctx, "send video fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, param.DiscordEditMode, nil)
			return
		}
		
		d.Robot.saveRecord(videoContent, imageContent, param.VideoRecordType, totalToken)
	})
}

func (d *DiscordRobot) getPrompt() string {
	return d.Prompt
}

func (d *DiscordRobot) getPerMsgLen() int {
	return 1800
}

func (d *DiscordRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	var err error
	if d.Msg != nil {
		_, err = d.Session.ChannelFileSend(d.Msg.ChannelID, "voice."+utils.DetectAudioFormat(voiceContent), bytes.NewReader(voiceContent))
		
	} else if d.Inter != nil {
		_, err = d.Session.InteractionResponseEdit(d.Inter.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:   "voice." + utils.DetectAudioFormat(voiceContent),
					Reader: bytes.NewReader(voiceContent),
				},
			},
		})
	}
	
	return err
}

func (d *DiscordRobot) Talk() {
	d.Robot.TalkingPreCheck(func() {
		gid := d.Inter.GuildID
		cid, replyToMessageID, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		if gid == "" || cid == "" {
			d.Robot.SendMsg(cid, i18n.GetMessage("talk_param_error", nil),
				replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(d.Session.VoiceConnections) != 0 {
			d.Robot.SendMsg(cid, i18n.GetMessage("bot_talking", nil),
				replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("recover panic", "err", err, "stack", string(debug.Stack()))
				}
			}()
			
			vc, err := d.Session.ChannelVoiceJoin(gid, cid, false, false)
			if err != nil {
				logger.Error("join voice fail", "err", err)
				return
			}
			
			wsURL := url.URL{Scheme: "wss", Host: "openspeech.bytedance.com", Path: "/api/v3/realtime/dialogue"}
			volDialog.VolWsConn, _, err = websocket.DefaultDialer.DialContext(context.Background(), wsURL.String(), http.Header{
				"X-Api-Resource-Id": []string{"volc.speech.dialog"},
				"X-Api-Access-Key":  []string{*conf.AudioConfInfo.VolAudioToken},
				"X-Api-App-Key":     []string{"PlgvMymc7f3tQnJ6"},
				"X-Api-App-ID":      []string{*conf.AudioConfInfo.VolAudioAppID},
				"X-Api-Connect-Id":  []string{uuid.New().String()},
			})
			if err != nil {
				logger.Error("connect vol fail", "err", err)
				return
			}
			
			err = utils.StartConnection(volDialog.VolWsConn)
			if err != nil {
				logger.Error("start connect fail", "err", err)
				return
			}
			err = utils.StartSession(volDialog.VolWsConn, userId, &utils.StartSessionPayload{
				ASR: utils.ASRPayload{
					Extra: map[string]interface{}{
						"end_smooth_window_ms": *conf.AudioConfInfo.VolEndSmoothWindow,
					},
				},
				TTS: utils.TTSPayload{
					Speaker: *conf.AudioConfInfo.VolTTSSpeaker,
					AudioConfig: utils.AudioConfig{
						Channel:    2,
						Format:     "pcm_s16le",
						SampleRate: 48000,
					},
				},
				Dialog: utils.DialogPayload{
					BotName:       *conf.AudioConfInfo.VolBotName,
					SystemRole:    *conf.AudioConfInfo.VolSystemRole,
					SpeakingStyle: *conf.AudioConfInfo.VolSpeakingStyle,
					Extra: map[string]interface{}{
						"strict_audit":   false,
						"audit_response": "抱歉这个问题我无法回答，你可以换个其他话题，我会尽力为你提供帮助。",
						"input_mod":      "audio_file",
					},
				},
			})
			if err != nil {
				logger.Error("start session fail", "err", err)
				return
			}
			
			volDialog.Ctx, volDialog.Cancel = context.WithCancel(context.Background())
			volDialog.CallUserId = userId
			
			go d.PlayAudioToDiscord(vc)
			
			go d.receiveVoice(vc)
		}()
	})
	
}

func (d *DiscordRobot) PlayAudioToDiscord(vc *discordgo.VoiceConnection) {
	defer func() {
		CloseTalk(vc)
	}()
	
	for {
		select {
		case <-volDialog.Ctx.Done():
			return
		default:
			msg, err := utils.ReceiveMessage(volDialog.VolWsConn)
			if err != nil {
				logger.Error("receive message", "err", err)
				return
			}
			
			switch msg.Type {
			case utils.MsgTypeFullServer:
				switch msg.Event {
				case 152, 153:
					logger.WarnCtx(d.Robot.Ctx, "session finished")
					return
				case 154:
					usage := utils.GetDialogUsage(msg.Payload)
					if usage.Usage != nil {
						volDialog.Token += usage.Usage.CachedAudioTokens + usage.Usage.OutputAudioTokens + usage.Usage.InputAudioTokens +
							usage.Usage.CachedTextTokens + usage.Usage.OutputTextTokens + usage.Usage.InputTextTokens
					}
				
				case 350, 451:
					logger.Info("start event", "event", msg.Event, "type", msg.TypeFlag(), "payload", string(msg.Payload))
				case 352:
					utils.HandleIncomingAudio(msg.Payload)
					volDialog.Audio = append(volDialog.Audio, msg.Payload...)
				case 351, 359:
					utils.HandleIncomingAudio(msg.Payload)
					volDialog.Audio = append(volDialog.Audio, msg.Payload...)
					d.sendAudioToDiscord(vc, volDialog.Audio)
					volDialog.Audio = volDialog.Audio[:0]
				}
			case utils.MsgTypeAudioOnlyServer:
				utils.HandleIncomingAudio(msg.Payload)
				volDialog.Audio = append(volDialog.Audio, msg.Payload...)
			case utils.MsgTypeError:
				logger.Error("Receive Error message", "code", msg.ErrorCode, "payload", string(msg.Payload))
			default:
				logger.Error("Received unexpected message type", "type", msg.Type)
			}
		}
	}
}

func (d *DiscordRobot) sendAudioToDiscord(vc *discordgo.VoiceConnection, audioContent []byte) {
	mono16k := bytesToInt16LE(audioContent)
	
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		logger.Error("gopus encoder fail", "err", err)
		return
	}
	encoder.SetBitrate(64000)
	
	const samplesPerFrame = 960
	const monoFrameSize = 320
	
	for i := 0; i < len(mono16k); i += monoFrameSize {
		end := i + monoFrameSize
		if end > len(mono16k) {
			end = len(mono16k)
		}
		
		monoFrame := mono16k[i:end]
		
		stereo48k := upsampleAndStereoLinear(monoFrame)
		
		opus, err := encoder.Encode(stereo48k, samplesPerFrame, 4000)
		if err != nil {
			logger.Error("gopus encode fail", "err", err)
			break
		}
		
		vc.OpusSend <- opus
		
		time.Sleep(20 * time.Millisecond)
	}
}

func upsampleAndStereoLinear(mono16k []int16) []int16 {
	inLen := len(mono16k)
	outLen := inLen * 3            // 16kHz -> 48kHz
	out := make([]int16, outLen*2) // *2 for stereo
	
	for i := 0; i < outLen; i++ {
		// 线性插值
		pos := float64(i) / 3.0
		idx := int(pos)
		if idx >= inLen-1 {
			idx = inLen - 2
		}
		frac := pos - float64(idx)
		sample := int16((1-frac)*float64(mono16k[idx]) + frac*float64(mono16k[idx+1]))
		
		out[2*i] = sample   // left
		out[2*i+1] = sample // right
	}
	return out
}

// PCM16 byte -> int16 slice (little endian)
func bytesToInt16LE(data []byte) []int16 {
	out := make([]int16, len(data)/2)
	for i := 0; i < len(out); i++ {
		out[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return out
}

func (d *DiscordRobot) receiveVoice(vc *discordgo.VoiceConnection) {
	defer func() {
		CloseTalk(vc)
	}()
	
	decoder, err := gopus.NewDecoder(16000, 1)
	if err != nil {
		logger.Error("Failed to create opus decoder", "err", err)
		return
	}
	
	_, _, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	for {
		select {
		case <-volDialog.Ctx.Done():
			return
		case packet := <-vc.OpusRecv:
			pcm, err := decoder.Decode(packet.Opus, 960, false)
			if err != nil && !errors.Is(err, io.EOF) {
				logger.Error("Failed to decode opus packet", "err", err)
				continue
			}
			
			if len(pcm) > 0 {
				buf := make([]byte, len(pcm)*2)
				for i, v := range pcm {
					buf[2*i] = byte(v)
					buf[2*i+1] = byte(v >> 8)
				}
				
				err = utils.SendAudio(volDialog.VolWsConn, userId, buf)
				if err != nil {
					logger.Error("Failed to send PCM data", "err", err)
				}
			}
		}
	}
}

func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// 1. Get the bot's own voice state.
	// s.State.User.ID is your bot's ID.
	botVoiceState, err := s.State.VoiceState(v.GuildID, s.State.User.ID)
	if err != nil || botVoiceState == nil || botVoiceState.ChannelID == "" {
		// If the bot isn't in a voice channel, there's no need to handle voice state updates.
		return
	}
	
	// 2. Check if the event is relevant to the bot's channel.
	// We need to check both v.ChannelID and v.BeforeUpdate.ChannelID for user joins and leaves.
	isRelevant := false
	if v.ChannelID != "" && v.ChannelID == botVoiceState.ChannelID {
		// The event occurred in the bot's channel (user joined).
		isRelevant = true
	} else if v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID == botVoiceState.ChannelID {
		// The event occurred in the bot's channel (user left).
		isRelevant = true
	}
	
	// If the event is not relevant to the bot's channel, return early.
	if !isRelevant {
		return
	}
	
	g, err := s.State.Guild(v.GuildID)
	if err != nil {
		logger.Error("get guild fail", "err", err)
		return
	}
	
	count := 0
	for _, vs := range g.VoiceStates {
		if vs.ChannelID == botVoiceState.ChannelID {
			count++
		}
	}
	
	if count <= 1 {
		if s.VoiceConnections[v.GuildID] != nil {
			CloseTalk(s.VoiceConnections[v.GuildID])
		} else {
			logger.Error("join voice fail", "err", err)
		}
	}
	
}

func CloseTalk(vc *discordgo.VoiceConnection) {
	err := volDialog.VolWsConn.Close()
	if err == nil {
		vc.Disconnect()
		volDialog.Cancel()
		db.InsertRecordInfo(context.Background(), &db.Record{
			UserId:     volDialog.CallUserId,
			Question:   "discord talk",
			Answer:     "",
			Token:      volDialog.Token,
			IsDeleted:  0,
			RecordType: param.TalkRecordType,
			Mode:       "vol",
		})
		volDialog.Token = 0
	}
}

func (d *DiscordRobot) setCommand(command string) {
	d.Command = command
}

func (d *DiscordRobot) getCommand() string {
	return d.Command
}

func (d *DiscordRobot) getUserName() string {
	return d.UserName
}
