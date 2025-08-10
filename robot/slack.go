package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	slackClient  *slack.Client
	socketClient *socketmode.Client
	slackUserId  string
)

type SlackRobot struct {
	Event    *slackevents.MessageEvent
	CmdEvent *slack.SlashCommand
	Callback *slack.InteractionCallback
	
	Robot   *RobotInfo
	Client  *slack.Client
	Ctx     context.Context
	Cancel  context.CancelFunc
	Command string
	Prompt  string
	BotName string
}

func StartSlackRobot() {
	if *conf.BaseConfInfo.SlackAppToken == "" || *conf.BaseConfInfo.SlackBotToken == "" {
		return
	}
	
	slackClient = slack.New(
		*conf.BaseConfInfo.SlackBotToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(*conf.BaseConfInfo.SlackAppToken),
	)
	socketClient = socketmode.New(slackClient)
	
	authResp, err := slackClient.AuthTest()
	if err != nil {
		logger.Error("Slack auth failed", "err", err)
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
	logger.Info("SlackBot Info", "username", slackUserId)
	socketClient.Run()
}

func NewSlackRobot(message *slackevents.MessageEvent, command *slack.SlashCommand,
	callback *slack.InteractionCallback) *SlackRobot {
	c, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	return &SlackRobot{
		Event:    message,
		CmdEvent: command,
		Callback: callback,
		Client:   slackClient,
		Ctx:      c,
		Cancel:   cancel,
	}
}

func SlackButtonHandler(callback *slack.InteractionCallback) {
	
	s := NewSlackRobot(nil, nil, callback)
	s.Robot = NewRobot(WithRobot(s))
	
	chatId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	
	for _, action := range callback.ActionCallback.BlockActions {
		s.Command = action.ActionID
		switch action.ActionID {
		case "chat", "photo", "video", "mcp", "task":
			s.openModal(callback.TriggerID, action.ActionID)
		case "state", "clear", "retry", "balance":
			s.Robot.ExecCmd(s.Command, func() {})
		default:
			if param.GeminiModels[action.ActionID] || param.OpenAIModels[action.ActionID] ||
				param.DeepseekModels[action.ActionID] || param.DeepseekLocalModels[action.ActionID] ||
				param.OpenRouterModels[action.ActionID] || param.VolModels[action.ActionID] {
				s.Robot.handleModeUpdate(action.ActionID)
			}
			
			if param.OpenRouterModelTypes[action.ActionID] {
				var blocks []slack.Block
				for k := range param.OpenRouterModels {
					if strings.Contains(k, action.ActionID+"/") {
						btnText := slack.NewTextBlockObject("plain_text", k, false, false)
						btn := slack.NewButtonBlockElement(k, k, btnText)
						btn.Value = k
						actionBlock := slack.NewActionBlock("select_model"+k, btn)
						blocks = append(blocks, actionBlock)
					}
				}
				
				_, _, err := s.Client.PostMessage(chatId, slack.MsgOptionBlocks(blocks...))
				if err != nil {
					logger.Error("post message failed", "err", err)
				}
			}
		}
	}
}

func SlackCmdHandler(command *slack.SlashCommand) {
	s := NewSlackRobot(nil, command, nil)
	s.Robot = NewRobot(WithRobot(s))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Slack exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		s.Command = command.Command
		s.Prompt = command.Text
		s.Robot.ExecCmd(s.Command, s.sendChatMessage)
		
	}()
}

func SlackMessageHandler(message *slackevents.MessageEvent) {
	r := NewSlackRobot(message, nil, nil)
	r.Robot = NewRobot(WithRobot(r))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Slack exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		r.Robot.Exec()
	}()
}

func (s *SlackRobot) checkValid() bool {
	// group need at bot
	atRobot := fmt.Sprintf("<@%s>", slackUserId)
	if (strings.HasPrefix(s.Event.Channel, "C") || strings.HasPrefix(s.Event.Channel, "G")) &&
		strings.Contains(s.Event.Text, atRobot) && *conf.BaseConfInfo.NeedATBOt {
		return false
	}
	
	s.Command, s.Prompt = ParseCommand(strings.ReplaceAll(s.Event.Text, atRobot, ""))
	return true
}

