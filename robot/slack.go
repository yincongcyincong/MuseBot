package robot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	SlackClient  *slack.Client
	socketClient *socketmode.Client
	slackUserId  string
)

type SlackRobot struct {
	Event    *slackevents.MessageEvent
	CmdEvent *slack.SlashCommand
	Callback *slack.InteractionCallback
	
	Robot   *RobotInfo
	Client  *slack.Client
	Command string
	Prompt  string
	BotName string
	
	ImageContent []byte
	VoiceContent []byte
	UserName     string
}

func StartSlackRobot(ctx context.Context) {
	if *conf.BaseConfInfo.SlackAppToken == "" || *conf.BaseConfInfo.SlackBotToken == "" {
		return
	}
	
	SlackClient = slack.New(
		*conf.BaseConfInfo.SlackBotToken,
		slack.OptionDebug(false),
		slack.OptionAppLevelToken(*conf.BaseConfInfo.SlackAppToken),
		slack.OptionLog(logger.Logger),
		slack.OptionHTTPClient(utils.GetRobotProxyClient()),
	)
	socketClient = socketmode.New(SlackClient)
	
	authResp, err := SlackClient.AuthTest()
	if err != nil {
		logger.ErrorCtx(ctx, "Slack auth failed", "err", err)
		return
	}
	slackUserId = authResp.UserID
	
	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				socketClient.Ack(*evt.Request)
				if innerEvent, ok := evt.Data.(slackevents.EventsAPIEvent); ok {
					if innerEvent.Type == slackevents.CallbackEvent {
						switch ev := innerEvent.InnerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							if ev.BotID == "" && (ev.Text != "" || len(ev.Message.Files) > 0) {
								SlackMessageHandler(ev)
							}
						}
					}
				}
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					continue
				}
				socketClient.Ack(*evt.Request)
				SlackCmdHandler(&cmd)
			
			case socketmode.EventTypeInteractive:
				
				interaction, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					continue
				}
				socketClient.Ack(*evt.Request)
				
				switch interaction.Type {
				case slack.InteractionTypeBlockActions:
					SlackButtonHandler(&interaction)
				case slack.InteractionTypeViewSubmission:
					submissionHandler(&interaction)
				}
			}
		}
		
	}()
	logger.Info("SlackBot Info", "username", authResp.User)
	err = socketClient.RunContext(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "SlackBot Run failed", "err", err)
	}
}

func NewSlackRobot(message *slackevents.MessageEvent, command *slack.SlashCommand,
	callback *slack.InteractionCallback) *SlackRobot {
	metrics.AppRequestCount.WithLabelValues("slack").Inc()
	sr := &SlackRobot{
		Event:    message,
		CmdEvent: command,
		Callback: callback,
		Client:   SlackClient,
	}
	if message != nil {
		sr.UserName = message.User
	}
	if callback != nil {
		sr.UserName = callback.User.Name
	}
	return sr
}

func SlackButtonHandler(callback *slack.InteractionCallback) {
	
	s := NewSlackRobot(nil, nil, callback)
	s.Robot = NewRobot(WithRobot(s))
	
	for _, action := range callback.ActionCallback.BlockActions {
		s.Command = action.ActionID
		switch action.ActionID {
		case "chat", "photo", "video", "mcp", "task":
			s.openModal(callback.TriggerID, action.ActionID)
		default:
			s.Robot.ExecCmd(s.Command, nil, nil, nil)
			
		}
	}
}

func SlackCmdHandler(command *slack.SlashCommand) {
	s := NewSlackRobot(nil, command, nil)
	s.Robot = NewRobot(WithRobot(s))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "Slack exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		s.Command = command.Command
		s.Prompt = command.Text
		s.Robot.AddUserInfo()
		s.Robot.ExecCmd(s.Command, s.sendChatMessage, nil, nil)
		
	}()
}

func SlackMessageHandler(message *slackevents.MessageEvent) {
	r := NewSlackRobot(message, nil, nil)
	r.Robot = NewRobot(WithRobot(r))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(r.Robot.Ctx, "Slack exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		r.Robot.Exec()
	}()
}

func (s *SlackRobot) checkValid() bool {
	// group need at bot
	atRobot := fmt.Sprintf("<@%s>", slackUserId)
	if (strings.HasPrefix(s.Event.Channel, "C") || strings.HasPrefix(s.Event.Channel, "G")) &&
		strings.Contains(s.Event.Text, atRobot) {
		return false
	}
	
	s.Command, s.Prompt = ParseCommand(strings.ReplaceAll(s.Event.Text, atRobot, ""))
	return s.getMessageContent()
}

