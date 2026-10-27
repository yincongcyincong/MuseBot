package robot

import (
	"bytes"
	"context"
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
	BotName string

	VoiceContent []byte
	baseRobot
}

func StartSlackRobot(ctx context.Context) {
	if conf.BaseConfInfo.SlackAppToken == "" || conf.BaseConfInfo.SlackBotToken == "" {
		return
	}

	SlackClient = slack.New(
		conf.BaseConfInfo.SlackBotToken,
		slack.OptionDebug(false),
		slack.OptionAppLevelToken(conf.BaseConfInfo.SlackAppToken),
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
		command := action.ActionID
		if command == param.Mcp {
			command = param.Task
		}
		s.Command = command
		switch command {
		case "chat", "photo", "video", "task":
			s.openModal(callback.TriggerID, command)
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

		default:
			// any other shared file (pdf, txt, csv, html, and text-like
			// files with other extensions such as .bat/.diff) is a document
			docContent, err := s.downloadSlackFile(file.URLPrivateDownload)
			if err != nil {
				logger.ErrorCtx(s.Robot.Ctx, "download document failed", "err", err)
				return false
			}
			if utils.DetectDocumentType(file.Name, docContent) == "" {
				logger.WarnCtx(s.Robot.Ctx, "unsupported document type", "fileName", file.Name)
				chatId, msgId, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
				s.Robot.SendMsg(chatId, i18n.GetMessage("doc_unsupported_type",
					map[string]interface{}{"file_name": file.Name}), msgId, "", nil)
				return true
			}

			s.DocContent = &param.DocInfo{FileName: file.Name, Content: docContent}
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
		s.executeLLM()
	})

}

func (s *SlackRobot) executeLLM() {
	s.Robot.runLLM(s, "mp3", false)
}

func (s *SlackRobot) downloadSlackFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+conf.BaseConfInfo.SlackBotToken)

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
		s.Prompt = utils.ReplaceCommand(s.Prompt, "/photo", s.BotName)
		s.Robot.generateMedia(s, mediaImage)
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
		s.Prompt = utils.ReplaceCommand(s.Prompt, "/video", s.BotName)
		s.Robot.generateMedia(s, mediaVideo)
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

func init() {
	RegisterChannel(Channel{
		Name: param.Slack,
		MissingCfg: func() []string {
			var missing []string
			missing = required(missing, "SLACK_BOT_TOKEN", conf.BaseConfInfo.SlackBotToken)
			return required(missing, "SLACK_APP_TOKEN", conf.BaseConfInfo.SlackAppToken)
		},
		Start: StartSlackRobot,
	})
}
