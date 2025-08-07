package robot

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"strings"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type MessageText struct {
	Text string `json:"text"`
}

var (
	cli       *larkws.Client
	BotName   string
	botClient *lark.Client
)

type LarkRobot struct {
	Message *larkim.P2MessageReceiveV1
	Robot   *RobotInfo
	Client  *lark.Client
	
	Ctx     context.Context
	Cancel  context.CancelFunc
	Command string
	Prompt  string
	BotName string
}

func StartLarkRobot() {
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(LarkMessageHandler)
	
	cli = larkws.NewClient(*conf.BaseConfInfo.LarkAPPID, *conf.BaseConfInfo.LarkAppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	
	botClient = lark.NewClient(*conf.BaseConfInfo.LarkAPPID, *conf.BaseConfInfo.LarkAppSecret)
	
	// get bot name
	resp, err := botClient.Application.Application.Get(context.Background(), larkapplication.NewGetApplicationReqBuilder().
		AppId(*conf.BaseConfInfo.LarkAPPID).Lang("zh_cn").Build())
	if err != nil {
		logger.Error("get robot name error", "error", err)
		return
	}
	BotName = larkcore.StringValue(resp.Data.App.AppName)
	
	err = cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func NewLarkRobot(ctx context.Context, message *larkim.P2MessageReceiveV1) *LarkRobot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	return &LarkRobot{
		Message: message,
		Client:  botClient,
		Ctx:     ctx,
		Cancel:  cancel,
		BotName: BotName,
	}
}

func LarkMessageHandler(ctx context.Context, message *larkim.P2MessageReceiveV1) error {
	l := NewLarkRobot(ctx, message)
	l.Robot = NewRobot(WithRobot(l))
	go l.Robot.Exec()
	return nil
}

func (l *LarkRobot) Exec() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	msgType := larkcore.StringValue(l.Message.Event.Message.MessageType)
	if msgType == larkim.MsgTypeText {
		textMsg := new(MessageText)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), textMsg)
		if err != nil {
			logger.Error("unmarshal text message error", "error", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		l.Command, l.Prompt = ParseCommand(textMsg.Text)
		botShowName := ""
		for _, at := range l.Message.Event.Message.Mentions {
			if larkcore.StringValue(at.Name) == l.BotName {
				botShowName = larkcore.StringValue(at.Key)
				break
			}
		}
		
		l.Prompt = strings.ReplaceAll(l.Prompt, "@"+botShowName, "")
	}
	
	logger.Info("web exec", "command", l.Command)
	
	switch l.Command {
	case "/chat":
		l.handleChat()
	case "/mode":
		l.sendModelSelection()
	case "/balance":
		l.showBalanceInfo()
	case "/state":
		l.showStateInfo()
	case "/clear":
		l.clearAllRecord()
	case "/retry":
		l.retryLastQuestion()
	case "/photo":
		l.sendImg()
	case "/video":
		l.sendVideo()
	case "/help":
		l.sendHelpConfigurationOptions()
	case "/task":
		l.sendMultiAgent("task_empty_content")
	case "/mcp":
		l.sendMultiAgent("mcp_empty_content")
	default:
		l.handleChat()
	}
}

func (l *LarkRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	l.Robot.SendMsg(chatId, helpText, msgId, tgbotapi.ModeMarkdown, nil)
}

func (l *LarkRobot) sendModelSelection() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(l.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			l.Robot.handleModeUpdate(prompt)
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
	
	l.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (l *LarkRobot) showBalanceInfo() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil),
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
	
	l.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
}

func (l *LarkRobot) showStateInfo() {
	chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
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
	l.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
	
}

func (l *LarkRobot) clearAllRecord() {
	chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	deleteSuccMsg := i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil)
	l.Robot.SendMsg(chatId, deleteSuccMsg,
		msgId, tgbotapi.ModeMarkdown, nil)
	return
	
}

func (l *LarkRobot) retryLastQuestion() {
	chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		l.Prompt = records.AQs[len(records.AQs)-1].Question
		l.handleChat()
	} else {
		l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
	
	return
	
}