func (s *SlackRobot) getMessageContent() bool {
	if s.Event != nil && s.Event.Message != nil && s.Event.Message.Files != nil && len(s.Event.Message.Files) > 0 {
		file := s.Event.Message.Files[0]
		var err error
		switch file.Mimetype {
		case "image/jpeg", "image/png", "image/gif", "image/webp":
			s.ImageContent, err = s.downloadSlackFile(file.URLPrivateDownload)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "download image failed", "err", err)
				return false
			}
		
		case "audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4":
			// 下载音频
			s.VoiceContent, err = s.downloadSlackFile(file.URLPrivateDownload)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "download audio failed", "err", err)
				return false
			}
			
			s.Prompt, err = s.Robot.GetAudioContent(s.VoiceContent)
			if err != nil {
				logger.Warn("generate text from audio failed", "err", err)
				return false
			}
		}
	}
	
	return true
}

func (s *SlackRobot) getMsgContent() string {
	return s.Command
}

func (s *SlackRobot) requestLLM(content string) {
	if !strings.Contains(content, "/") && !strings.Contains(content, "$") && s.Prompt == "" {
		s.Prompt = content
	}
	s.Robot.ExecCmd(content, s.sendChatMessage, nil, nil)
}

func (s *SlackRobot) sendChatMessage() {
	s.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			s.executeChain()
		} else {
			s.executeLLM()
		}
	})
	
}

func (s *SlackRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go s.Robot.ExecChain(s.Prompt, messageChan)
	
	go s.Robot.HandleUpdate(messageChan, "mp3")
}

func (s *SlackRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go s.Robot.ExecLLM(s.Prompt, messageChan)
	
	go s.Robot.HandleUpdate(messageChan, "mp3")
}

func (s *SlackRobot) getContent(content string) (string, error) {
	if s.Event.Message == nil || len(s.Event.Message.Files) == 0 {
		return content, nil
	}
	
	file := s.Event.Message.Files[0]
	var err error
	
	switch file.Mimetype {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		content, err = s.Robot.GetImageContent(s.ImageContent, content)
		if err != nil {
			logger.Warn("generate text from image failed", "err", err)
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported file type: %s", file.Mimetype)
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}

func (s *SlackRobot) downloadSlackFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.SlackBotToken)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %s", resp.Status)
	}
	
	return io.ReadAll(resp.Body)
}

func (s *SlackRobot) sendImg() {
	s.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := s.Prompt
		prompt = utils.ReplaceCommand(prompt, "/photo", s.BotName)
		if prompt == "" {
			logger.Warn("prompt is empty")
			s.Robot.SendMsg(chatId, i18n.GetMessage("photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var err error
		lastImageContent := s.ImageContent
		if len(lastImageContent) == 0 && strings.Contains(s.Command, "edit_photo") {
			lastImageContent, err = s.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := s.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "generate image fail", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = s.sendMedia(imageContent, utils.DetectImageFormat(imageContent), "image")
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "send image fail", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		s.Robot.saveRecord(imageContent, lastImageContent, param.ImageRecordType, totalToken)
	})
}

func (s *SlackRobot) sendMedia(media []byte, contentType, sType string) error {
	chatId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	if sType == "image" {
		uploadParams := slack.UploadFileV2Parameters{
			Filename: "image." + contentType,
			Reader:   bytes.NewReader(media),
			Title:    "image",
			FileSize: len(media),
			Channel:  chatId,
		}
		
		_, err := s.Client.UploadFileV2(uploadParams)
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "upload image to slack fail", "err", err)
			return err
		}
	} else {
		uploadParams := slack.UploadFileV2Parameters{
			Filename: "video." + contentType,
			Reader:   bytes.NewReader(media),
			Title:    "video",
			FileSize: len(media),
			Channel:  chatId,
		}
		
		_, err := s.Client.UploadFileV2(uploadParams)
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "upload image to slack fail", "err", err)
			return err
		}
	}
	
	return nil
}

