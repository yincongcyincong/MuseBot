package robot

import (
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	OfficialAccountApp *officialAccount.OfficialAccount
)

const (
	msgHandling = 1
	msgFinished = 2
	msgSent     = 3

	msgChangePhoto    = 10
	msgRecognizePhoto = 11
	msgSaveVoice      = 12
)

type WechatRobot struct {
	Event contract.EventInterface
	Robot *RobotInfo
	App   *officialAccount.OfficialAccount

	OriginPrompt string
	VoiceContent []byte

	TextMsg  *serverModel.MessageText
	VoiceMsg *serverModel.MessageVoice
	ImageMsg *serverModel.MessageImage
	baseRobot
}

func StartWechatRobot() {
	var err error
	OfficialAccountApp, err = officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  conf.BaseConfInfo.WechatAppID,
		Secret: conf.BaseConfInfo.WechatAppSecret,
		Token:  conf.BaseConfInfo.WechatToken,
		AESKey: conf.BaseConfInfo.WechatEncodingAESKey,
		Log: officialAccount.Log{
			Level:  "info",
			File:   "./wechat/info.log",
			Error:  "./wechat/info.log",
			Stdout: true,
		},
		HttpDebug: false,
		Debug:     false,
	})
	if err != nil {
		logger.Error("Wechat init error: ", err)
		return
	}

	go deleteMsgMapData()
}

func deleteMsgMapData() {
	ticker := time.NewTicker(time.Minute) // 每分钟执行一次
	defer ticker.Stop()

	for {
		<-ticker.C
		TencentMsgMap.Range(func(key, value interface{}) bool {
			msg, ok := value.(*TencentWechatMessage)
			if !ok {
				return true
			}
			if time.Now().Sub(msg.StartTime) > 5*time.Minute {
				logger.Info("msg deleted", "msgId", key)
				TencentMsgMap.Delete(key)
			}
			return true
		})
	}
}

func NewWechatRobot(event contract.EventInterface) (*WechatRobot, bool) {
	metrics.AppRequestCount.WithLabelValues("wechat").Inc()
	w := &WechatRobot{
		Event: event,
		App:   OfficialAccountApp,
	}
	msgId := ""
	switch event.GetMsgType() {
	case models.CALLBACK_MSG_TYPE_TEXT:
		msg := &serverModel.MessageText{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil, false
		}
		w.TextMsg = msg
		msgId = msg.MsgID
	case models.CALLBACK_MSG_TYPE_IMAGE:
		msg := &serverModel.MessageImage{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil, false
		}
		w.ImageMsg = msg
		msgId = msg.MsgID
	case models.CALLBACK_MSG_TYPE_VOICE:
		msg := &serverModel.MessageVoice{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil, false
		}
		w.VoiceMsg = msg
		msgId = msg.MsgID
	}

	if _, ok := TencentMsgMap.Load(msgId); !ok {
		TencentMsgMap.Store(msgId, &TencentWechatMessage{
			Status:    msgHandling,
			Msg:       "",
			StartTime: time.Now(),
		})

		w.Robot = NewRobot(WithRobot(w))
		return w, true
	}

	logger.InfoCtx(w.Robot.Ctx, "msg still handling", "msgId", msgId)
	w.Robot = NewRobot(WithRobot(w))
	return w, false
}

