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
	"sync"
	"time"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
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

var (
	OfficialAccountApp *officialAccount.OfficialAccount
	WechatMsgMap       sync.Map
)

const (
	msgHandling = 1
	msgFinished = 2
	msgSent     = 3
)

type WechatMessage struct {
	Msg       interface{}
	Status    int
	StartTime time.Time
}

type WechatRobot struct {
	Event contract.EventInterface
	Robot *RobotInfo
	App   *officialAccount.OfficialAccount
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	OriginPrompt string
	TextMsg      *serverModel.MessageText
	VoiceMsg     *serverModel.MessageVoice
	ImageMsg     *serverModel.MessageImage
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
		WechatMsgMap.Range(func(key, value interface{}) bool {
			msg, ok := value.(*WechatMessage)
			if !ok {
				return true
			}
			if time.Now().Sub(msg.StartTime) > 10*time.Minute {
				logger.Info("msg deleted", "msgId", key)
				WechatMsgMap.Delete(key)
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
	
	if _, ok := WechatMsgMap.Load(msgId); !ok {
		WechatMsgMap.Store(msgId, &WechatMessage{
			Status:    msgHandling,
			Msg:       "",
			StartTime: time.Now(),
		})
		
		w.Robot = NewRobot(WithRobot(w))
		return w, true
	}
	
	logger.Info("msg still handling", "msgId", msgId)
	w.Robot = NewRobot(WithRobot(w))
	return w, false
}

func (w *WechatRobot) checkValid() bool {
	if w.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_TEXT {
		msg := serverModel.MessageText{}
		err := w.Event.ReadMessage(&msg)
		if err != nil {
			logger.Error("WechatRobot", "err", err)
			return false
		}
		w.OriginPrompt = msg.Content
		w.Command, w.Prompt = ParseCommand(msg.Content)
		logger.Info("WechatRobot msg", "Command", w.Command, "Prompt", w.Prompt, "msgId", msg.MsgID)
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
	w.Robot.SendMsg(chatId, helpText, msgId, tgbotapi.ModeMarkdown, nil)
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
		} else {
			modelList = []string{
				godeepseek.AzureDeepSeekR1,
				godeepseek.OpenRouterDeepSeekR1,
				godeepseek.OpenRouterDeepSeekR1DistillLlama70B,
				godeepseek.OpenRouterDeepSeekR1DistillLlama8B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen14B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B,
				godeepseek.OpenRouterDeepSeekR1DistillQwen32B,
				"llama2", // maps to LLAVA
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
	case param.LLAVA:
		modelList = []string{"llama2"}
	case param.OpenRouter:
		for k := range param.OpenRouterModels {
			modelList = append(modelList, k)
		}
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
		
		prompt := strings.TrimSpace(w.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var lastImageContent []byte
		var err error
		if len(lastImageContent) == 0 {
			lastImageContent, err = w.Robot.GetLastImageContent()
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
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(imageUrl) > 0 && len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		// 构建 base64 图片
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		format := utils.DetectImageFormat(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		
		fileName := "./data/" + utils.RandomFilename(format)
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
			Mode:       mode,
		})
	})
}

func (w *WechatRobot) sendVideo() {
	// 检查 prompt
	w.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := w.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(w.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			w.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var (
			videoUrl     string
			videoContent []byte
			err          error
		)
		
		mode := *conf.BaseConfInfo.MediaType
		var totalToken int
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			videoUrl, totalToken, err = llm.GenerateVolVideo(prompt, nil)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt, nil)
		default:
			err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 下载视频内容如果是 URL 模式
		if len(videoUrl) != 0 && len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(videoContent) == 0 {
			w.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		format := utils.DetectVideoMimeType(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		fileName := "./data/" + utils.RandomFilename(format)
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
			Mode:       mode,
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
	
	// send response message
	go w.handleUpdate(messageChan)
}

func (w *WechatRobot) handleUpdate(messageChan *MsgChan) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var msg *param.MsgInfo
	
	chatId, messageId, _ := w.Robot.GetChatIdAndMsgIdAndUserID()
	
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

func (w *WechatRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go w.handleUpdate(messageChan)
	
	go w.Robot.ExecLLM(w.Prompt, messageChan)
	
}

func (w *WechatRobot) GetContent(content string) (string, error) {
	if len(content) != 0 {
		return content, nil
	}
	
	msgType := w.Event.GetMsgType()
	
	switch msgType {
	case models.CALLBACK_MSG_TYPE_IMAGE:
		msg := serverModel.MessageImage{}
		err := w.Event.ReadMessage(&msg)
		if err != nil {
			logger.Error("read message", "err", err)
			return "", err
		}
		resp, err := w.App.Media.Get(w.Ctx, msg.MediaID)
		if err != nil {
			logger.Error("get media fail", "err", err)
			return "", err
		}
		
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return w.Robot.GetImageContent(data)
	
	case models.CALLBACK_MSG_TYPE_VOICE:
		msg := serverModel.MessageVoice{}
		err := w.Event.ReadMessage(&msg)
		if err != nil {
			logger.Error("read message", "err", err)
			return "", err
		}
		resp, err := w.App.Media.Get(w.Ctx, msg.MediaID)
		if err != nil {
			logger.Error("get media fail", "err", err)
			return "", err
		}
		
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
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
		if msgInfo, ok := WechatMsgMap.Load(msgId); ok {
			wechatMsg := msgInfo.(*WechatMessage)
			if wechatMsg.Status != msgHandling {
				return wechatMsg.Msg
			}
		}
		time.Sleep(1 * time.Second)
	}
	
	return ""
}

func WechatMsgSent(msgId string) {
	if msgInfo, ok := WechatMsgMap.Load(msgId); ok {
		wechatMsg := msgInfo.(*WechatMessage)
		wechatMsg.Status = msgSent
		WechatMsgMap.Store(msgId, wechatMsg)
	}
}

func (w *WechatRobot) GetPerMsgLen() int {
	return 1800
}
