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
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type Web struct {
	Command    string
	UserId     int64
	RealUserId string
	Prompt     string
	BodyData   []byte
	RecordId   int64
	
	OriginalPrompt string
	
	W       http.ResponseWriter
	Flusher http.Flusher
	
	Robot *RobotInfo
}

func NewWeb(command string, userId int64, realUserId, prompt, originalPrompt string, bodyData []byte, w http.ResponseWriter, flusher http.Flusher) *Web {
	metrics.AppRequestCount.WithLabelValues("web").Inc()
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
	web.Robot = NewRobot()
	return web
}

func (web *Web) checkValid() bool {
	return true
}

func (web *Web) getMsgContent() string {
	return web.Command
}

func (web *Web) Exec() {
	logger.InfoCtx(web.Robot.Ctx, "web exec", "command", web.Command, "userId", web.UserId, "prompt", web.Prompt)
	switch web.Command {
	case "/chat":
		web.InsertRecord()
		web.sendChatMessage()
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
		web.InsertRecord()
		web.sendChatMessage()
	}
}

func (web *Web) sendHelpConfigurationOptions() {
	web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "help_text", nil))
	db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     i18n.GetMessage(*conf.BaseConfInfo.Lang, "help_text", nil),
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
}

func (web *Web) sendModeConfigurationOptions() {
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
		if web.Prompt != "" {
			web.Robot.handleModeUpdate(web.Prompt)
			return
		}
		switch *conf.BaseConfInfo.Type {
		case param.AI302:
			modelList = append(modelList, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}))
		case param.OpenRouter:
			modelList = append(modelList, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}))
		case param.Ollama:
			modelList = append(modelList, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://ollama.com/",
			}))
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
	
	web.SendMsg(totalContent)
	
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
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil))
		return
	}
	
	balance := llm.GetBalanceInfo(web.Robot.Ctx)
	
	// handle balance info msg
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_title", nil), balance.IsAvailable)
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_content", nil)
	
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}
	
	web.SendMsg(msgContent)
	
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
	userId := web.RealUserId
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "get user info fail", "err", err)
		web.SendMsg(err.Error())
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
		logger.WarnCtx(web.Robot.Ctx, "get today token fail", "err", err)
	}
	
	// get this week token
	startOf7DaysAgo := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	weekToken, err := db.GetTokenByUserIdAndTime(userId, startOf7DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "get week token fail", "err", err)
	}
	
	// handle balance info msg
	startOf30DaysAgo := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	monthToken, err := db.GetTokenByUserIdAndTime(userId, startOf30DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "get week token fail", "err", err)
	}
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	web.SendMsg(msgContent)
	
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
	userId := web.RealUserId
	db.DeleteMsgRecord(userId)
	deleteSuccMsg := i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil)
	web.SendMsg(deleteSuccMsg)
	
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
	userId := web.RealUserId
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		web.Prompt = records.AQs[len(records.AQs)-1].Question
		web.sendChatMessage()
	} else {
		web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil))
	}
	
	return
	
}

func (web *Web) sendMultiAgent(agentType string) {
	
	prompt := strings.TrimSpace(web.Prompt)
	if prompt == "" {
		logger.WarnCtx(web.Robot.Ctx, "prompt is empty")
		web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil))
		return
	}
	
	messageChan := make(chan string)
	
	dpReq := &llm.LLMTaskReq{
		Content:     prompt,
		UserId:      web.RealUserId,
		ChatId:      web.RealUserId,
		MsgId:       web.RealUserId,
		HTTPMsgChan: messageChan,
		PerMsgLen:   10000000,
		
		Ctx: web.Robot.Ctx,
	}
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(web.Robot.Ctx, "multi agent panic", "err", err, "stack", string(debug.Stack()))
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
		prompt := strings.TrimSpace(web.Prompt)
		if prompt == "" {
			logger.WarnCtx(web.Robot.Ctx, "prompt is empty")
			web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil))
			return
		}
		
		lastImageContent := web.BodyData
		var err error
		if len(lastImageContent) == 0 && strings.Contains(web.Command, "edit_photo") {
			lastImageContent, err = web.Robot.GetLastImageContent()
			if err != nil {
				logger.WarnCtx(web.Robot.Ctx, "get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := web.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.WarnCtx(web.Robot.Ctx, "generate image fail", "err", err)
			web.SendMsg(err.Error())
			return
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
			Mode:       *conf.BaseConfInfo.MediaType,
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
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
}

func (web *Web) sendVideo() {
	// 检查 prompt
	web.Robot.TalkingPreCheck(func() {
		
		prompt := strings.TrimSpace(web.Prompt)
		if prompt == "" {
			logger.WarnCtx(web.Robot.Ctx, "prompt is empty")
			web.SendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil))
			return
		}
		
		videoContent, totalToken, err := web.Robot.CreateVideo(prompt, web.BodyData)
		if err != nil {
			logger.WarnCtx(web.Robot.Ctx, "generate video fail", "err", err)
			web.SendMsg(err.Error())
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
			Mode:       *conf.BaseConfInfo.MediaType,
		})
		
		db.InsertRecordInfo(&db.Record{
			UserId:     web.RealUserId,
			Question:   web.OriginalPrompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
	
}

func (web *Web) sendChatMessage() {
	web.Robot.TalkingPreCheck(func() {
		
		prompt, err := web.GetContent(strings.TrimSpace(web.Prompt))
		if err != nil {
			logger.ErrorCtx(web.Robot.Ctx, "get content fail", "err", err)
			web.SendMsg(err.Error())
			return
		}
		
		messageChan := make(chan string)
		l := llm.NewLLM(
			llm.WithChatId(web.RealUserId),
			llm.WithUserId(web.RealUserId),
			llm.WithMsgId(web.RealUserId),
			llm.WithHTTPMsgChan(messageChan),
			llm.WithContent(prompt),
			llm.WithPerMsgLen(1000000),
			llm.WithRecordId(web.RecordId),
			llm.WithContext(web.Robot.Ctx),
		)
		go func() {
			defer close(messageChan)
			err := l.CallLLM()
			if err != nil {
				logger.WarnCtx(web.Robot.Ctx, "Error sending message", "err", err)
				web.SendMsg(err.Error())
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
	if len(web.BodyData) == 0 {
		return content, nil
	}
	
	if utils.DetectAudioFormat(web.BodyData) != "unknown" {
		content, err = web.Robot.GetAudioContent(web.BodyData)
		if err != nil {
			logger.WarnCtx(web.Robot.Ctx, "generate text from audio failed", "err", err)
			return "", err
		}
	}
	
	if utils.DetectImageFormat(web.BodyData) != "unknown" {
		content, err = web.Robot.GetImageContent(web.BodyData, content)
		if err != nil {
			logger.WarnCtx(web.Robot.Ctx, "get content from image failed", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.WarnCtx(web.Robot.Ctx, "content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}

func (web *Web) SendMsg(msgContent string) {
	_, err := web.W.Write([]byte(msgContent))
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "send message fail", "err", err)
	}
	web.Flusher.Flush()
}

func (web *Web) InsertRecord() {
	id, err := db.InsertRecordInfo(&db.Record{
		UserId:     web.RealUserId,
		Question:   web.Prompt,
		RecordType: param.TextRecordType,
	})
	if err != nil {
		logger.ErrorCtx(web.Robot.Ctx, "insert record fail", "err", err)
		return
	}
	
	web.RecordId = id
}
