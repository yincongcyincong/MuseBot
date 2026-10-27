package robot

import (
	"context"
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

	OriginPrompt string
	VoiceContent []byte
	TextMsg      *serverModel.MessageText
	VoiceMsg     *serverModel.MessageVoice
	ImageMsg     *serverModel.MessageImage
	baseRobot
}

func StartComWechatRobot(ctx context.Context) {
	var err error
	ComWechatApp, err = work.NewWork(&work.UserConfig{
		CorpID:    conf.BaseConfInfo.ComWechatCorpID,
		AgentID:   utils.ParseInt(conf.BaseConfInfo.ComWechatAgentID),
		Secret:    conf.BaseConfInfo.ComWechatSecret,
		Token:     conf.BaseConfInfo.ComWechatToken,
		AESKey:    conf.BaseConfInfo.ComWechatEncodingAESKey,
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

	resp, err := ComWechatApp.Agent.Get(ctx, utils.ParseInt(conf.BaseConfInfo.ComWechatAgentID))
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
		c.Robot.generateMedia(c, mediaImage)
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
				AgentID:                utils.ParseInt(conf.BaseConfInfo.ComWechatAgentID),
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
				AgentID:                utils.ParseInt(conf.BaseConfInfo.ComWechatAgentID),
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
	c.Robot.TalkingPreCheck(func() {
		c.Robot.generateMedia(c, mediaVideo)
	})
}

func (c *ComWechatRobot) sendChatMessage() {
	c.Robot.TalkingPreCheck(func() {
		c.executeLLM()
	})

}

func (c *ComWechatRobot) executeLLM() {
	c.Robot.runLLM(c, "amr", false)
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
			AgentID:                utils.ParseInt(conf.BaseConfInfo.ComWechatAgentID),
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

func init() {
	RegisterChannel(Channel{
		Name: param.ComWechat,
		MissingCfg: func() []string {
			var missing []string
			missing = required(missing, "COM_WECHAT_SECRET", conf.BaseConfInfo.ComWechatSecret)
			missing = required(missing, "COM_WECHAT_AGENT_ID", conf.BaseConfInfo.ComWechatAgentID)
			return required(missing, "COM_WECHAT_ENCODING_AES_KEY", conf.BaseConfInfo.ComWechatEncodingAESKey)
		},
		Start: StartComWechatRobot,
	})
}
