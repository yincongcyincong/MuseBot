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
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work/message/request"
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
	ComWechatApp *work.Work
)

type ComWechatRobot struct {
	Event contract.EventInterface
	Robot *RobotInfo
	App   *work.Work
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	OriginPrompt string
	ImageContent []byte
	TextMsg      *serverModel.MessageText
	VoiceMsg     *serverModel.MessageVoice
	ImageMsg     *serverModel.MessageImage
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
		logger.Error("ComWechatApp init error: ", err)
		return
	}
	
	resp, err := ComWechatApp.Agent.Get(ctx, utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID))
	if err != nil {
		logger.Error("ComWechatApp get agent error: ", err)
		return
	}
	logger.Info("ComWechatbot", "username", resp.Name)
}

func NewComWechatRobot(event contract.EventInterface) *ComWechatRobot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	
	c := &ComWechatRobot{
		Event:  event,
		App:    ComWechatApp,
		Ctx:    ctx,
		Cancel: cancel,
	}
	
	switch event.GetMsgType() {
	case models.CALLBACK_MSG_TYPE_TEXT:
		msg := &serverModel.MessageText{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil
		}
		c.TextMsg = msg
	case models.CALLBACK_MSG_TYPE_IMAGE:
		msg := &serverModel.MessageImage{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil
		}
		c.ImageMsg = msg
	case models.CALLBACK_MSG_TYPE_VOICE:
		msg := &serverModel.MessageVoice{}
		err := event.ReadMessage(msg)
		if err != nil {
			logger.Error("ComWechatRobot", "err", err)
			return nil
		}
		c.VoiceMsg = msg
	}
	
	return c
}

func (c *ComWechatRobot) checkValid() bool {
	if c.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_TEXT {
		c.OriginPrompt = c.TextMsg.Content
		c.Command, c.Prompt = ParseCommand(c.TextMsg.Content)
		logger.Info("ComWechatRobot msg", "Command", c.Command, "Prompt", c.Prompt)
	}
	
	if c.Event.GetMsgType() == models.CALLBACK_MSG_TYPE_IMAGE {
		_, msgId, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
		if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
			if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
				if msgInfo.Status == msgChangePhoto || msgInfo.Status == msgRecognizePhoto {
					logger.Info("ComWechatRobot handle photo msg", "msgId", msgId, "userId", userId)
					c.passiveExecCmd()
					return false
				}
			}
		}
	}
	
	return true
}

func (c *ComWechatRobot) getMsgContent() string {
	return c.Command
}

func (c *ComWechatRobot) requestLLMAndResp(content string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("ComWechatRobot panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		if !strings.Contains(content, "/") && c.Prompt == "" {
			c.Prompt = content
		}
		c.Robot.ExecCmd(content, c.sendChatMessage)
	}()
}

func (c *ComWechatRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
	c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "help_text", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}

