package robot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"runtime/debug"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type MessageText struct {
	Text string `json:"text"`
}

var (
	cli           *larkws.Client
	BotName       string
	LarkBotClient *lark.Client
)

type LarkRobot struct {
	Message *larkim.P2MessageReceiveV1
	Robot   *RobotInfo
	Client  *lark.Client
	
	Command      string
	Prompt       string
	BotName      string
	ImageContent []byte
	AudioContent []byte
	UserName     string
}

func StartLarkRobot(ctx context.Context) {
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(LarkMessageHandler)
	
	cli = larkws.NewClient(conf.BaseConfInfo.LarkAPPID, conf.BaseConfInfo.LarkAppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
		larkws.WithLogger(logger.Logger),
	)
	
	LarkBotClient = lark.NewClient(conf.BaseConfInfo.LarkAPPID, conf.BaseConfInfo.LarkAppSecret,
		lark.WithHttpClient(utils.GetRobotProxyClient()))
	
	// get bot name
	resp, err := LarkBotClient.Application.Application.Get(ctx, larkapplication.NewGetApplicationReqBuilder().
		AppId(conf.BaseConfInfo.LarkAPPID).Lang("zh_cn").Build())
	if err != nil || !resp.Success() {
		logger.ErrorCtx(ctx, "get robot name error", "error", err, "resp", resp)
		return
	}
	BotName = larkcore.StringValue(resp.Data.App.AppName)
	logger.Info("LarkBot Info", "username", BotName)
	
	err = cli.Start(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "start larkbot fail", "err", err)
	}
}

func NewLarkRobot(message *larkim.P2MessageReceiveV1) *LarkRobot {
	metrics.AppRequestCount.WithLabelValues("lark").Inc()
	return &LarkRobot{
		Message: message,
		Client:  LarkBotClient,
		BotName: BotName,
	}
}

func LarkMessageHandler(ctx context.Context, message *larkim.P2MessageReceiveV1) error {
	l := NewLarkRobot(message)
	l.Robot = NewRobot(WithRobot(l), WithContext(ctx))
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(ctx, "exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		userInfo, err := LarkBotClient.Contact.V3.User.Get(l.Robot.Ctx, larkcontact.NewGetUserReqBuilder().
			UserId(*message.Event.Sender.SenderId.UserId).UserIdType("user_id").Build())
		if err != nil || userInfo.Code != 0 {
			logger.ErrorCtx(ctx, "get user info error", "err", err, "user_info", userInfo)
		} else {
			l.UserName = *userInfo.Data.User.Name
		}
		
		l.Robot.Exec()
	}()
	
	return nil
}

func (l *LarkRobot) checkValid() bool {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	// group need to at bot
	atBot, err := l.GetMessageContent()
	if err != nil {
		logger.ErrorCtx(l.Robot.Ctx, "get message content error", "err", err)
		l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return false
	}
	if larkcore.StringValue(l.Message.Event.Message.ChatType) == "group" {
		if !atBot {
			logger.Warn("no at bot")
			return false
		}
	}
	
	return true
}

func (l *LarkRobot) getMsgContent() string {
	return l.Command
}

func (l *LarkRobot) requestLLM(content string) {
	if !strings.Contains(content, "/") && !strings.Contains(content, "$") && l.Prompt == "" {
		l.Prompt = content
	}
	l.Robot.ExecCmd(content, l.sendChatMessage, nil, nil)
}

func (l *LarkRobot) sendImg() {
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(l.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			l.Robot.SendMsg(chatId, i18n.GetMessage("photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		lastImageContent := l.ImageContent
		var err error
		if len(lastImageContent) == 0 && strings.Contains(l.Command, "edit_photo") {
			lastImageContent, err = l.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := l.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = l.sendMedia(imageContent, utils.DetectImageFormat(imageContent), "image")
		if err != nil {
			logger.Warn("send image fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		l.Robot.saveRecord(imageContent, lastImageContent, param.ImageRecordType, totalToken)
	})
}

func (l *LarkRobot) sendMedia(media []byte, contentType, sType string) error {
	postContent := make([]larkim.MessagePostElement, 0)
	chatId, _, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	if sType == "image" {
		imageKey, err := l.getImageInfo(media)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "create image fail", "err", err)
			return err
		}
		
		postContent = append(postContent, &larkim.MessagePostImage{
			ImageKey: imageKey,
		})
		
	} else {
		fileKey, err := l.getVideoInfo(media)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "get image info fail", "err", err)
			return err
		}
		
		postContent = append(postContent, &larkim.MessagePostMedia{
			FileKey: fileKey,
		})
	}
	
	msgContent, _ := larkim.NewMessagePost().ZhCn(larkim.NewMessagePostContent().AppendContent(postContent).Build()).Build()
	res, err := l.Client.Im.Message.Create(l.Robot.Ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypePost).
			ReceiveId(chatId).
			Content(msgContent).
			Build()).
		Build())
	if err != nil || !res.Success() {
		logger.Warn("send message fail", "err", err, "resp", res)
		return err
	}
	
	return nil
}

