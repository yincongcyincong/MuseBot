package robot

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime/debug"
	"strings"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work/message/request"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	ComWechatApp *work.Work
)

type ComWechatRobot struct {
	Event contract.EventInterface
	Robot *RobotInfo
	App   *work.Work
	
	Command      string
	Prompt       string
	OriginPrompt string
	ImageContent []byte
	VoiceContent []byte
	TextMsg      *serverModel.MessageText
	VoiceMsg     *serverModel.MessageVoice
	ImageMsg     *serverModel.MessageImage
	UserName     string
}

func StartComWechatRobot(ctx context.Context) {
	var err error
	ComWechatApp, err = work.NewWork(&work.UserConfig{
		CorpID:    *conf.BaseConfInfo.ComWechatCorpID,
		AgentID:   utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
		Secret:    *conf.BaseConfInfo.ComWechatSecret,
		Token:     *conf.BaseConfInfo.ComWechatToken,
		AESKey:    *conf.BaseConfInfo.ComWechatEncodingAESKey,
		HttpDebug: false,
		OAuth: work.OAuth{
			Callback: "https://github.com/yincongcyincong/MuseBot",
			Scopes:   nil,
		},
	})
	if err != nil {
		logger.ErrorCtx(ctx, "ComWechatApp init error: ", err)
		return
	}
	
	resp, err := ComWechatApp.Agent.Get(ctx, utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID))
	if err != nil {
		logger.ErrorCtx(ctx, "ComWechatApp get agent error: ", "err", err)
		return
	}
	logger.InfoCtx(ctx, "ComWechatbot", "username", resp.Name)
}

func NewComWechatRobot(event contract.EventInterface) *ComWechatRobot {
	metrics.AppRequestCount.WithLabelValues("com_wechat").Inc()
	c := &ComWechatRobot{
		Event: event,
		App:   ComWechatApp,
	}
	
	switch event.GetMsgType() {
	case models.CALLBACK_MSG_TYPE_TEXT:
		msg := &serverModel.MessageText{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "ComWechatRobot", "err", err)
			return nil
		}
		c.TextMsg = msg
		c.UserName = c.TextMsg.FromUserName
	case models.CALLBACK_MSG_TYPE_IMAGE:
		msg := &serverModel.MessageImage{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "ComWechatRobot", "err", err)
			return nil
		}
		c.ImageMsg = msg
		c.UserName = c.ImageMsg.FromUserName
	case models.CALLBACK_MSG_TYPE_VOICE:
		msg := &serverModel.MessageVoice{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "ComWechatRobot", "err", err)
			return nil
		}
		c.VoiceMsg = msg
		c.UserName = c.VoiceMsg.FromUserName
	}
	
	return c
}

func (c *ComWechatRobot) checkValid() bool {
	chatId, msgId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
	var err error
	if c.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_TEXT {
		c.OriginPrompt = c.TextMsg.Content
		c.Command, c.Prompt = ParseCommand(c.TextMsg.Content)
		logger.InfoCtx(c.Robot.Ctx, "ComWechatRobot msg", "Command", c.Command, "Prompt", c.Prompt)
	}
	
	if c.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_IMAGE {
		c.ImageContent, err = c.getMedia()
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "get media fail", "err", err)
			c.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}
	}
	
	if c.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_VOICE {
		c.VoiceContent, err = c.getMedia()
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "get media fail", "err", err)
			c.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}
		if c.VoiceContent != nil {
			data, err := utils.AmrToOgg(c.VoiceContent)
			if err != nil {
				logger.ErrorCtx(c.Robot.Ctx, "convert amr to wav fail", "err", err)
				c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				
				return false
			}
			c.Prompt, err = c.Robot.GetAudioContent(data)
			if err != nil {
				logger.ErrorCtx(c.Robot.Ctx, "get audio content fail", "err", err)
				c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return false
			}
		}
	}
	
	return true
}