func (w *WechatRobot) checkValid() bool {
	chatId, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	var err error
	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_TEXT {
		w.OriginPrompt = w.TextMsg.Content
		w.Command, w.Prompt = ParseCommand(w.TextMsg.Content)
		logger.InfoCtx(w.Robot.Ctx, "WechatRobot msg", "Command", w.Command, "Prompt", w.Prompt, "msgId", w.TextMsg.MsgID)
	}

	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_IMAGE {
		w.ImageContent, err = w.getMedia()
		if err != nil {
			logger.Error("read media fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}

	}

	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_VOICE {
		w.VoiceContent, err = w.getMedia()
		if err != nil {
			logger.Error("read media fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}

		if w.VoiceContent != nil {
			data, err := utils.AmrToOgg(w.VoiceContent)
			if err != nil {
				logger.Error("convert amr to wav fail", "err", err)
				w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return false
			}
			w.Prompt, err = w.Robot.GetAudioContent(data)
			if err != nil {
				logger.WarnCtx(w.Robot.Ctx, "convert audio to text fail", "err", err)
				w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return false
			}
		}
	}

	return true
}

func (w *WechatRobot) getMsgContent() string {
	return w.Command
}

func (w *WechatRobot) requestLLM(content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("WechatRobot panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		if !strings.Contains(content, "/") && !strings.Contains(content, "$") && w.Prompt == "" {
			w.Prompt = content
		}
		w.Robot.ExecCmd(content, w.sendChatMessage, nil, nil)
	}()
}

func (w *WechatRobot) sendImg() {
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
		if !conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate image")
			w.Robot.SendMsg(chatId, "only wechat_active is true can generate image", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}

		w.Robot.generateMedia(w, mediaImage)
	})
}

func (w *WechatRobot) sendMedia(mediaContent []byte, contentType, sType string) error {
	fileName := utils.GetAbsPath("data/" + utils.RandomFilename(contentType))
	err := os.WriteFile(fileName, mediaContent, 0666)
	if err != nil {
		logger.Error("save image fail", "err", err)
		return err
	}

	mediaResp, err := w.App.Media.UploadImage(w.Robot.Ctx, fileName)
	if err != nil {
		logger.Error("upload image fail", "err", err)
		return err
	}
	resp, err := w.App.CustomerService.Message(w.Robot.Ctx, messages.NewMedia(mediaResp.MediaID, sType, nil)).
		SetTo(w.Event.GetFromUserName()).SetBy(w.Event.GetToUserName()).Send(w.Robot.Ctx)
	if err != nil {
		logger.Error("send image fail", "err", err, "resp", resp)
		return err
	}

	return nil
}

func (w *WechatRobot) sendVideo() {
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
		if !conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate video")
			w.Robot.SendMsg(chatId, "only wechat_active is true can generate video", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}

		w.Robot.generateMedia(w, mediaVideo)
	})
}

func (w *WechatRobot) sendChatMessage() {
	w.Robot.TalkingPreCheck(func() {
		w.executeLLM()
	})

}

func (w *WechatRobot) executeLLM() {
	w.Robot.runLLM(w, "amr", false)
}

func (w *WechatRobot) GetLLMContent() interface{} {
	_, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	for i := 0; i < 15; i++ {
		if msgInfo, ok := TencentMsgMap.Load(msgId); ok {
			wechatMsg := msgInfo.(*TencentWechatMessage)
			if wechatMsg.Status != msgHandling && wechatMsg.Status != msgChangePhoto && wechatMsg.Status != msgSaveVoice {
				return wechatMsg.Msg
			}
		}
		time.Sleep(1 * time.Second)
	}

	return ""
}

func WechatMsgSent(msgId string) {
	if msgInfo, ok := TencentMsgMap.Load(msgId); ok {
		wechatMsg := msgInfo.(*TencentWechatMessage)
		wechatMsg.Status = msgSent
		TencentMsgMap.Store(msgId, wechatMsg)
	}
}

func (w *WechatRobot) getPerMsgLen() int {
	return 1800
}

func (w *WechatRobot) getMedia() ([]byte, error) {
	mediaId := ""
	if w.ImageMsg != nil {
		mediaId = w.ImageMsg.MediaID
	}
	if w.VoiceMsg != nil {
		mediaId = w.VoiceMsg.MediaID
	}

	resp, err := w.App.Media.Get(w.Robot.Ctx, mediaId)
	if err != nil {
		logger.Error("get media fail", "err", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("read media fail", "err", err)
		return nil, err
	}

	return data, nil
}

func (w *WechatRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	fileName := utils.GetAbsPath("data/" + utils.RandomFilename(utils.DetectAudioFormat(voiceContent)))
	err := os.WriteFile(fileName, voiceContent, 0666)
	if err != nil {
		logger.Error("save voice fail", "err", err)
		return err
	}

	mediaResp, err := w.App.Media.UploadVoice(w.Robot.Ctx, fileName)
	if err != nil {
		logger.Error("upload voice fail", "err", err)
		return err
	}
	resp, err := w.App.CustomerService.Message(w.Robot.Ctx, messages.NewMedia(mediaResp.MediaID, "voice", nil)).
		SetTo(w.Event.GetFromUserName()).SetBy(w.Event.GetToUserName()).Send(w.Robot.Ctx)
	if err != nil {
		logger.Error("send voice fail", "err", err, "resp", resp)
		return err
	}

	return nil
}