func (l *LarkRobot) sendMultiAgent(agentType string) {
	chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(l.Prompt)
	if prompt == "" {
		logger.Warn("prompt is empty")
		l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
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
			logger.Warn("execute task fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}()
}

func (l *LarkRobot) sendImg() {
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(l.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var lastImageContent []byte
		var err error
		if len(lastImageContent) == 0 {
			lastImageContent, err = l.Robot.GetLastImageContent()
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
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(imageUrl) > 0 && len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		// 构建 base64 图片
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		format := utils.DetectImageFormat(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		
		originImageURI := ""
		
		//if len(l.BodyData) > 0 {
		//	base64Content = base64.StdEncoding.EncodeToString(l.BodyData)
		//	format = utils.DetectImageFormat(l.BodyData)
		//	originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		//}
		
		// save data record
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   l.Prompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

func (l *LarkRobot) sendVideo() {
	// 检查 prompt
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(l.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
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
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 下载视频内容如果是 URL 模式
		if len(videoUrl) != 0 && len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(videoContent) == 0 {
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", utils.DetectVideoMimeType(videoContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   l.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
	
}

func (l *LarkRobot) handleChat() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	content, err := l.GetContent(strings.TrimSpace(l.Prompt))
	if err != nil {
		logger.Error("get content fail", "err", err)
		l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	if conf.RagConfInfo.Store != nil {
		//l.executeChain(content)
	} else {
		l.executeLLM(content)
	}
	
}

func (l *LarkRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var msg *param.MsgInfo
	
	chatId, messageId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	originalMsgID := l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
		messageId, "", nil)
	
	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from llm!"
		}
		
		if msg.MsgId == 0 {
			resp, err := l.Client.Im.Message.Reply(l.Ctx, larkim.NewReplyMessageReqBuilder().
				MessageId(messageId).
				Body(larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeText).
					Content(larkim.NewMessageTextBuilder().
						Text(msg.Content).Build()).
					Build()).
				Build())
			if err != nil {
				logger.Warn("send message fail", "err", err)
			}
			if !resp.Success() {
				logger.Warn("send message fail", "resp", resp)
			}
			
		} else {
			resp, err := l.Client.Im.Message.Update(l.Ctx, larkim.NewUpdateMessageReqBuilder().
				MessageId(originalMsgID).
				Body(larkim.NewUpdateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeText).
					Content(larkim.NewMessageTextBuilder().
						Text(msg.Content).Build()).
					Build()).
				Build())
			if err != nil {
				logger.Warn("send message fail", "err", err)
			}
			if !resp.Success() {
				logger.Warn("send message fail", "resp", resp)
			}
		}
	}
}

func (l *LarkRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	go l.handleUpdate(messageChan)
	
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		llmClient := llm.NewLLM(
			llm.WithChatId(chatId),
			llm.WithUserId(userId),
			llm.WithMsgId(msgId),
			llm.WithMessageChan(messageChan),
			llm.WithContent(content),
		)
		
		err := llmClient.CallLLM()
		if err != nil {
			logger.Error("get content fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
		
	})
	
}

func (l *LarkRobot) GetContent(content string) (string, error) {
	if len(content) != 0 {
		return content, nil
	}
	
	var err error
	msgType := larkcore.StringValue(l.Message.Event.Message.MessageType)
	
	switch msgType {
	case larkim.MsgTypeImage:
		msgImage := new(larkim.MessageImage)
		err = json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), msgImage)
		if err != nil {
			logger.Warn("unmarshal message image failed", "err", err)
			return "", err
		}
		resp, err := l.Client.Im.Image.Get(l.Ctx,
			larkim.NewGetImageReqBuilder().ImageKey(msgImage.ImageKey).Build())
		if err != nil {
			logger.Error("get image failed", "err", err)
			return "", err
		}
		
		bs, err := ioutil.ReadAll(resp.File)
		if err != nil {
			logger.Error("read image failed", "err", err)
			return "", err
		}
		
		content, err = l.Robot.GetImageContent(bs)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return "", err
		}
	case larkim.MsgTypeAudio:
		msgAudio := new(larkim.MessageAudio)
		err = json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), msgAudio)
		if err != nil {
			logger.Warn("unmarshal message audio failed", "err", err)
			return "", err
		}
		resp, err := l.Client.Im.File.Get(l.Ctx,
			larkim.NewGetFileReqBuilder().FileKey(msgAudio.FileKey).Build())
		if err != nil {
			logger.Error("get audio failed", "err", err)
			return "", err
		}
		
		bs, err := ioutil.ReadAll(resp.File)
		if err != nil {
			logger.Error("read audio failed", "err", err)
			return "", err
		}
		
		content, err = l.Robot.GetAudioContent(bs)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}