func (c *ComWechatRobot) getMsgContent() string {
	return c.Command
}

func (c *ComWechatRobot) requestLLM(content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorCtx(c.Robot.Ctx, "ComWechatRobot panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		if !strings.Contains(content, "/") && !strings.Contains(content, "$") && c.Prompt == "" {
			c.Prompt = content
		}
		c.Robot.ExecCmd(content, c.sendChatMessage, nil, nil)
	}()
}

func (c *ComWechatRobot) sendImg() {
	c.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(c.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			c.Robot.SendMsg(chatId, i18n.GetMessage("photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		lastImageContent := c.ImageContent
		var err error
		if len(lastImageContent) == 0 && strings.Contains(c.Command, "edit_photo") {
			lastImageContent, err = c.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := c.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = c.sendMedia(imageContent, utils.DetectImageFormat(imageContent), "image")
		if err != nil {
			logger.Warn("send image fail", "err", err)
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		c.Robot.saveRecord(imageContent, lastImageContent, param.ImageRecordType, totalToken)
	})
}

func (c *ComWechatRobot) sendMedia(mediaContent []byte, contentType, sType string) error {
	_, _, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
	
	if sType == "image" {
		fileName := utils.GetAbsPath("data/" + utils.RandomFilename(contentType))
		err := os.WriteFile(fileName, mediaContent, 0666)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "save image fail", "err", err)
			return err
		}
		
		mediaID, err := c.App.Media.UploadTempImage(c.Robot.Ctx, fileName, nil)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "upload image fail", "err", err)
			return err
		}
		
		_, err = c.App.Message.SendImage(c.Robot.Ctx, &request.RequestMessageSendImage{
			RequestMessageSend: request.RequestMessageSend{
				ToUser:                 userId,
				MsgType:                sType,
				AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
				DuplicateCheckInterval: 1800,
			},
			Image: &request.RequestImage{
				MediaID: mediaID.MediaID,
			},
		})
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "send image fail", "err", err)
			return err
		}
	} else {
		fileName := utils.GetAbsPath("data/" + utils.RandomFilename(contentType))
		err := os.WriteFile(fileName, mediaContent, 0666)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "save image fail", "err", err)
			return err
		}
		mediaID, err := c.App.Media.UploadTempVideo(c.Robot.Ctx, fileName, nil)
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "upload image fail", "err", err)
			return err
		}
		
		_, err = c.App.Message.SendVideo(c.Robot.Ctx, &request.RequestMessageSendVideo{
			RequestMessageSend: request.RequestMessageSend{
				ToUser:                 userId,
				MsgType:                sType,
				AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
				DuplicateCheckInterval: 1800,
			},
			Video: &request.RequestVideo{
				MediaID: mediaID.MediaID,
			},
		})
		if err != nil {
			logger.ErrorCtx(c.Robot.Ctx, "send image fail", "err", err)
			return err
		}
	}
	
	return nil
}

func (c *ComWechatRobot) sendVideo() {
	// 检查 prompt
	c.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(c.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			c.Robot.SendMsg(chatId, i18n.GetMessage("photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		videoContent, totalToken, err := c.Robot.CreateVideo(prompt, c.ImageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = c.sendMedia(videoContent, utils.DetectVideoMimeType(videoContent), "video")
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		c.Robot.saveRecord(videoContent, c.ImageContent, param.VideoRecordType, totalToken)
	})
	
}

func (c *ComWechatRobot) sendChatMessage() {
	c.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			c.executeChain()
		} else {
			c.executeLLM()
		}
	})
	
}

func (c *ComWechatRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go c.Robot.ExecChain(c.Prompt, messageChan)
	
	// send response message
	go c.Robot.HandleUpdate(messageChan, "amr")
}

func (c *ComWechatRobot) sendText(messageChan *MsgChan) {
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			c.SendMsg(msg)
		}
	}
	
	if msg != nil {
		c.SendMsg(msg)
	}
	
}

