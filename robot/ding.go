package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkaiinteraction "github.com/alibabacloud-go/dingtalk/ai_interaction_1_0"
	dingtalkoauth2 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	dingtalkrobot "github.com/alibabacloud-go/dingtalk/robot_1_0"
	teaUtil "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	dingLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	dingUtils "github.com/open-dingtalk/dingtalk-stream-sdk-go/utils"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	dingBotClient   *client.StreamClient
	accessTokenInfo = new(AccessToken)
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiredTime int64  `json:"expired_time"`
}

type DingRobot struct {
	Robot   *RobotInfo
	Client  *client.StreamClient
	Message *chatbot.BotCallbackDataModel
	
	Ctx          context.Context
	Cancel       context.CancelFunc
	Command      string
	Prompt       string
	BotName      string
	OriginPrompt string
}

type DingResp struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func StartDingRobot(ctx context.Context) {
	dingLogger.SetLogger(logger.Logger)
	dingBotClient = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(*conf.BaseConfInfo.DingClientId, *conf.BaseConfInfo.DingClientSecret)),
		client.WithUserAgent(client.NewDingtalkGoSDKUserAgent()),
		client.WithProxy(*conf.BaseConfInfo.RobotProxy),
		client.WithSubscription(dingUtils.SubscriptionTypeKCallback, "/v1.0/im/bot/messages/get",
			chatbot.NewDefaultChatBotFrameHandler(OnChatReceive).OnEventReceived),
	)
	
	err := dingBotClient.Start(ctx)
	if err != nil {
		logger.Error("start dingbot fail", "err", err)
		return
	}
	
	logger.Info("DingRobot Info", "username", dingBotClient.UserAgent)
}

func OnChatReceive(ctx context.Context, message *chatbot.BotCallbackDataModel) ([]byte, error) {
	d := NewDingRobot(ctx, message)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("ding exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()
		
		d.Robot.Exec()
	}()
	
	return nil, nil
}

func NewDingRobot(ctx context.Context, message *chatbot.BotCallbackDataModel) *DingRobot {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	return &DingRobot{
		Message: message,
		Client:  dingBotClient,
		Ctx:     ctx,
		Cancel:  cancel,
		BotName: BotName,
	}
}

func (d *DingRobot) checkValid() bool {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	_, err := d.GetAccessToken()
	if err != nil {
		logger.Error("get access token error", "err", err)
		return false
	}
	
	// group need to at bot
	atBot, err := d.GetMessageContent()
	if err != nil {
		logger.Error("get message content error", "err", err)
		d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return false
	}
	if d.Message.ConversationType == "2" {
		if !atBot {
			logger.Warn("no at bot")
			return false
		}
	}
	
	return true
}

func (d *DingRobot) getMsgContent() string {
	return d.Command
}

func (d *DingRobot) requestLLMAndResp(content string) {
	if !strings.Contains(content, "/") && d.Prompt == "" {
		d.Prompt = content
	}
	d.Robot.ExecCmd(content, d.sendChatMessage)
}

func (d *DingRobot) sendHelpConfigurationOptions() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "help_text", nil), msgId, tgbotapi.ModeMarkdown, nil)
}