func (l *LarkRobot) sendVideo() {
	// 检查 prompt
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(l.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			l.Robot.SendMsg(chatId, i18n.GetMessage("video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		videoContent, totalToken, err := l.Robot.CreateVideo(prompt, l.ImageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = l.sendMedia(videoContent, utils.DetectVideoMimeType(videoContent), "video")
		if err != nil {
			logger.Warn("send video fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		l.Robot.saveRecord(videoContent, l.ImageContent, param.VideoRecordType, totalToken)
	})
	
}

func (l *LarkRobot) sendChatMessage() {
	l.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			l.executeChain()
		} else {
			l.executeLLM()
		}
	})
	
}

func (l *LarkRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go l.Robot.ExecChain(l.Prompt, messageChan)
	
	go l.Robot.HandleUpdate(messageChan, "opus")
}

func (l *LarkRobot) sendTextStream(messageChan *MsgChan) {
	var msg *param.MsgInfo
	chatId, messageId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	for msg = range messageChan.NormalMessageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from llm!"
		}
		
		if msg.MsgId == "" {
			msgId := l.Robot.SendMsg(chatId, msg.Content, messageId, tgbotapi.ModeMarkdown, nil)
			msg.MsgId = msgId
		} else {
			
			resp, err := l.Client.Im.Message.Update(l.Robot.Ctx, larkim.NewUpdateMessageReqBuilder().
				MessageId(msg.MsgId).
				Body(larkim.NewUpdateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					Content(GetMarkdownContent(msg.Content)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.Warn("send message fail", "err", err, "resp", resp)
				continue
			}
		}
	}
}

func (l *LarkRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go l.Robot.HandleUpdate(messageChan, "opus")
	
	go l.Robot.ExecLLM(l.Prompt, messageChan)
	
}

func GetMarkdownContent(content string) string {
	markdownMsg, _ := larkim.NewMessagePost().ZhCn(larkim.NewMessagePostContent().AppendContent(
		[]larkim.MessagePostElement{
			&MessagePostMarkdown{
				Text: content,
			},
		}).Build()).Build()
	
	return markdownMsg
}

type MessagePostMarkdown struct {
	Text string `json:"text,omitempty"`
}

func (m *MessagePostMarkdown) Tag() string {
	return "md"
}

func (m *MessagePostMarkdown) IsPost() {
}

func (m *MessagePostMarkdown) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"tag":  "md",
		"text": m.Text,
	}
	return json.Marshal(data)
}

type MessagePostContent struct {
	Title   string                 `json:"title"`
	Content [][]MessagePostElement `json:"content"`
}

type MessagePostElement struct {
	Tag      string `json:"tag"`
	Text     string `json:"text"`
	ImageKey string `json:"image_key"`
	UserName string `json:"user_name"`
}

func (l *LarkRobot) GetMessageContent() (bool, error) {
	_, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	msgType := larkcore.StringValue(l.Message.Event.Message.MessageType)
	botShowName := ""
	if msgType == larkim.MsgTypeText {
		textMsg := new(MessageText)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), textMsg)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "unmarshal text message error", "error", err)
			return false, err
		}
		l.Command, l.Prompt = ParseCommand(textMsg.Text)
		for _, at := range l.Message.Event.Message.Mentions {
			if larkcore.StringValue(at.Name) == l.BotName {
				botShowName = larkcore.StringValue(at.Key)
				break
			}
		}
		
		l.Prompt = strings.ReplaceAll(l.Prompt, "@"+botShowName, "")
		for _, at := range l.Message.Event.Message.Mentions {
			if larkcore.StringValue(at.Name) == l.BotName {
				botShowName = larkcore.StringValue(at.Name)
				break
			}
		}
	} else if msgType == larkim.MsgTypePost {
		postMsg := new(MessagePostContent)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), postMsg)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "unmarshal text message error", "error", err)
			return false, err
		}
		
		for _, msgPostContents := range postMsg.Content {
			for _, msgPostContent := range msgPostContents {
				switch msgPostContent.Tag {
				case "text":
					command, prompt := ParseCommand(msgPostContent.Text)
					if command != "" {
						l.Command = command
					}
					if prompt != "" {
						l.Prompt = prompt
					}
				case "img":
					resp, err := l.Client.Im.V1.MessageResource.Get(l.Robot.Ctx,
						larkim.NewGetMessageResourceReqBuilder().
							MessageId(msgId).
							FileKey(msgPostContent.ImageKey).
							Type("image").
							Build())
					if err != nil || !resp.Success() {
						logger.ErrorCtx(l.Robot.Ctx, "get image failed", "err", err, "resp", resp)
						return false, err
					}
					
					bs, err := io.ReadAll(resp.File)
					if err != nil {
						logger.ErrorCtx(l.Robot.Ctx, "read image failed", "err", err)
						return false, err
					}
					l.ImageContent = bs
				case "at":
					if l.BotName == msgPostContent.UserName {
						botShowName = msgPostContent.UserName
					}
					
				}
			}
		}
	} else if msgType == larkim.MsgTypeAudio {
		msgAudio := new(larkim.MessageAudio)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), msgAudio)
		if err != nil {
			logger.Warn("unmarshal message audio failed", "err", err)
			return false, err
		}
		resp, err := l.Client.Im.V1.MessageResource.Get(l.Robot.Ctx,
			larkim.NewGetMessageResourceReqBuilder().
				MessageId(msgId).
				FileKey(msgAudio.FileKey).
				Type("file").
				Build())
		if err != nil || !resp.Success() {
			logger.ErrorCtx(l.Robot.Ctx, "get image failed", "err", err, "resp", resp)
			return false, err
		}
		
		bs, err := io.ReadAll(resp.File)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "read image failed", "err", err)
			return false, err
		}
		l.AudioContent = bs
		
		l.Prompt, err = l.Robot.GetAudioContent(bs)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return false, err
		}
	} else if msgType == larkim.MsgTypeImage {
		msgImage := new(larkim.MessageImage)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), msgImage)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "unmarshal message image failed", "err", err)
			return false, err
		}
		
		resp, err := l.Client.Im.V1.MessageResource.Get(l.Robot.Ctx,
			larkim.NewGetMessageResourceReqBuilder().
				MessageId(msgId).
				FileKey(msgImage.ImageKey).
				Type("image").
				Build())
		if err != nil || !resp.Success() {
			logger.ErrorCtx(l.Robot.Ctx, "get image failed", "err", err, "resp", resp)
			return false, err
		}
		
		l.ImageContent, err = io.ReadAll(resp.File)
		if err != nil {
			logger.ErrorCtx(l.Robot.Ctx, "read image failed", "err", err)
			return false, err
		}
	}
	
	l.Prompt = strings.ReplaceAll(l.Prompt, "@"+l.BotName, "")
	return botShowName == l.BotName, nil
}

