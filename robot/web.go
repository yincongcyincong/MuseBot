package robot

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Web struct {
	Command    string
	UserId     int64
	RealUserId string
	Prompt     string
	ImageData  []byte
	
	OriginalPrompt string
	
	W       http.ResponseWriter
	Flusher http.Flusher
	
	Robot *RobotInfo
}

func NewWeb(command string, userId int64, realUserId, prompt, originalPrompt string, imageData []byte, w http.ResponseWriter, flusher http.Flusher) *Web {
	web := &Web{
		Command:        command,
		UserId:         userId,
		RealUserId:     realUserId,
		Prompt:         prompt,
		OriginalPrompt: originalPrompt,
		ImageData:      imageData,
		W:              w,
		Flusher:        flusher,
	}
	web.Robot = NewRobot(WithRobot(web))
	return web
}

func (web *Web) Exec() {
	switch web.Command {
	case "/chat":
		web.handleChat()
	case "/mode":
		web.sendModeConfigurationOptions()
	case "/balance":
		web.showBalanceInfo()
	case "/state":
		web.showStateInfo()
	case "/clear":
		web.clearAllRecord()
	case "/retry":
		web.retryLastQuestion()
	case "/photo":
		web.sendImg()
	case "/video":
		web.sendVideo()
	case "/help":
		web.sendHelpConfigurationOptions()
	case "/task":
		web.sendMultiAgent("task_empty_content")
	case "/mcp":
		web.sendMultiAgent("mcp_empty_content")
	default:
		web.handleChat()
	}
}

func (web *Web) sendModeConfigurationOptions() {
	
	return
	
}
func (web *Web) showBalanceInfo() {
	
	return
	
}

func (web *Web) showStateInfo() {
	
	return
	
}

func (web *Web) clearAllRecord() {
	chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
	return
	
}

func (web *Web) retryLastQuestion() {
	
	return
	
}

func (web *Web) sendHelpConfigurationOptions() {
	
	return
	
}

func (web *Web) sendMultiAgent(agentType string) {
	
	return
	
}

func (web *Web) sendImg() {
	web.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := web.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(web.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		lastImageContent := web.ImageData
		var err error
		if len(lastImageContent) == 0 {
			lastImageContent, err = web.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		var imageUrl string
		var imageContent []byte
		
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			imageUrl, err = llm.GenerateVolImg(prompt, lastImageContent)
		case param.OpenAi:
			imageContent, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
		case param.Gemini:
			imageContent, err = llm.GenerateGeminiImg(prompt, lastImageContent)
		default:
			err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(imageUrl) > 0 && len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		// 构建 base64 图片
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		format := utils.DetectImageFormat(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		
		fmt.Fprintf(web.W, "%s", dataURI)
		web.Flusher.Flush()
		
		originImageURI := ""
		
		if len(web.ImageData) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(web.ImageData)
			format = utils.DetectImageFormat(imageContent)
			originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		}
		
		// 保存记录
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      param.ImageTokenUsage,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
		})
		
	})
	
}

func (web *Web) sendVideo() {
	
	return
	
}

func (web *Web) handleChat() {
	web.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(web.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		messageChan := make(chan string)
		l := llm.NewLLM(
			llm.WithChatId(chatId),
			llm.WithUserId(userId),
			llm.WithMsgId(msgId),
			llm.WithHTTPChain(messageChan),
			llm.WithContent(web.Prompt),
		)
		go func() {
			defer close(messageChan)
			err := l.CallLLM()
			if err != nil {
				logger.Warn("Error sending message", "err", err)
				web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			}
		}()
		
		for msg := range messageChan {
			fmt.Fprintf(web.W, "%s", msg)
			web.Flusher.Flush()
		}
		
	})
}