func (s *SlackRobot) sendVideo() {
	s.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := s.Prompt
		prompt = utils.ReplaceCommand(prompt, "/video", s.BotName)
		if prompt == "" {
			logger.Warn("prompt is empty")
			s.Robot.SendMsg(chatId, i18n.GetMessage("video_empty_content", nil),
				replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var err error
		imageContent := s.ImageContent
		videoContent, totalToken, err := s.Robot.CreateVideo(prompt, imageContent)
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "generate video failed", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = s.sendMedia(videoContent, utils.DetectVideoMimeType(videoContent), "video")
		if err != nil {
			logger.ErrorCtx(s.Robot.Ctx, "send video failed", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		s.Robot.saveRecord(videoContent, imageContent, param.VideoRecordType, totalToken)
	})
}

func (s *SlackRobot) openModal(triggerID, actionID string) {
	chatId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	modalRequest := slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           slack.NewTextBlockObject("plain_text", "prompt", false, false),
		Close:           slack.NewTextBlockObject("plain_text", "cancel", false, false),
		Submit:          slack.NewTextBlockObject("plain_text", "submit", false, false),
		CallbackID:      chatId,
		PrivateMetadata: actionID,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewInputBlock(
					"input_block",
					slack.NewTextBlockObject("plain_text", "input prompt", false, false),
					slack.NewTextBlockObject("plain_text", "input prompt", false, false),
					slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "prompt...", false, false), "input_action"),
				),
			},
		},
	}
	
	_, err := s.Client.OpenView(triggerID, modalRequest)
	if err != nil {
		logger.ErrorCtx(s.Robot.Ctx, "open modal failed", "err", err)
	}
}

func submissionHandler(callback *slack.InteractionCallback) {
	s := NewSlackRobot(nil, nil, callback)
	s.Robot = NewRobot(WithRobot(s))
	
	s.Command = callback.View.PrivateMetadata
	values := callback.View.State.Values
	inputBlock := values["input_block"]
	for _, v := range inputBlock {
		s.Prompt += v.Value
	}
	s.Callback.Channel.ID = callback.View.CallbackID
	
	s.Robot.AddUserInfo()
	s.getMessageContent()
	s.Robot.ExecCmd(s.Command, nil, nil, nil)
	
}

func (s *SlackRobot) getPrompt() string {
	return s.Prompt
}

func (s *SlackRobot) getPerMsgLen() int {
	return 1800
}

func (s *SlackRobot) sendTextStream(messageChan *MsgChan) {
	chatId, messageId, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	
	for msg := range messageChan.NormalMessageChan {
		if msg.Content == "" {
			msg.Content = "get nothing from llm!"
		}
		
		if msg.MsgId == "" {
			msg.MsgId = s.Robot.SendMsg(chatId, msg.Content, messageId, tgbotapi.ModeMarkdown, nil)
		} else {
			_, _, _, err := s.Client.UpdateMessage(
				chatId,
				msg.MsgId,
				slack.MsgOptionText(msg.Content, false),
				slack.MsgOptionTS(messageId),
			)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "update message failed", "err", err)
				continue
			}
		}
	}
}

func (s *SlackRobot) sendText(messageChan *MsgChan) {
	
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			s.sendMsg(msg)
		}
	}
	
	if msg != nil {
		s.sendMsg(msg)
	}
	
}

func (s *SlackRobot) sendMsg(msg *param.MsgInfo) {
	chatId, msgId, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	blocks := utils.ExtractContentBlocks(msg.Content)
	
	for _, b := range blocks {
		switch b.Type {
		case "text":
			s.Robot.SendMsg(chatId, b.Content, msgId, tgbotapi.ModeMarkdown, nil)
		case "video", "image":
			content, err := utils.DownloadFile(b.Media.URL)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "download file fail", "err", err)
				continue
			}
			contentType := ""
			if b.Type == "video" {
				contentType = utils.DetectVideoMimeType(content)
			} else {
				contentType = utils.DetectImageFormat(content)
			}
			
			err = s.sendMedia(content, contentType, b.Type)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "send media fail", "err", err)
			}
		}
	}
}

func (s *SlackRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	chatId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	format := utils.DetectAudioFormat(voiceContent)
	uploadParams := slack.UploadFileV2Parameters{
		Filename: "voice." + format,
		Reader:   bytes.NewReader(voiceContent),
		Title:    "voice",
		FileSize: len(voiceContent),
		Channel:  chatId,
	}
	
	_, err := s.Client.UploadFileV2(uploadParams)
	if err != nil {
		logger.Warn("upload voice to slack fail", "err", err)
		return err
	}
	
	return nil
}

func (s *SlackRobot) setCommand(command string) {
	s.Command = command
}

func (s *SlackRobot) getCommand() string {
	return s.Command
}

func (s *SlackRobot) getUserName() string {
	return s.UserName
}

func (s *SlackRobot) setPrompt(prompt string) {
	s.Prompt = prompt
}

func (s *SlackRobot) getAudio() []byte {
	return s.VoiceContent
}

func (s *SlackRobot) getImage() []byte {
	return s.ImageContent
}
