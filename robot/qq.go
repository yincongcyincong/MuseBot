package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"golang.org/x/oauth2"
)

var (
	QQApi         openapi.OpenAPI
	QQTokenSource oauth2.TokenSource
)

type QQRobot struct {
	event         *dto.WSPayload
	Robot         *RobotInfo
	QQApi         openapi.OpenAPI
	QQTokenSource oauth2.TokenSource
	
	C2CMessage *dto.WSC2CMessageData
	ATMessage  *dto.WSATMessageData
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	BotName      string
	OriginPrompt string
}

func StartQQRobot(ctx context.Context) {
	QQTokenSource = token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     *conf.BaseConfInfo.QQAppID,
			AppSecret: *conf.BaseConfInfo.QQAppSecret,
		})
	if err := token.StartRefreshAccessToken(ctx, QQTokenSource); err != nil {
		logger.Error("start refresh access token error", "err", err)
		return
	}
	
	QQApi = botgo.NewOpenAPI(*conf.BaseConfInfo.QQAppID, QQTokenSource).WithTimeout(5 * time.Second).SetDebug(true)
	resp, err := QQApi.Me(ctx)
	if err != nil {
		logger.Error("get me error", "err", err)
		return
	}
	logger.Info("QQRobot Info", "username", resp.Username)
	
	event.RegisterHandlers(
		event.ATMessageEventHandler(QQATMessageEventHandler),
		event.C2CMessageEventHandler(C2CMessageEventHandler),
	)
}

func C2CMessageEventHandler(event *dto.WSPayload, message *dto.WSC2CMessageData) error {
	d := NewQQRobot(event, message, nil)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("QQ exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		d.Robot.Exec()
	}()
	return nil
}

func QQATMessageEventHandler(event *dto.WSPayload, atMessage *dto.WSATMessageData) error {
	d := NewQQRobot(event, nil, atMessage)
	fmt.Println(event.EventID, atMessage.Content)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("QQ exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		d.Robot.Exec()
	}()
	
	return nil
}

func NewQQRobot(event *dto.WSPayload, c2cMessage *dto.WSC2CMessageData, atMessage *dto.WSATMessageData) *QQRobot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	return &QQRobot{
		QQApi:         QQApi,
		QQTokenSource: QQTokenSource,
		C2CMessage:    c2cMessage,
		ATMessage:     atMessage,
		event:         event,
		Ctx:           ctx,
		Cancel:        cancel,
		BotName:       BotName,
	}
}

func (q *QQRobot) checkValid() bool {
	if q.C2CMessage != nil {
		q.Command, q.Prompt = ParseCommand(q.C2CMessage.Content)
	}
	if q.ATMessage != nil {
		q.Command, q.Prompt = ParseCommand(q.ATMessage.Content)
	}
	
	return true
}

func (q *QQRobot) getMsgContent() string {
	return q.Command
}

func (q *QQRobot) requestLLMAndResp(content string) {
	if !strings.Contains(content, "/") && q.Prompt == "" {
		q.Prompt = content
	}
	q.Robot.ExecCmd(content, q.sendChatMessage)
}

func (q *QQRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	q.Robot.SendMsg(chatId, helpText, msgId, tgbotapi.ModeMarkdown, nil)
}

