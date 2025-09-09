package robot

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
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
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	OriginPrompt string
	ImageContent []byte
	
	TextMsg  *serverModel.MessageText
	VoiceMsg *serverModel.MessageVoice
	ImageMsg *serverModel.MessageImage
}

func StartWechatRobot() {
	var err error
	OfficialAccountApp, err = officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  *conf.BaseConfInfo.WechatAppID,
		Secret: *conf.BaseConfInfo.WechatAppSecret,
		Token:  *conf.BaseConfInfo.WechatToken,
		AESKey: *conf.BaseConfInfo.WechatEncodingAESKey,
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	
	w := &WechatRobot{
		Event:  event,
		App:    OfficialAccountApp,
		Ctx:    ctx,
		Cancel: cancel,
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
		
		w.Robot = NewRobot(WithRobot(w), WithTencentRobot(w))
		return w, true
	}
	
	logger.Info("msg still handling", "msgId", msgId)
	w.Robot = NewRobot(WithRobot(w))
	return w, false
}

func (w *WechatRobot) checkValid() bool {
	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_TEXT {
		w.OriginPrompt = w.TextMsg.Content
		w.Command, w.Prompt = ParseCommand(w.TextMsg.Content)
		logger.Info("WechatRobot msg", "Command", w.Command, "Prompt", w.Prompt, "msgId", w.TextMsg.MsgID)
	}
	
	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_IMAGE {
		_, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
			if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
				if msgInfo.Status == msgChangePhoto || msgInfo.Status == msgRecognizePhoto {
					logger.Info("WechatRobot handle photo msg", "msgId", msgId, "userId", userId)
					w.passiveExecCmd()
					return false
				}
			}
		}
	}
	
	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_VOICE {
		_, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
			if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
				if msgInfo.Status == msgSaveVoice {
					logger.Info("WechatRobot handle voice msg", "msgId", msgId, "userId", userId)
					w.passiveExecCmd()
					return false
				}
			}
		}
	}
	
	return true
}

func (w *WechatRobot) getMsgContent() string {
	return w.Command
}

func (w *WechatRobot) requestLLMAndResp(content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("WechatRobot panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		if !strings.Contains(content, "/") && w.Prompt == "" {
			w.Prompt = content
		}
		w.Robot.ExecCmd(content, w.sendChatMessage)
	}()
}

func (w *WechatRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "help_text", nil), msgId, tgbotapi.ModeMarkdown, nil)
}