func (c *ComWechatRobot) sendModeConfigurationOptions() {
	chatId, msgId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(c.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			c.Robot.handleModeUpdate(prompt)
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
	case param.OpenRouter, param.AI302:
		if c.Prompt != "" {
			c.Robot.handleModeUpdate(c.Prompt)
			return
		}
		switch *conf.BaseConfInfo.MediaType {
		case param.AI302:
			c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
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
	
	c.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (c *ComWechatRobot) sendImg() {
	c.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(c.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var lastImageContent []byte
		var err error
		if len(lastImageContent) == 0 && strings.Contains(c.Command, "edit_photo") {
			lastImageContent, err = c.Robot.GetLastImageContent()
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
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(imageUrl) > 0 && len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
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
			return
		}
		
		mediaID, err := c.App.Media.UploadTempImage(c.Ctx, fileName, nil)
		if err != nil {
			logger.Error("upload image fail", "err", err)
			return
		}
		
		_, err = c.App.Message.SendImage(c.Ctx, &request.RequestMessageSendImage{
			RequestMessageSend: request.RequestMessageSend{
				ToUser:                 userId,
				MsgType:                "image",
				AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
				DuplicateCheckInterval: 1800,
			},
			Image: &request.RequestImage{
				MediaID: mediaID.MediaID,
			},
		})
		if err != nil {
			logger.Error("send image fail", "err", err)
			return
		}
		
		// save data record
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   c.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

func (c *ComWechatRobot) sendVideo() {
	// 检查 prompt
	c.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(c.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
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
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 下载视频内容如果是 URL 模式
		if len(videoUrl) != 0 && len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(videoContent) == 0 {
			c.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
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
		mediaID, err := c.App.Media.UploadTempVideo(c.Ctx, fileName, nil)
		if err != nil {
			logger.Error("upload image fail", "err", err)
			return
		}
		
		_, err = c.App.Message.SendVideo(c.Ctx, &request.RequestMessageSendVideo{
			RequestMessageSend: request.RequestMessageSend{
				ToUser:                 userId,
				MsgType:                "video",
				AgentID:                utils.ParseInt(*conf.BaseConfInfo.ComWechatAgentID),
				DuplicateCheckInterval: 1800,
			},
			Video: &request.RequestVideo{
				MediaID: mediaID.MediaID,
			},
		})
		if err != nil {
			logger.Error("send image fail", "err", err)
			return
		}
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   c.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
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
	go c.handleUpdate(messageChan)
}

func (c *ComWechatRobot) handleUpdate(messageChan *MsgChan) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var msg *param.MsgInfo
	
	chatId, messageId, _ := c.Robot.GetChatIdAndMsgIdAndUserID()
	
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			c.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		msg = new(param.MsgInfo)
		return
	}
	
	c.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
}

func (c *ComWechatRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go c.handleUpdate(messageChan)
	
	go c.Robot.ExecLLM(c.Prompt, messageChan)
	
}

func (c *ComWechatRobot) getContent(content string) (string, error) {
	
	msgType := c.Event.GetMsgType()
	
	switch msgType {
	case models.CALLBACK_MSG_TYPE_IMAGE:
		data, err := c.getMedia()
		if err != nil {
			return "", err
		}
		return c.Robot.GetImageContent(data, content)
	
	case models.CALLBACK_MSG_TYPE_VOICE:
		data, err := c.getMedia()
		if err != nil {
			logger.Error("read media fail", "err", err)
			return "", err
		}
		
		data, err = utils.AmrToOgg(data)
		if err != nil {
			logger.Error("convert amr to wav fail", "err", err)
			return "", err
		}
		return c.Robot.GetAudioContent(data)
		
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

func (c *ComWechatRobot) passiveExecCmd() {
	c.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := c.Robot.GetChatIdAndMsgIdAndUserID()
		if !*conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate image")
			c.Robot.SendMsg(chatId, "only wechat_active is true can generate image", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if c.TextMsg != nil {
			status := msgChangePhoto
			switch c.Command {
			case "/change_photo", "change_photo":
				status = msgChangePhoto
			case "/rec_photo", "rec_photo":
				status = msgRecognizePhoto
			}
			TencentMsgMap.Store(userId, &TencentWechatMessage{
				Msg:       c.Prompt,
				Status:    status,
				StartTime: time.Now(),
			})
			c.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_pre_prompt_success", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if c.ImageMsg != nil {
			if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
				if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
					switch msgInfo.Status {
					case msgChangePhoto:
						c.Prompt = msgInfo.Msg
						data, err := c.getMedia()
						if err != nil {
							logger.Error("get media fail", "err", err)
							c.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						c.ImageContent = data
						c.sendImg()
					case msgRecognizePhoto:
						c.Prompt = msgInfo.Msg
						c.executeLLM()
					}
					
					TencentMsgMap.Delete(userId)
				}
			}
		}
	})
}

func (c *ComWechatRobot) getMedia() ([]byte, error) {
	resp, err := c.App.Media.Get(c.Ctx, c.ImageMsg.MediaID)
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