func (s *SlackRobot) getMsgContent() string {
	return s.Command
}

func (s *SlackRobot) requestLLMAndResp(content string) {
	if !strings.Contains(content, "/") && s.Prompt == "" {
		s.Prompt = content
	}
	s.Robot.ExecCmd(content, s.sendChatMessage)
}

func (s *SlackRobot) sendChatMessage() {
	s.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			s.executeChain(s.Prompt)
		} else {
			s.executeLLM(s.Prompt)
		}
	})
	
}

func (s *SlackRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	go s.Robot.ExecChain(content, messageChan)
	
	go s.handleUpdate(messageChan)
}

func (s *SlackRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	go s.Robot.ExecLLM(content, messageChan)
	go s.handleUpdate(messageChan)
}

func (s *SlackRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	chatId, messageId, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	
	// 先发送一个提示消息，告诉用户机器人在处理
	originalMsgID := s.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
		messageId, "", nil)
	
	for msg := range messageChan {
		if msg.Content == "" {
			msg.Content = "get nothing from llm!"
		}
		
		if originalMsgID != "" {
			msg.MsgId = originalMsgID
		}
		
		if msg.MsgId == "" && originalMsgID == "" {
			newMsgTimestamp, _, err := s.Client.PostMessage(
				chatId,
				slack.MsgOptionText(msg.Content, false),
				slack.MsgOptionTS(messageId),
			)
			if err != nil {
				logger.Error("send new message failed", "err", err)
				continue
			}
			msg.MsgId = newMsgTimestamp
		} else {
			_, _, _, err := s.Client.UpdateMessage(
				chatId,
				msg.MsgId,
				slack.MsgOptionText(msg.Content, false),
				slack.MsgOptionTS(messageId),
			)
			if err != nil {
				logger.Error("update message failed", "err", err)
				continue
			}
			originalMsgID = ""
		}
	}
}

