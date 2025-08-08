package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/yincongcyincong/MuseBot/rag"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
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
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	BotName      string
	ImageContent []byte
}

func StartLarkRobot() {
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(LarkMessageHandler)
	
	cli = larkws.NewClient(*conf.BaseConfInfo.LarkAPPID, *conf.BaseConfInfo.LarkAppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	
	botClient = lark.NewClient(*conf.BaseConfInfo.LarkAPPID, *conf.BaseConfInfo.LarkAppSecret,
		lark.WithHttpClient(utils.GetRobotProxyClient()))
	
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
	go func() {
		if err := recover(); err != nil {
			logger.Error("exec panic", "err", err, "stack", string(debug.Stack()))
		}
		l.Robot.Exec()
	}()
	
	return nil
}

func (l *LarkRobot) checkValid() bool {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	// group need to at bot
	atBot, err := l.GetMessageContent()
	if err != nil {
		logger.Error("get message content error", "err", err)
		l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return false
	}
	if larkcore.StringValue(l.Message.Event.Message.ChatType) == "group" && *conf.BaseConfInfo.NeedATBOt {
		if !atBot {
			logger.Warn("no at bot")
			return false
		}
	}
	
	return true
}

func (l *LarkRobot) getMsgContent() string {
	return l.Command
}

func (l *LarkRobot) requestLLMAndResp(content string) {
	l.Robot.ExecCmd(content)
}

func (l *LarkRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	l.Robot.SendMsg(chatId, helpText, msgId, tgbotapi.ModeMarkdown, nil)
}

func (l *LarkRobot) sendModeConfigurationOptions() {
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
		l.sendChatMessage()
	} else {
		l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
	
	return
	
}

func (l *LarkRobot) sendMultiAgent(agentType string) {
	l.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(l.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		messageChan := make(chan *param.MsgInfo)
		
		dpReq := &llm.LLMTaskReq{
			Content:     prompt,
			UserId:      userId,
			ChatId:      chatId,
			MsgId:       msgId,
			MessageChan: messageChan,
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
		
		go l.handleUpdate(messageChan)
	})
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
		
		originalMsgID := l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			msgId, "", nil)
		
		lastImageContent := l.ImageContent
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
		
		if len(l.ImageContent) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(l.ImageContent)
			format = utils.DetectImageFormat(l.ImageContent)
			originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		}
		
		resp, err := l.Client.Im.V1.Image.Create(l.Ctx, larkim.NewCreateImageReqBuilder().
			Body(larkim.NewCreateImageReqBodyBuilder().
				ImageType("message").
				Image(bytes.NewReader(imageContent)).
				Build()).
			Build())
		if err != nil || !resp.Success() {
			logger.Warn("create image fail", "err", err, "resp", resp)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		msgContent, _ := larkim.NewMessagePost().ZhCn(larkim.NewMessagePostContent().AppendContent(
			[]larkim.MessagePostElement{
				&larkim.MessagePostImage{
					ImageKey: larkcore.StringValue(resp.Data.ImageKey),
				},
			}).Build()).Build()
		
		updateRes, err := l.Client.Im.Message.Update(l.Ctx, larkim.NewUpdateMessageReqBuilder().
			MessageId(originalMsgID).
			Body(larkim.NewUpdateMessageReqBodyBuilder().
				MsgType(larkim.MsgTypePost).
				Content(msgContent).
				Build()).
			Build())
		if err != nil || !updateRes.Success() {
			logger.Warn("send message fail", "err", err, "resp", resp)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
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
		
		originalMsgID := l.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			msgId, "", nil)
		
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
		
		resp, err := l.Client.Im.V1.File.Create(l.Ctx, larkim.NewCreateFileReqBuilder().
			Body(larkim.NewCreateFileReqBodyBuilder().
				FileType(utils.DetectVideoMimeType(videoContent)).
				FileName(fmt.Sprintf("%s.%s", prompt, utils.DetectVideoMimeType(videoContent))).
				Duration(*conf.VideoConfInfo.Duration).
				File(bytes.NewReader(videoContent)).
				Build()).
			Build())
		if err != nil || !resp.Success() {
			logger.Warn("create image fail", "err", err, "resp", resp)
			l.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		msgContent, _ := larkim.NewMessagePost().ZhCn(larkim.NewMessagePostContent().AppendContent(
			[]larkim.MessagePostElement{
				&larkim.MessagePostMedia{
					FileKey: larkcore.StringValue(resp.Data.FileKey),
				},
			}).Build()).Build()
		
		updateRes, err := l.Client.Im.Message.Update(l.Ctx, larkim.NewUpdateMessageReqBuilder().
			MessageId(originalMsgID).
			Body(larkim.NewUpdateMessageReqBodyBuilder().
				MsgType(larkim.MsgTypePost).
				Content(msgContent).
				Build()).
			Build())
		if err != nil || !updateRes.Success() {
			logger.Warn("send message fail", "err", err, "resp", resp)
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

func (l *LarkRobot) sendChatMessage() {
	chatId, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	content, err := l.GetContent(strings.TrimSpace(l.Prompt))
	if err != nil {
		logger.Error("get content fail", "err", err)
		l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	l.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			l.executeChain(content)
		} else {
			l.executeLLM(content)
		}
	})
	
}

func (l *LarkRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	chatId, msgId, userId := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
			}
			close(messageChan)
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		text, err := l.GetContent(content)
		if err != nil {
			logger.Error("get content fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		
		dpLLM := rag.NewRag(llm.WithMessageChan(messageChan), llm.WithContent(content),
			llm.WithChatId(chatId), llm.WithMsgId(msgId),
			llm.WithUserId(userId))
		
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err = chains.Run(ctx, qaChain, text)
		if err != nil {
			logger.Warn("execute chain fail", "err", err)
			l.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
	}()
	// send response message
	go l.handleUpdate(messageChan)
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
		if originalMsgID != "" {
			msg.MsgId = originalMsgID
		}
		
		if msg.MsgId == "" && originalMsgID == "" {
			resp, err := l.Client.Im.Message.Reply(l.Ctx, larkim.NewReplyMessageReqBuilder().
				MessageId(messageId).
				Body(larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					Content(GetMarkdownContent(msg.Content)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.Warn("send message fail", "err", err, "resp", resp)
				continue
			}
			msg.MsgId = larkcore.StringValue(resp.Data.MessageId)
		} else {
			
			resp, err := l.Client.Im.Message.Update(l.Ctx, larkim.NewUpdateMessageReqBuilder().
				MessageId(msg.MsgId).
				Body(larkim.NewUpdateMessageReqBodyBuilder().
					MsgType(larkim.MsgTypePost).
					Content(GetMarkdownContent(msg.Content)).
					Build()).
				Build())
			if err != nil || !resp.Success() {
				logger.Warn("send message fail", "err", err, "resp", resp)
				continue
			}
			originalMsgID = ""
		}
	}
}

func (l *LarkRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	go l.handleUpdate(messageChan)
	
	go func() {
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
		
	}()
	
}

func (l *LarkRobot) GetContent(content string) (string, error) {
	if len(content) != 0 {
		return content, nil
	}
	
	var err error
	msgType := larkcore.StringValue(l.Message.Event.Message.MessageType)
	_, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	
	switch msgType {
	case larkim.MsgTypeImage:
		msgImage := new(larkim.MessageImage)
		err = json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), msgImage)
		if err != nil {
			logger.Warn("unmarshal message image failed", "err", err)
			return "", err
		}
		
		resp, err := l.Client.Im.V1.MessageResource.Get(l.Ctx,
			larkim.NewGetMessageResourceReqBuilder().
				MessageId(msgId).
				FileKey(msgImage.ImageKey).
				Type("image").
				Build())
		if err != nil || !resp.Success() {
			logger.Error("get image failed", "err", err, "resp", resp)
			return "", err
		}
		
		bs, err := io.ReadAll(resp.File)
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
		resp, err := l.Client.Im.V1.MessageResource.Get(l.Ctx,
			larkim.NewGetMessageResourceReqBuilder().
				MessageId(msgId).
				FileKey(msgAudio.FileKey).
				Type("file").
				Build())
		if err != nil || !resp.Success() {
			logger.Error("get image failed", "err", err, "resp", resp)
			return "", err
		}
		
		bs, err := io.ReadAll(resp.File)
		if err != nil {
			logger.Error("read image failed", "err", err)
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

func GetMarkdownContent(content string) string {
	markdownMsg, _ := larkim.NewMessagePost().ZhCn(larkim.NewMessagePostContent().AppendContent(
		[]larkim.MessagePostElement{
			&MessagePostMarkdown{
				Text: strings.ReplaceAll(content, "\n", "\\n"),
			},
		}).Build()).Build()
	
	return markdownMsg
}

type MessagePostMarkdown struct {
	Text string `json:"text,omitempty"`
}

func (m *MessagePostMarkdown) Tag() string {
	return "md"
}

func (m *MessagePostMarkdown) IsPost() {
}

func (m *MessagePostMarkdown) MarshalJSON() ([]byte, error) {
	return []byte(`{"tag":"md","text":"` + m.Text + `"}`), nil
}

type MessagePostContent struct {
	Title   string                 `json:"title"`
	Content [][]MessagePostElement `json:"content"`
}

type MessagePostElement struct {
	Tag      string `json:"tag"`
	Text     string `json:"text"`
	ImageKey string `json:"image_key"`
	UserName string `json:"user_name"`
}

func (l *LarkRobot) GetMessageContent() (bool, error) {
	_, msgId, _ := l.Robot.GetChatIdAndMsgIdAndUserID()
	msgType := larkcore.StringValue(l.Message.Event.Message.MessageType)
	botShowName := ""
	if msgType == larkim.MsgTypeText {
		textMsg := new(MessageText)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), textMsg)
		if err != nil {
			logger.Error("unmarshal text message error", "error", err)
			return false, err
		}
		l.Command, l.Prompt = ParseCommand(textMsg.Text)
		for _, at := range l.Message.Event.Message.Mentions {
			if larkcore.StringValue(at.Name) == l.BotName {
				botShowName = larkcore.StringValue(at.Key)
				break
			}
		}
		
		l.Prompt = strings.ReplaceAll(l.Prompt, "@"+botShowName, "")
		for _, at := range l.Message.Event.Message.Mentions {
			if larkcore.StringValue(at.Name) == l.BotName {
				botShowName = larkcore.StringValue(at.Key)
				break
			}
		}
	} else if msgType == larkim.MsgTypePost {
		postMsg := new(MessagePostContent)
		err := json.Unmarshal([]byte(larkcore.StringValue(l.Message.Event.Message.Content)), postMsg)
		if err != nil {
			logger.Error("unmarshal text message error", "error", err)
			return false, err
		}
		
		for _, msgPostContents := range postMsg.Content {
			for _, msgPostContent := range msgPostContents {
				switch msgPostContent.Tag {
				case "text":
					command, prompt := ParseCommand(msgPostContent.Text)
					if command != "" {
						l.Command = command
					}
					if prompt != "" {
						l.Prompt = prompt
					}
				case "img":
					resp, err := l.Client.Im.V1.MessageResource.Get(l.Ctx,
						larkim.NewGetMessageResourceReqBuilder().
							MessageId(msgId).
							FileKey(msgPostContent.ImageKey).
							Type("image").
							Build())
					if err != nil || !resp.Success() {
						logger.Error("get image failed", "err", err, "resp", resp)
						return false, err
					}
					
					bs, err := io.ReadAll(resp.File)
					if err != nil {
						logger.Error("read image failed", "err", err)
						return false, err
					}
					l.ImageContent = bs
				case "at":
					if l.BotName == msgPostContent.UserName {
						botShowName = msgPostContent.UserName
					}
					
				}
			}
		}
	}
	
	l.Prompt = strings.ReplaceAll(l.Prompt, "@"+l.BotName, "")
	return botShowName == l.BotName, nil
}