func (w *WechatRobot) sendModeConfigurationOptions() {
	chatId, msgId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(w.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			w.Robot.handleModeUpdate(prompt)
		}
		return
	}
	
	var modelList []string
	
	switch *conf.BaseConfInfo.Type {
	case param.DeepSeek:
		if *conf.BaseConfInfo.CustomUrl == "" || *conf.BaseConfInfo.CustomUrl == "https://api.deepseek.com/" {
			for k := range param.DeepseekModels {
				modelList = append(modelList, k)
			}
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			modelList = append(modelList, k)
		}
	case param.OpenAi:
		for k := range param.OpenAIModels {
			modelList = append(modelList, k)
		}
	case param.OpenRouter, param.AI302, param.Ollama:
		if w.Prompt != "" {
			w.Robot.handleModeUpdate(w.Prompt)
			return
		}
		switch *conf.BaseConfInfo.Type {
		case param.AI302:
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://ollama.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		}
		
		return
	case param.Vol:
		for k := range param.VolModels {
			modelList = append(modelList, k)
		}
	}
	totalContent := ""
	for _, model := range modelList {
		totalContent += fmt.Sprintf(`%s

`, model)
	}
	
	w.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (w *WechatRobot) sendImg() {
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		if !*conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate image")
			w.Robot.SendMsg(chatId, "only wechat_active is true can generate image", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		prompt := strings.TrimSpace(w.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		lastImageContent := w.ImageContent
		var err error
		if len(lastImageContent) == 0 && strings.Contains(w.Command, "edit_photo") {
			lastImageContent, err = w.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := w.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.Error("create photo fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		format := utils.DetectImageFormat(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		
		fileName := utils.GetAbsPath("data/" + utils.RandomFilename(format))
		err = os.WriteFile(fileName, imageContent, 0666)
		if err != nil {
			logger.Error("save image fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		mediaResp, err := w.App.Media.UploadImage(w.Ctx, fileName)
		if err != nil {
			logger.Error("upload image fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		resp, err := w.App.CustomerService.Message(w.Ctx, messages.NewMedia(mediaResp.MediaID, "image", nil)).
			SetTo(w.Event.GetFromUserName()).SetBy(w.Event.GetToUserName()).Send(w.Ctx)
		if err != nil {
			logger.Error("send image fail", "err", err, "resp", resp)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// save data record
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   w.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
}

func (w *WechatRobot) sendVideo() {
	// 检查 prompt
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		if !*conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate video")
			w.Robot.SendMsg(chatId, "only wechat_active is true can generate video", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		prompt := strings.TrimSpace(w.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		videoContent, totalToken, err := w.Robot.CreateVideo(prompt, w.ImageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		format := utils.DetectVideoMimeType(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		fileName := utils.GetAbsPath("data/" + utils.RandomFilename(format))
		err = os.WriteFile(fileName, videoContent, 0666)
		if err != nil {
			logger.Error("save image fail", "err", err)
			return
		}
		mediaResp, err := w.App.Media.UploadVideo(w.Ctx, fileName)
		if err != nil {
			logger.Error("upload image fail", "err", err)
			return
		}
		
		resp, err := w.App.CustomerService.Message(w.Ctx, messages.NewMedia(mediaResp.MediaID, "video", nil)).
			SetTo(w.Event.GetFromUserName()).Send(w.Ctx)
		if err != nil {
			logger.Error("send image fail", "err", err, "resp", resp)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   w.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
	
}

func (w *WechatRobot) sendChatMessage() {
	w.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			w.executeChain()
		} else {
			w.executeLLM()
		}
	})
	
}

func (w *WechatRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go w.Robot.ExecChain(w.Prompt, messageChan)
	
	go w.Robot.HandleUpdate(messageChan, "amr")
}

func (w *WechatRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go w.Robot.HandleUpdate(messageChan, "amr")
	
	go w.Robot.ExecLLM(w.Prompt, messageChan)
	
}

func (w *WechatRobot) getContent(content string) (string, error) {
	msgType := w.Event.GetMsgType()
	
	switch msgType {
	case models.CALLBACK_MSG_TYPE_IMAGE:
		data, err := w.getMedia()
		if err != nil {
			return "", err
		}
		return w.Robot.GetImageContent(data, content)
	
	case models.CALLBACK_MSG_TYPE_VOICE:
		data, err := w.getMedia()
		if err != nil {
			logger.Error("read media fail", "err", err)
			return "", err
		}
		
		data, err = utils.AmrToOgg(data)
		if err != nil {
			logger.Error("convert amr to wav fail", "err", err)
			return "", err
		}
		return w.Robot.GetAudioContent(data)
		
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}

func (w *WechatRobot) getPrompt() string {
	return w.Prompt
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

func (w *WechatRobot) passiveExecCmd() {
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		if !*conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate image")
			w.Robot.SendMsg(chatId, "only wechat_active is true can generate image", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if w.TextMsg != nil {
			status := msgChangePhoto
			switch w.Command {
			case "/change_photo", "change_photo":
				status = msgChangePhoto
			case "/rec_photo", "rec_photo":
				status = msgRecognizePhoto
			case "/save_voice", "save_voice":
				status = msgSaveVoice
			}
			TencentMsgMap.Store(userId, &TencentWechatMessage{
				Msg:       w.Prompt,
				Status:    status,
				StartTime: time.Now(),
			})
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_pre_prompt_success", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if w.ImageMsg != nil {
			if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
				if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
					switch msgInfo.Status {
					case msgChangePhoto:
						data, err := w.getMedia()
						if err != nil {
							logger.Error("read media fail", "err", err)
							w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						w.Prompt = msgInfo.Msg
						w.ImageContent = data
						w.sendImg()
					case msgRecognizePhoto:
						w.Prompt = msgInfo.Msg
						w.executeLLM()
					}
					
					TencentMsgMap.Delete(userId)
				}
			}
		}
		
		if w.VoiceMsg != nil {
			if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
				if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
					switch msgInfo.Status {
					case msgSaveVoice:
						data, err := w.getMedia()
						if err != nil {
							logger.Error("read media fail", "err", err)
							w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						data, err = utils.AmrToMp3(data)
						if err != nil {
							logger.Error("convert amr to wav fail", "err", err)
							w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						
						fileName := utils.GetAbsPath("data/" + utils.RandomFilename(utils.DetectAudioFormat(data)))
						err = os.WriteFile(fileName, data, 0666)
						if err != nil {
							logger.Error("save image fail", "err", err)
							w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						
						w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "save_audio_success", map[string]interface{}{
							"filename": fileName,
						}), msgId, tgbotapi.ModeMarkdown, nil)
						
					}
					
					TencentMsgMap.Delete(userId)
				}
			}
		}
	})
}

func (w *WechatRobot) getMedia() ([]byte, error) {
	mediaId := ""
	if w.ImageMsg != nil {
		mediaId = w.ImageMsg.MediaID
	}
	if w.VoiceMsg != nil {
		mediaId = w.VoiceMsg.MediaID
	}
	
	resp, err := w.App.Media.Get(w.Ctx, mediaId)
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

func (w *WechatRobot) sendText(messageChan *MsgChan) {
	chatId, messageId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			w.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		msg = new(param.MsgInfo)
		return
	}
	
	w.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
}

func (w *WechatRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	fileName := utils.GetAbsPath("data/" + utils.RandomFilename(utils.DetectAudioFormat(voiceContent)))
	err := os.WriteFile(fileName, voiceContent, 0666)
	if err != nil {
		logger.Error("save voice fail", "err", err)
		return err
	}
	
	mediaResp, err := w.App.Media.UploadVoice(w.Ctx, fileName)
	if err != nil {
		logger.Error("upload voice fail", "err", err)
		return err
	}
	resp, err := w.App.CustomerService.Message(w.Ctx, messages.NewMedia(mediaResp.MediaID, "voice", nil)).
		SetTo(w.Event.GetFromUserName()).SetBy(w.Event.GetToUserName()).Send(w.Ctx)
	if err != nil {
		logger.Error("send voice fail", "err", err, "resp", resp)
		return err
	}
	
	return nil
}
