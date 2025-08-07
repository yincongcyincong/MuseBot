package robot

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	
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

type Web struct {
	Command    string
	UserId     int64
	RealUserId string
	Prompt     string
	BodyData   []byte
	
	OriginalPrompt string
	
	W       http.ResponseWriter
	Flusher http.Flusher
	
	Robot *RobotInfo
}

func NewWeb(command string, userId int64, realUserId, prompt, originalPrompt string, bodyData []byte, w http.ResponseWriter, flusher http.Flusher) *Web {
	web := &Web{
		Command:        command,
		UserId:         userId,
		RealUserId:     realUserId,
		Prompt:         prompt,
		OriginalPrompt: originalPrompt,
		BodyData:       bodyData,
		W:              w,
		Flusher:        flusher,
	}
	web.Robot = NewRobot(WithRobot(web))
	return web
}

func (web *Web) Exec() {
	logger.Info("web exec", "command", web.Command, "userId", web.UserId, "prompt", web.Prompt)
	
	switch web.Command {
	case "/chat":
		web.handleChat()
	case "/mode":
		web.sendModelSelection()
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

func (web *Web) sendHelpConfigurationOptions() {
	chatId, msgId, _ := web.Robot.GetChatIdAndMsgIdAndUserID()
	
	helpText := `
Available Commands:

/chat   - Start a normal chat session

/mode   - Set the LLM mode

/balance - Check your current balance (tokens or credits)

/state  - View your current session state and settings

/clear  - Clear all conversation history

/retry  - Retry your last question

/photo  - Create a Image base on your prompt or your Image

/video  - Generate a video based on your prompt

/task   - Let multiple agents collaborate to complete a task

/mcp    - Use Multi-Agent Control Panel for complex task planning

/help   - Show this help message

`
	web.Robot.SendMsg(chatId, helpText, msgId, tgbotapi.ModeMarkdown, nil)
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     helpText,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
}

func (web *Web) sendModelSelection() {
	chatId, msgId, _ := web.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(web.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			web.Robot.handleModeUpdate(prompt)
			db.InsertRecordInfo(&db.Record{
				UserId:     web.RealUserId,
				Question:   web.OriginalPrompt,
				Answer:     i18n.GetMessage(*conf.BaseConfInfo.Lang, "mode_choose", nil) + prompt,
				Token:      0, // llm already calculate it
				IsDeleted:  0,
				RecordType: param.WEBRecordType,
			})
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
	
	web.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
	
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     totalContent,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
}

func (web *Web) showBalanceInfo() {
	chatId, msgId, _ := web.Robot.GetChatIdAndMsgIdAndUserID()
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	balance := llm.GetBalanceInfo()
	
	// handle balance info msg
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_title", nil), balance.IsAvailable)
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_content", nil)
	
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}
	
	web.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     msgContent,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
	
}

func (web *Web) showStateInfo() {
	chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		userInfo, err = db.GetUserByID(userId)
	}
	
	// get today token
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	todayTokey, err := db.GetTokenByUserIdAndTime(userId, startOfDay.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get today token fail", "err", err)
	}
	
	// get this week token
	startOf7DaysAgo := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	weekToken, err := db.GetTokenByUserIdAndTime(userId, startOf7DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}
	
	// handle balance info msg
	startOf30DaysAgo := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	monthToken, err := db.GetTokenByUserIdAndTime(userId, startOf30DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	web.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     msgContent,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
	
}

func (web *Web) clearAllRecord() {
	chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	deleteSuccMsg := i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil)
	web.Robot.SendMsg(chatId, deleteSuccMsg,
		msgId, tgbotapi.ModeMarkdown, nil)
	
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     deleteSuccMsg,
		Token:      0, // llm already calculate it
		IsDeleted:  1,
		RecordType: param.WEBRecordType,
	})
	return
	
}

func (web *Web) retryLastQuestion() {
	chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		web.Prompt = records.AQs[len(records.AQs)-1].Question
		web.handleChat()
	} else {
		web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
	
	return
	
}