func (c *ComWechatRobot) SendMsg(msg *param.MsgInfo) {
	chatId, _, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
	blocks := utils.ExtractContentBlocks(msg.Content)
	for _, b := range blocks {
		switch b.Type {
		case "text":
			c.Robot.SendMsg(chatId, strings.TrimSpace(b.Content), "", tgbotapi.ModeMarkdown, nil)
		case "video", "image":
			content, err := utils.DownloadFile(b.Media.URL)
			if err != nil {
				logger.ErrorCtx(c.Robot.Ctx, "download file fail", "err", err)
				continue
			}
			contentType := ""
			if b.Type == "video" {
				contentType = utils.DetectVideoMimeType(content)
			} else {
				contentType = utils.DetectImageFormat(content)
			}
			
			err = c.sendMedia(content, contentType, b.Type)
			if err != nil {
				logger.ErrorCtx(c.Robot.Ctx, "send media fail", "err", err)
			}
		}
	}
}

func (c *ComWechatRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go c.Robot.HandleUpdate(messageChan, "amr")
	
	go c.Robot.ExecLLM(c.Prompt, messageChan)
	
}

func (c *ComWechatRobot) getContent(content string) (string, error) {
	
	msgType := c.Event.GetMsgType()
	
	switch msgType {
	case models.CALLBACK_MSG_TYPE_IMAGE:
		return c.Robot.GetImageContent(c.ImageContent, content)
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}

func (c *ComWechatRobot) getPrompt() string {
	return c.Prompt
}

func (c *ComWechatRobot) getPerMsgLen() int {
	return 3500
}

func (c *ComWechatRobot) getMedia() ([]byte, error) {
	mediaId := ""
	if c.ImageMsg != nil {
		mediaId = c.ImageMsg.MediaID
	}
	if c.VoiceMsg != nil {
		mediaId = c.VoiceMsg.MediaID
	}
	
	resp, err := c.App.Media.Get(c.Robot.Ctx, mediaId)
	if err != nil {
		logger.ErrorCtx(c.Robot.Ctx, "get media fail", "err", err)
		return nil, err
	}
	
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(c.Robot.Ctx, "read media fail", "err", err)
		return nil, err
	}
	
	return data, nil
}

func (c *ComWechatRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	_, _, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
	
	fileName := utils.GetAbsPath("data/" + utils.RandomFilename(utils.DetectAudioFormat(voiceContent)))
	err := os.WriteFile(fileName, voiceContent, 0666)
	if err != nil {
		logger.ErrorCtx(c.Robot.Ctx, "save voice fail", "err", err)
		return err
	}
	
	mediaResp, err := c.App.Media.UploadTempVoice(c.Robot.Ctx, fileName, nil)
	if err != nil {
		logger.ErrorCtx(c.Robot.Ctx, "upload voice fail", "err", err)
		return err
	}
	resp, err := c.App.Message.SendVoice(c.Robot.Ctx, &request.RequestMessageSendVoice{
		RequestMessageSend: request.RequestMessageSend{
			ToUser:                 userId,
			MsgType:                "voice",
			AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
			DuplicateCheckInterval: 1800,
		},
		Voice: &request.RequestVoice{
			MediaID: mediaResp.MediaID,
		},
	})
	if err != nil {
		logger.ErrorCtx(c.Robot.Ctx, "send image fail", "err", err, "resp", resp)
		return err
	}
	
	return nil
}

func (c *ComWechatRobot) setCommand(command string) {
	c.Command = command
}

func (c *ComWechatRobot) getCommand() string {
	return c.Command
}

func (c *ComWechatRobot) getUserName() string {
	return c.UserName
}

func (c *ComWechatRobot) setPrompt(prompt string) {
	c.Prompt = prompt
}

func (c *ComWechatRobot) getAudio() []byte {
	return c.VoiceContent
}

func (c *ComWechatRobot) getImage() []byte {
	return c.ImageContent
}