func (q *QQRobot) sendModeConfigurationOptions() {
	chatId, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(q.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			q.Robot.handleModeUpdate(prompt)
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
	
	q.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (q *QQRobot) sendImg() {
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var err error
		var lastImageContent []byte
		attachment := q.GetAttachment()
		if attachment != nil {
			lastImageContent, err = utils.DownloadFile(attachment.URL)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(lastImageContent) == 0 {
			lastImageContent, err = q.Robot.GetLastImageContent()
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
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(imageUrl) > 0 && len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		format := utils.DetectImageFormat(imageContent)
		data, err := q.UploadFile(imageContent, "image/"+format, 1)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		_, err = q.QQApi.PostC2CMessage(q.Ctx, q.C2CMessage.Author.ID, &dto.MessageToCreate{
			MsgType: dto.RichMediaMsg,
			MsgID:   msgId,
			MessageReference: &dto.MessageReference{
				MessageID:             msgId,
				IgnoreGetMessageError: false,
			},
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
		})
		
		if err != nil {
			logger.Warn("post message fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		originImageURI := ""
		if len(lastImageContent) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(lastImageContent)
			format = utils.DetectImageFormat(lastImageContent)
			originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		}
		
		// save data record
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   q.Prompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

func (q *QQRobot) sendVideo() {
	// 检查 prompt
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var imageContent []byte
		var err error
		attachment := q.GetAttachment()
		if attachment != nil {
			imageContent, err = utils.DownloadFile(attachment.URL)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		var (
			videoUrl     string
			videoContent []byte
		)
		
		mode := *conf.BaseConfInfo.MediaType
		var totalToken int
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			videoUrl, totalToken, err = llm.GenerateVolVideo(prompt, imageContent)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt, imageContent)
		default:
			err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if len(videoUrl) != 0 && len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}
		
		if len(videoContent) == 0 {
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectVideoMimeType(videoContent)
		data, err := q.UploadFile(imageContent, "video/"+format, 2)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		_, err = q.QQApi.PostC2CMessage(q.Ctx, q.C2CMessage.Author.ID, &dto.MessageToCreate{
			MsgType: dto.RichMediaMsg,
			MsgID:   msgId,
			MessageReference: &dto.MessageReference{
				MessageID:             msgId,
				IgnoreGetMessageError: false,
			},
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
		})
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   q.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
	
}

func (q *QQRobot) sendChatMessage() {
	q.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			q.executeChain()
		} else {
			q.executeLLM()
		}
	})
	
}

func (q *QQRobot) executeChain() {
	messageChan := make(chan *param.MsgInfo)
	go q.Robot.ExecChain(q.Prompt, messageChan)
	
	// send response message
	go q.handleUpdate(messageChan)
}

func (q *QQRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	chatId, messageId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	//thinkMsgID := q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
	//messageId, "", nil)
	
	var msg *param.MsgInfo
	for msg = range messageChan {
		if msg.Finished {
			q.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		msg = new(param.MsgInfo)
		return
	}
	
	q.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
	
	//err := q.QQApi.RetractC2CMessage(q.Ctx, chatId, thinkMsgID)
	//if err != nil {
	//	logger.Warn("retract message failed", "err", err)
	//}
}

func (q *QQRobot) executeLLM() {
	messageChan := make(chan *param.MsgInfo)
	go q.handleUpdate(messageChan)
	
	go q.Robot.ExecLLM(q.Prompt, messageChan)
	
}

func (q *QQRobot) GetContent(content string) (string, error) {
	if len(content) != 0 {
		return content, nil
	}
	
	attachment := q.GetAttachment()
	if attachment == nil {
		return "", errors.New("no attachments found")
	}
	
	switch {
	case strings.Contains(attachment.ContentType, "image"):
		data, err := utils.DownloadFile(attachment.URL)
		if err != nil {
			logger.Error("get image content fail", "err", err)
			return "", err
		}
		
		content, err = q.Robot.GetImageContent(data)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return "", err
		}
	
	case strings.Contains(attachment.ContentType, "voice"):
		data, err := utils.DownloadFile(attachment.URL)
		if err != nil {
			logger.Error("get image content fail", "err", err)
			return "", err
		}
		
		content, err = q.Robot.GetAudioContent(data)
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

func (q *QQRobot) getPrompt() string {
	return q.Prompt
}

type UploadFileRequest struct {
	FileType   int    `json:"file_type"`
	URL        string `json:"url"`
	SrvSendMsg bool   `json:"srv_send_msg"`
	FileData   string `json:"file_data"`
}

func (q *QQRobot) UploadFile(fileContent []byte, format string, fileType int) ([]byte, error) {
	chatId, _, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	
	apiURL := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/files", chatId)
	
	reqBody := UploadFileRequest{
		FileType:   fileType,
		URL:        "data:" + format + ";base64," + base64.StdEncoding.EncodeToString(fileContent),
		SrvSendMsg: false,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}
	
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	
	tokenInfo, err := q.QQTokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("get token error: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenInfo.AccessToken) // 如果需要鉴权的话
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()
	
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}
	
	return respData, err
}

func (q *QQRobot) GetAttachment() *dto.MessageAttachment {
	if q.C2CMessage != nil && len(q.C2CMessage.Attachments) != 0 {
		return q.C2CMessage.Attachments[0]
	}
	
	if q.ATMessage != nil && len(q.ATMessage.Attachments) != 0 {
		return q.ATMessage.Attachments[0]
	}
	
	return nil
}