func (web *Web) sendMultiAgent(agentType string) {
	chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(web.Prompt)
	if prompt == "" {
		logger.Warn("prompt is empty")
		web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	messageChan := make(chan string)
	
	dpReq := &llm.LLMTaskReq{
		Content:     prompt,
		UserId:      userId,
		ChatId:      chatId,
		MsgId:       msgId,
		HTTPMsgChan: messageChan,
	}
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("multi agent panic", "err", err, "stack", string(debug.Stack()))
			}
			close(messageChan)
		}()
		
		var err error
		if agentType == "mcp_empty_content" {
			err = dpReq.ExecuteMcp()
		} else {
			err = dpReq.ExecuteTask()
		}
		
		if err != nil {
			http.Error(web.W, err.Error(), http.StatusInternalServerError)
		}
	}()
	
	totalContent := ""
	for msg := range messageChan {
		fmt.Fprintf(web.W, "%s", msg)
		totalContent += msg
		web.Flusher.Flush()
	}
	
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     totalContent,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
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
		
		lastImageContent := web.BodyData
		var err error
		if len(lastImageContent) == 0 {
			lastImageContent, err = web.Robot.GetLastImageContent()
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
		
		if len(web.BodyData) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(web.BodyData)
			format = utils.DetectImageFormat(web.BodyData)
			originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		}
		
		// save message record
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      0,
			IsDeleted:  0,
			RecordType: param.WEBRecordType,
			Mode:       mode,
		})
		
		// save data record
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

func (web *Web) sendVideo() {
	// 检查 prompt
	web.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := web.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(web.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			web.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
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
			videoUrl, totalToken, err = llm.GenerateVolVideo(prompt)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt)
		default:
			err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 下载视频内容如果是 URL 模式
		if len(videoUrl) != 0 && len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(videoContent) == 0 {
			web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", utils.DetectVideoMimeType(videoContent), base64Content)
		
		fmt.Fprintf(web.W, "%s", dataURI)
		web.Flusher.Flush()
		
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Token:      0,
			IsDeleted:  0,
			RecordType: param.WEBRecordType,
			Mode:       mode,
		})
		
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
	
}

func (web *Web) handleChat() {
	web.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := web.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt, err := web.GetContent(strings.TrimSpace(web.Prompt))
		if err != nil {
			logger.Error("get content fail", "err", err)
			web.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		
		messageChan := make(chan string)
		l := llm.NewLLM(
			llm.WithChatId(chatId),
			llm.WithUserId(userId),
			llm.WithMsgId(msgId),
			llm.WithHTTPChain(messageChan),
			llm.WithContent(prompt),
		)
		go func() {
			defer close(messageChan)
			err := l.CallLLM()
			if err != nil {
				logger.Warn("Error sending message", "err", err)
				web.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			}
		}()
		
		totalContent := ""
		for msg := range messageChan {
			fmt.Fprintf(web.W, "%s", msg)
			totalContent += msg
			web.Flusher.Flush()
		}
		
		originDataURI := ""
		base64Content := base64.StdEncoding.EncodeToString(web.BodyData)
		if format := utils.DetectImageFormat(web.BodyData); format != "unknown" {
			originDataURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		} else if format := utils.DetectAudioFormat(web.BodyData); format != "unknown" {
			originDataURI = fmt.Sprintf("data:audio/%s;base64,%s", format, base64Content)
		}
		
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     totalContent,
			Content:    originDataURI,
			Token:      0, // llm already calculate it
			IsDeleted:  0,
			RecordType: param.WEBRecordType,
		})
	})
	
}

func (web *Web) GetContent(content string) (string, error) {
	var err error
	
	if content != "" {
		return content, nil
	}
	
	if len(web.BodyData) == 0 {
		logger.Warn("BodyData is empty")
		return "", errors.New("BodyData is empty")
	}
	
	if utils.DetectAudioFormat(web.BodyData) != "unknown" {
		content, err = web.Robot.GetAudioContent(web.BodyData)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return "", err
		}
	}
	
	if utils.DetectImageFormat(web.BodyData) != "unknown" {
		content, err = web.Robot.GetImageContent(web.BodyData)
		if err != nil {
			logger.Warn("get content from image failed", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}