func (s *SlackRobot) GetContent(content string) (string, error) {
	if len(content) != 0 {
		return content, nil
	}
	
	if len(s.Event.Message.Files) == 0 {
		return "", errors.New("no content or files found")
	}
	
	file := s.Event.Message.Files[0]
	var bs []byte
	var err error
	
	switch file.Mimetype {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		bs, err = s.downloadSlackFile(file.URLPrivateDownload)
		if err != nil {
			logger.Error("download image failed", "err", err)
			return "", err
		}
		content, err = s.Robot.GetImageContent(bs)
		if err != nil {
			logger.Warn("generate text from image failed", "err", err)
			return "", err
		}
	
	case "audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4":
		// 下载音频
		bs, err = s.downloadSlackFile(file.URLPrivateDownload)
		if err != nil {
			logger.Error("download audio failed", "err", err)
			return "", err
		}
		// 调用语音转文字
		content, err = s.Robot.GetAudioContent(bs)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
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

func (s *SlackRobot) sendModeConfigurationOptions() {
	channelId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	var blocks []slack.Block
	
	switch *conf.BaseConfInfo.Type {
	case param.DeepSeek:
		if *conf.BaseConfInfo.CustomUrl == "" || *conf.BaseConfInfo.CustomUrl == "https://api.deepseek.com/" {
			for k := range param.DeepseekModels {
				btnText := slack.NewTextBlockObject("plain_text", k, false, false)
				btn := slack.NewButtonBlockElement(k, k, btnText)
				btn.Value = k
				actionBlock := slack.NewActionBlock("select_model"+k, btn)
				blocks = append(blocks, actionBlock)
			}
		} else {
			// 直接写死 Deepseek 的几个按钮
			models := []string{
				godeepseek.AzureDeepSeekR1,
				godeepseek.OpenRouterDeepSeekR1,
				godeepseek.OpenRouterDeepSeekR1DistillLlama70B,
				godeepseek.OpenRouterDeepSeekR1DistillLlama8B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen14B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen32B,
				"llama2",
			}
			for _, m := range models {
				val := m
				if m == "llama2" {
					val = param.LLAVA
				}
				btnText := slack.NewTextBlockObject("plain_text", m, false, false)
				btn := slack.NewButtonBlockElement(m, val, btnText)
				btn.Value = val
				actionBlock := slack.NewActionBlock("select_model"+m, btn)
				blocks = append(blocks, actionBlock)
			}
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			btnText := slack.NewTextBlockObject("plain_text", k, false, false)
			btn := slack.NewButtonBlockElement(k, k, btnText)
			btn.Value = k
			actionBlock := slack.NewActionBlock("select_model"+k, btn)
			blocks = append(blocks, actionBlock)
		}
	case param.OpenAi:
		for k := range param.OpenAIModels {
			btnText := slack.NewTextBlockObject("plain_text", k, false, false)
			btn := slack.NewButtonBlockElement(k, k, btnText)
			btn.Value = k
			actionBlock := slack.NewActionBlock("select_model"+k, btn)
			blocks = append(blocks, actionBlock)
		}
	case param.LLAVA:
		btnText := slack.NewTextBlockObject("plain_text", "llama2", false, false)
		btn := slack.NewButtonBlockElement(param.LLAVA, param.LLAVA, btnText)
		btn.Value = param.LLAVA
		actionBlock := slack.NewActionBlock("select_model"+param.LLAVA, btn)
		blocks = append(blocks, actionBlock)
	case param.OpenRouter:
		for k := range param.OpenRouterModelTypes {
			btnText := slack.NewTextBlockObject("plain_text", k, false, false)
			btn := slack.NewButtonBlockElement(k, k, btnText)
			btn.Value = k
			actionBlock := slack.NewActionBlock("select_model"+k, btn)
			blocks = append(blocks, actionBlock)
		}
	case param.Vol:
		for k := range param.VolModels {
			btnText := slack.NewTextBlockObject("plain_text", k, false, false)
			btn := slack.NewButtonBlockElement(k, k, btnText)
			btn.Value = k
			actionBlock := slack.NewActionBlock("select_model"+k, btn)
			blocks = append(blocks, actionBlock)
		}
	}
	
	// 发送或更新消息，包含按钮
	_, _, err := s.Client.PostMessage(
		channelId,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_mode", nil), false),
	)
	if err != nil {
		logger.Warn("send mode config options failed", "err", err)
		return
	}
	
}

func (s *SlackRobot) sendImg() {
	s.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := s.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := s.Prompt
		prompt = utils.ReplaceCommand(prompt, "/photo", s.BotName)
		if prompt == "" {
			logger.Warn("prompt is empty")
			s.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		thinkingMsg := s.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil), msgId, "", nil)
		
		var err error
		var lastImageContent []byte
		if s.Event != nil && s.Event.Message.Files != nil && len(s.Event.Message.Files) > 0 {
			file := s.Event.Message.Files[0]
			lastImageContent, err = s.downloadSlackFile(file.URLPrivateDownload)
			if err != nil {
				logger.Error("download image failed", "err", err)
			}
		}
		if len(lastImageContent) == 0 {
			lastImageContent, err = s.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		var imageUrl string
		var imageContent []byte
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		switch mode {
		case param.Vol:
			imageUrl, totalToken, err = llm.GenerateVolImg(prompt, lastImageContent)
		case param.OpenAi:
			imageContent, totalToken, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
		case param.Gemini:
			imageContent, totalToken, err = llm.GenerateGeminiImg(prompt, lastImageContent)
		default:
			err = fmt.Errorf("unsupported media type: %s", mode)
		}
		
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			s.Client.PostMessage(chatId, slack.MsgOptionText("生成图片失败: "+err.Error(), false))
			return
		}
		
		if len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				return
			}
		}
		
		uploadParams := slack.UploadFileV2Parameters{
			Filename: "image." + utils.DetectImageFormat(imageContent),
			Reader:   bytes.NewReader(imageContent),
			Title:    "image",
			FileSize: len(imageContent),
			Channel:  chatId,
		}
		
		_, err = s.Client.UploadFileV2(uploadParams)
		if err != nil {
			logger.Warn("upload image to slack fail", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if thinkingMsg != "" {
			_, _, err := s.Client.DeleteMessage(chatId, thinkingMsg)
			if err != nil {
				logger.Warn("delete thinking message fail", "err", err)
			}
		}
		
		// 你可以记录数据库
		dataURI := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageContent)
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

func (s *SlackRobot) sendVideo() {
	s.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, userID := s.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := s.Prompt
		prompt = utils.ReplaceCommand(prompt, "/video", s.BotName)
		if prompt == "" {
			logger.Warn("prompt is empty")
			s.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil),
				replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		thinkingMsg := s.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil), replyToMessageID, "", nil)
		
		var err error
		var videoURL string
		var videoContent []byte
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		
		// 调用对应服务生成视频
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			videoURL, totalToken, err = llm.GenerateVolVideo(prompt)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt)
		default:
			err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate video failed", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 发送视频文件消息
		if len(videoURL) != 0 {
			// 通过文件上传接口上传视频URL（下载后上传）
			videoContent, err = utils.DownloadFile(videoURL)
			if err != nil {
				logger.Warn("download video failed", "err", err)
				return
			}
		}
		
		uploadParams := slack.UploadFileV2Parameters{
			Filename: "video." + utils.DetectVideoMimeType(videoContent),
			Reader:   bytes.NewReader(videoContent),
			Title:    "video",
			FileSize: len(videoContent),
			Channel:  chatId,
		}
		
		_, err = s.Client.UploadFileV2(uploadParams)
		if err != nil {
			logger.Warn("upload image to slack fail", "err", err)
			s.Robot.SendMsg(chatId, err.Error(), replyToMessageID, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if thinkingMsg != "" {
			_, _, err := s.Client.DeleteMessage(chatId, thinkingMsg)
			if err != nil {
				logger.Warn("delete thinking message fail", "err", err)
			}
		}
		
		// 记录到数据库
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", utils.DetectVideoMimeType(videoContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userID,
			Question:   prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
}

func (s *SlackRobot) sendHelpConfigurationOptions() {
	chatId, _, _ := s.Robot.GetChatIdAndMsgIdAndUserID()
	
	blocks := []slack.Block{
		slack.NewActionBlock("action_block",
			slack.NewButtonBlockElement("mode", "mode", slack.NewTextBlockObject("plain_text", "mode", false, false)),
			slack.NewButtonBlockElement("clear", "clear", slack.NewTextBlockObject("plain_text", "clear", false, false)),
		),
		slack.NewActionBlock("action_block2",
			slack.NewButtonBlockElement("balance", "balance", slack.NewTextBlockObject("plain_text", "balance", false, false)),
			slack.NewButtonBlockElement("state", "state", slack.NewTextBlockObject("plain_text", "state", false, false)),
		),
		slack.NewActionBlock("action_block3",
			slack.NewButtonBlockElement("retry", "retry", slack.NewTextBlockObject("plain_text", "retry", false, false)),
			slack.NewButtonBlockElement("chat", "chat", slack.NewTextBlockObject("plain_text", "chat", false, false)),
		),
		slack.NewActionBlock("action_block4",
			slack.NewButtonBlockElement("photo", "photo", slack.NewTextBlockObject("plain_text", "photo", false, false)),
			slack.NewButtonBlockElement("video", "video", slack.NewTextBlockObject("plain_text", "video", false, false)),
		),
	}
	
	_, _, err := s.Client.PostMessage(chatId, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		logger.Error("post message failed", "err", err)
	}
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
		logger.Error("open modal failed", "err", err)
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
	
	s.Robot.ExecCmd(s.Command, func() {})
	
}

func (s *SlackRobot) getPrompt() string {
	return s.Prompt
}