func (d *DingRobot) sendModeConfigurationOptions() {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	
	prompt := strings.TrimSpace(d.Prompt)
	if prompt != "" {
		if param.GeminiModels[prompt] || param.OpenAIModels[prompt] ||
			param.DeepseekModels[prompt] || param.DeepseekLocalModels[prompt] ||
			param.OpenRouterModels[prompt] || param.VolModels[prompt] {
			d.Robot.handleModeUpdate(prompt)
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
		if d.Prompt != "" {
			d.Robot.handleModeUpdate(d.Prompt)
			return
		}
		switch *conf.BaseConfInfo.Type {
		case param.AI302:
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "mix_mode_choose", map[string]interface{}{
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
	
	d.Robot.SendMsg(chatId, totalContent, msgId, "", nil)
}

func (d *DingRobot) sendImg() {
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(d.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		accessToken, err := d.GetAccessToken()
		if err != nil {
			logger.Warn("get access token fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var lastImageContent []byte
		if d.Message.Msgtype == "richText" {
			cMap := d.Message.Content.(map[string]interface{})
			for _, c := range cMap["richText"].([]interface{}) {
				if cMap, ok := c.(map[string]interface{}); ok {
					if _, ok := cMap["downloadCode"].(string); ok {
						lastImageContent, err = d.GetImageContent(accessToken, cMap)
						if err != nil {
							logger.Warn("get last image record fail", "err", err)
						}
					}
				}
			}
		}
		
		if len(lastImageContent) == 0 && strings.Contains(d.Command, "edit_photo") {
			lastImageContent, err = d.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := d.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectImageFormat(imageContent)
		
		mediaId, err := d.UploadFileWithType(accessToken, "image", "image."+format, imageContent)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		_, err = d.SimpleReplyMarkdown(d.Ctx, []byte(fmt.Sprintf("![image](%s)", mediaId)))
		if err != nil {
			logger.Warn("send image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		// 构建 base64 图片
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
			Question:   d.Prompt,
			Answer:     dataURI,
			Content:    originImageURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
}

func (d *DingRobot) sendVideo() {
	// 检查 prompt
	d.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := d.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(d.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			d.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		accessToken, err := d.GetAccessToken()
		if err != nil {
			logger.Warn("get access token fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var imageContent []byte
		if d.Message.Msgtype == "richText" {
			cMap := d.Message.Content.(map[string]interface{})
			for _, c := range cMap["richText"].([]interface{}) {
				if cMap, ok := c.(map[string]interface{}); ok {
					if _, ok := cMap["downloadCode"].(string); ok {
						imageContent, err = d.GetImageContent(accessToken, cMap)
						if err != nil {
							logger.Warn("get last image record fail", "err", err)
						}
					}
				}
			}
		}
		
		videoContent, totalToken, err := d.Robot.CreateVideo(prompt, imageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectVideoMimeType(videoContent)
		mediaId, err := d.UploadFileWithType(accessToken, "video", "video."+format, videoContent)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		_, err = d.VideoReplyMarkdown(d.Ctx, mediaId, format)
		if err != nil {
			logger.Warn("send image fail", "err", err)
			d.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   d.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       *conf.BaseConfInfo.MediaType,
		})
	})
	
}

func (d *DingRobot) sendChatMessage() {
	d.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			d.executeChain()
		} else {
			d.executeLLM()
		}
	})
	
}

func (d *DingRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go d.Robot.ExecChain(d.Prompt, messageChan)
	
	// send response message
	go d.Robot.HandleUpdate(messageChan, "amr")
}

func (d *DingRobot) sendText(messageChan *MsgChan) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	chatId, messageId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			d.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		return
	}
	
	if !msg.Finished {
		d.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
	}
}

//func (d *DingRobot) handleUpdate(messageChan chan *param.MsgInfo) {
//	defer func() {
//		if err := recover(); err != nil {
//			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
//		}
//	}()
//
//	var msg *param.MsgInfo
//
//	interClient, err := d.CreateDingInteractClient()
//	if err != nil {
//		logger.Error("create ding interact client failed", "err", err)
//		return
//	}
//
//	accessToken, _ := d.GetAccessToken()
//
//	prepareRequest := &dingtalkaiinteraction.PrepareRequest{
//		OpenConversationId: tea.String(d.Message.ConversationId),
//		ContentType:        tea.String("ai_card"),
//		Content:            tea.String(GetAIContent("")),
//	}
//
//	resp, err := interClient.PrepareWithOptions(prepareRequest, &dingtalkaiinteraction.PrepareHeaders{
//		XAcsDingtalkAccessToken: tea.String(accessToken),
//	}, &teaUtil.RuntimeOptions{})
//	if err != nil {
//		logger.Error("prepare failed", "err", err, "resp", resp)
//		return
//	}
//
//	for msg = range messageChan {
//		if msg == nil || len(msg.Content) == 0 {
//			msg.Content = "get nothing from llm!"
//		}
//
//		updateHeaders := &dingtalkaiinteraction.UpdateHeaders{}
//		updateHeaders.XAcsDingtalkAccessToken = tea.String(accessToken)
//		updateRequest := &dingtalkaiinteraction.UpdateRequest{
//			ConversationToken: tea.String(d.Message.ConversationId),
//			ContentType:       tea.String("ai_card"),
//			Content:           tea.String(GetAIContent(msg.Content)),
//		}
//		updateRsp, err := interClient.UpdateWithOptions(updateRequest, updateHeaders, &teaUtil.RuntimeOptions{})
//		if err != nil {
//			logger.Error("send message failed", "err", err, "updateRsp", updateRsp)
//		}
//	}
//
//	finishRequest := &dingtalkaiinteraction.FinishRequest{
//		ConversationToken: tea.String(d.Message.ConversationId),
//	}
//	finishRsp, err := interClient.FinishWithOptions(finishRequest, &dingtalkaiinteraction.FinishHeaders{
//		XAcsDingtalkAccessToken: tea.String(accessToken),
//	}, &teaUtil.RuntimeOptions{})
//	if err != nil {
//		logger.Error("prepare failed", "err", err, "resp", finishRsp)
//		return
//	}
//
//}

func (d *DingRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go d.Robot.HandleUpdate(messageChan, "amr")
	
	go d.Robot.ExecLLM(d.Prompt, messageChan)
	
}

func (d *DingRobot) getContent(content string) (string, error) {
	msgType := d.Message.Msgtype
	switch msgType {
	case "picture":
		if c, ok := d.Message.Content.(map[string]interface{}); ok {
			accessToken, err := d.GetAccessToken()
			if err != nil {
				logger.Error("get access token failed", "err", err)
				return "", err
			}
			
			data, err := d.GetImageContent(accessToken, c)
			if err != nil {
				logger.Error("get image content fail", "err", err)
				return "", err
			}
			
			content, err = d.Robot.GetImageContent(data, content)
			if err != nil {
				logger.Warn("generate text from audio failed", "err", err)
				return "", err
			}
		}
	case "audio":
		if c, ok := d.Message.Content.(map[string]interface{}); ok {
			accessToken, err := d.GetAccessToken()
			if err != nil {
				logger.Error("get access token failed", "err", err)
				return "", err
			}
			
			data, err := d.GetImageContent(accessToken, c)
			if err != nil {
				logger.Error("get image content fail", "err", err)
				return "", err
			}
			
			content, err = d.Robot.GetAudioContent(data)
			if err != nil {
				logger.Warn("generate text from audio failed", "err", err)
				return "", err
			}
		}
	}
	
	if content == "" {
		logger.Warn("content extraction returned empty")
		return "", errors.New("content is empty")
	}
	
	return content, nil
}

func (d *DingRobot) GetMessageContent() (bool, error) {
	if d.Message == nil {
		return false, errors.New("callback data is nil")
	}
	
	botAt := false
	content := d.Message.Text.Content
	if content == "" && d.Message.Msgtype == "richText" {
		cMap := d.Message.Content.(map[string]interface{})
		for _, c := range cMap["richText"].([]interface{}) {
			if cMap, ok := c.(map[string]interface{}); ok {
				if txt, ok := cMap["text"].(string); ok {
					content = txt
					break
				}
			}
		}
	}
	
	d.OriginPrompt = content
	d.Command, d.Prompt = ParseCommand(content)
	for _, atID := range d.Message.AtUsers {
		if atID.DingtalkId == d.Message.ChatbotUserId {
			botAt = true
			break
		}
	}
	
	return botAt, nil
}

func (d *DingRobot) getPrompt() string {
	return d.Prompt
}

func (d *DingRobot) SimpleReplyMarkdown(ctx context.Context, content []byte) (*DingResp, error) {
	requestBody := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": "muse-bot",
			"text":  string(content),
		},
	}
	return d.ReplyMessage(ctx, requestBody)
}

func (d *DingRobot) VideoReplyMarkdown(ctx context.Context, mediaId, format string) (*DingResp, error) {
	requestBody := map[string]interface{}{
		"msgtype": "video",
		"video": map[string]interface{}{
			"duration":     strconv.Itoa(*conf.VideoConfInfo.Duration),
			"videoMediaId": mediaId,
			"videoType":    format,
			"picMediaId":   "muse-bot",
		},
	}
	return d.ReplyMessage(ctx, requestBody)
}

func (d *DingRobot) VoiceReplyMarkdown(ctx context.Context, mediaId string, duration int) (*DingResp, error) {
	requestBody := map[string]interface{}{
		"msgtype": "audio",
		"audio": map[string]interface{}{
			"duration": duration,
			"mediaId":  mediaId,
		},
	}
	return d.ReplyMessage(ctx, requestBody)
}

func (d *DingRobot) ReplyMessage(ctx context.Context, requestBody map[string]interface{}) (*DingResp, error) {
	requestJsonBody, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.Message.SessionWebhook, bytes.NewReader(requestJsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	sendRsp := new(DingResp)
	err = json.Unmarshal(data, sendRsp)
	if err != nil {
		return nil, err
	}
	
	if sendRsp.ErrCode != 0 {
		return nil, errors.New(sendRsp.ErrMsg)
	}
	
	return sendRsp, err
}

func (d *DingRobot) CreateDingRobotClient() (result *dingtalkrobot.Client, err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	result = &dingtalkrobot.Client{}
	result, err = dingtalkrobot.NewClient(config)
	return result, err
}

func (d *DingRobot) CreateDingAuthClient() (_result *dingtalkoauth2.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkoauth2.Client{}
	_result, _err = dingtalkoauth2.NewClient(config)
	return _result, _err
}

func (d *DingRobot) CreateDingInteractClient() (_result *dingtalkaiinteraction.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkaiinteraction.Client{}
	_result, _err = dingtalkaiinteraction.NewClient(config)
	return _result, _err
}

func (d *DingRobot) GetAccessToken() (string, error) {
	if accessTokenInfo.ExpiredTime != 0 && accessTokenInfo.ExpiredTime < time.Now().Unix() {
		return accessTokenInfo.AccessToken, nil
	}
	
	getTokenRequest := &dingtalkoauth2.GetTokenRequest{
		ClientId:     tea.String(*conf.BaseConfInfo.DingClientId),
		ClientSecret: tea.String(*conf.BaseConfInfo.DingClientSecret),
		GrantType:    tea.String("client_credentials"),
	}
	
	authClient, err := d.CreateDingAuthClient()
	if err != nil {
		return "", err
	}
	
	resp, err := authClient.GetToken(tea.String(d.Message.ChatbotCorpId), getTokenRequest)
	if err != nil {
		return "", err
	}
	
	accessTokenInfo.AccessToken = tea.StringValue(resp.Body.AccessToken)
	accessTokenInfo.ExpiredTime = time.Now().Unix() + int64(tea.Int32Value(resp.Body.ExpiresIn)) - 300
	return accessTokenInfo.AccessToken, nil
}

func (d *DingRobot) GetAIContent(content string) string {
	data := map[string]interface{}{
		"templateId": *conf.BaseConfInfo.DingTemplateId,
		"cardData": map[string]interface{}{
			"content": content,
		},
	}
	dataByte, _ := json.Marshal(data)
	return string(dataByte)
}

type uploadResp struct {
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
	MediaId  string `json:"media_id"`
	FileType string `json:"type"`
}

func (d *DingRobot) UploadFileWithType(accessToken, fileType, filePath string, fileContent []byte) (string, error) {
	url := fmt.Sprintf("https://oapi.dingtalk.com/media/upload?access_token=%s", accessToken)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// media 字段，上传文件内容
	part, err := writer.CreateFormFile("media", filePath)
	if err != nil {
		return "", err
	}
	
	_, err = part.Write(fileContent)
	if err != nil {
		return "", err
	}
	
	err = writer.WriteField("type", fileType)
	if err != nil {
		return "", err
	}
	
	err = writer.Close()
	if err != nil {
		return "", err
	}
	
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var res uploadResp
	if err := json.Unmarshal(byteBody, &res); err != nil {
		return "", err
	}
	
	if res.ErrCode != 0 {
		return "", fmt.Errorf("dingtalk upload error: %s", res.ErrMsg)
	}
	
	return res.MediaId, nil
}

func (d *DingRobot) GetImageContent(accessToken string, c map[string]interface{}) ([]byte, error) {
	dingClient, err := d.CreateDingRobotClient()
	if err != nil {
		logger.Error("create ding client failed", "err", err)
		return nil, err
	}
	
	if dc, ok := c["downloadCode"].(string); ok {
		robotMessageFileDownloadRequest := &dingtalkrobot.RobotMessageFileDownloadRequest{
			DownloadCode: tea.String(dc),
			RobotCode:    tea.String(*conf.BaseConfInfo.DingClientId),
		}
		
		resp, err := dingClient.RobotMessageFileDownloadWithOptions(robotMessageFileDownloadRequest, &dingtalkrobot.RobotMessageFileDownloadHeaders{
			XAcsDingtalkAccessToken: tea.String(accessToken),
		}, &teaUtil.RuntimeOptions{})
		if err != nil {
			logger.Error("download file failed", "err", err)
			return nil, err
		}
		
		data, err := utils.DownloadFile(tea.StringValue(resp.Body.DownloadUrl))
		if err != nil {
			logger.Error("download file failed", "err", err)
			return nil, err
		}
		
		return data, nil
	}
	
	return nil, errors.New("download code not exist")
}

func (d *DingRobot) getPerMsgLen() int {
	return 4500
}

func (d *DingRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	format := utils.DetectAudioFormat(voiceContent)
	
	accessToken, err := d.GetAccessToken()
	if err != nil {
		logger.Warn("get access token fail", "err", err)
		return err
	}
	
	mediaId, err := d.UploadFileWithType(accessToken, "voice", "voice."+format, voiceContent)
	if err != nil {
		logger.Error("upload file fail", "err", err)
		return err
	}
	
	_, err = d.VoiceReplyMarkdown(d.Ctx, mediaId, duration)
	if err != nil {
		logger.Error("send voice fail", "err", err)
		return err
	}
	
	return nil
}