func (l *LarkRobot) getPrompt() string {
	return l.Prompt
}

func (l *LarkRobot) getPerMsgLen() int {
	return 4500
}

func (l *LarkRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	_, messageId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	resp, err := l.Client.Im.V1.File.Create(l.Robot.Ctx, larkim.NewCreateFileReqBuilder().
		Body(larkim.NewCreateFileReqBodyBuilder().
			FileType("opus").
			FileName(utils.RandomFilename(".ogg")).
			Duration(duration).
			File(bytes.NewReader(voiceContent)).
			Build()).
		Build())
	if err != nil || !resp.Success() {
		logger.Warn("create voice fail", "err", err, "resp", resp)
		return errors.New("request upload file fail")
	}
	
	audio := larkim.MessageAudio{
		FileKey: *resp.Data.FileKey,
	}
	msgContent, _ := audio.String()
	
	updateRes, err := l.Client.Im.Message.Reply(l.Robot.Ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(messageId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeAudio).
			Content(msgContent).
			Build()).
		Build())
	if err != nil || !updateRes.Success() {
		logger.Warn("send message fail", "err", err, "resp", updateRes)
		return errors.New("send voice fail")
	}
	
	return err
}

func (l *LarkRobot) setCommand(command string) {
	l.Command = command
}

func (l *LarkRobot) getCommand() string {
	return l.Command
}

func (l *LarkRobot) getUserName() string {
	return l.UserName
}

func (l *LarkRobot) getVideoInfo(videoContent []byte) (string, error) {
	format := utils.DetectVideoMimeType(videoContent)
	resp, err := l.Client.Im.V1.File.Create(l.Robot.Ctx, larkim.NewCreateFileReqBuilder().
		Body(larkim.NewCreateFileReqBodyBuilder().
			FileType(format).
			FileName(utils.RandomFilename(format)).
			Duration(conf.VideoConfInfo.Duration).
			File(bytes.NewReader(videoContent)).
			Build()).
		Build())
	if err != nil || !resp.Success() {
		logger.ErrorCtx(l.Robot.Ctx, "create image fail", "err", err, "resp", resp)
		return "", err
	}
	
	return larkcore.StringValue(resp.Data.FileKey), nil
}

func (l *LarkRobot) getImageInfo(imageContent []byte) (string, error) {
	resp, err := l.Client.Im.V1.Image.Create(l.Robot.Ctx, larkim.NewCreateImageReqBuilder().
		Body(larkim.NewCreateImageReqBodyBuilder().
			ImageType("message").
			Image(bytes.NewReader(imageContent)).
			Build()).
		Build())
	if err != nil || !resp.Success() {
		logger.Warn("create image fail", "err", err, "resp", resp)
		return "", err
	}
	
	return larkcore.StringValue(resp.Data.ImageKey), nil
}

func (l *LarkRobot) setPrompt(prompt string) {
	l.Prompt = prompt
}

func (l *LarkRobot) getAudio() []byte {
	return l.AudioContent
}

func (l *LarkRobot) getImage() []byte {
	return l.ImageContent
}

func (l *LarkRobot) setImage(image []byte) {
	l.ImageContent = image
}
