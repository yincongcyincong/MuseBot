package robot

import (
	"bytes"
	"context"
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
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	dingLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	dingUtils "github.com/open-dingtalk/dingtalk-stream-sdk-go/utils"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	DingBotClient   *client.StreamClient
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

	BotName      string
	OriginPrompt string
	baseRobot
}

type DingResp struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func StartDingRobot(ctx context.Context) {
	dingLogger.SetLogger(logger.Logger)
	DingBotClient = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(conf.BaseConfInfo.DingClientId, conf.BaseConfInfo.DingClientSecret)),
		client.WithUserAgent(client.NewDingtalkGoSDKUserAgent()),
		client.WithProxy(conf.BaseConfInfo.RobotProxy),
		client.WithSubscription(dingUtils.SubscriptionTypeKCallback, "/v1.0/im/bot/messages/get",
			chatbot.NewDefaultChatBotFrameHandler(OnChatReceive).OnEventReceived),
	)

	err := DingBotClient.Start(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "start dingbot fail", "err", err)
		return
	}

	logger.Info("DingRobot Info", "username", DingBotClient.UserAgent)
}

func OnChatReceive(ctx context.Context, message *chatbot.BotCallbackDataModel) ([]byte, error) {
	d := NewDingRobot(message)
	d.Robot = NewRobot(WithRobot(d), WithContext(ctx))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(ctx, "ding exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()

		d.Robot.Exec()
	}()

	return nil, nil
}

func NewDingRobot(message *chatbot.BotCallbackDataModel) *DingRobot {
	metrics.AppRequestCount.WithLabelValues("dingding").Inc()
	d := &DingRobot{
		Message: message,
		Client:  DingBotClient,
		BotName: BotName,
	}
	d.UserName = message.SenderNick
	return d
}

func (d *DingRobot) checkValid() bool {
	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()
	_, err := d.GetAccessToken()
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "get access token error", "err", err)
		return false
	}

	// group need to at bot
	atBot, err := d.GetMessageContent()
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "get message content error", "err", err)
		d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return false
	}
	if d.Message.ConversationType == "2" {
		if !atBot {
			logger.WarnCtx(d.Robot.Ctx, "no at bot")
			return false
		}
	}

	return true
}

func (d *DingRobot) getMsgContent() string {
	return d.Command
}

func (d *DingRobot) requestLLM(content string) {
	if !strings.Contains(content, "/") && !strings.Contains(content, "$") && d.Prompt == "" {
		d.Prompt = content
	}
	d.Robot.ExecCmd(content, d.sendChatMessage, nil, nil)
}

func (d *DingRobot) sendImg() {
	d.Robot.TalkingPreCheck(func() {
		d.Robot.generateMedia(d, mediaImage)
	})
}

func (d *DingRobot) sendMedia(media []byte, contentType, sType string) error {
	accessToken, err := d.GetAccessToken()
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "get access token fail", "err", err)
		return err
	}
	if sType == "image" {
		mediaId, err := d.UploadFileWithType(accessToken, "image", "image."+contentType, media)
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "upload file fail", "err", err)
			return err
		}

		_, err = d.SimpleReplyMarkdown(d.Robot.Ctx, []byte(fmt.Sprintf("![image](%s)", mediaId)))
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "send image fail", "err", err)
			return err
		}
	} else {
		mediaId, err := d.UploadFileWithType(accessToken, "video", "video."+contentType, media)
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "upload file fail", "err", err)
			return err
		}

		_, err = d.VideoReplyMarkdown(d.Robot.Ctx, mediaId, contentType)
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "send image fail", "err", err)
			return err
		}
	}

	return nil
}

func (d *DingRobot) sendVideo() {
	d.Robot.TalkingPreCheck(func() {
		d.Robot.generateMedia(d, mediaVideo)
	})
}

func (d *DingRobot) sendChatMessage() {
	d.Robot.TalkingPreCheck(func() {
		d.executeLLM()
	})

}

func (d *DingRobot) executeLLM() {
	d.Robot.runLLM(d, "amr", false)
}

func (d *DingRobot) GetMessageContent() (bool, error) {
	if d.Message == nil {
		return false, errors.New("callback data is nil")
	}

	chatId, msgId, _ := d.Robot.GetChatIdAndMsgIdAndUserID()

	accessToken, err := d.GetAccessToken()
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "get access token failed", "err", err)
		d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return false, err
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
				if _, ok := cMap["downloadCode"].(string); ok {
					// the download api is type-agnostic, sniff the bytes to
					// tell images and documents apart
					mediaContent, err := d.GetImageContent(accessToken, cMap)
					if err != nil {
						logger.Warn("get media content fail", "err", err)
					} else if utils.DetectImageFormat(mediaContent) != "unknown" {
						d.ImageContent = mediaContent
					} else {
						fileName, _ := cMap["fileName"].(string)
						if fileName == "" {
							fileName = "doc" + utils.DetectDocumentType("", mediaContent)
						}
						if utils.DetectDocumentType(fileName, mediaContent) != "" {
							d.DocContent = &param.DocInfo{FileName: fileName, Content: mediaContent}
						} else {
							logger.WarnCtx(d.Robot.Ctx, "unsupported richText media type", "fileName", fileName)
							d.Robot.SendMsg(chatId, i18n.GetMessage("doc_unsupported_type",
								map[string]interface{}{"file_name": fileName}), msgId, "", nil)
						}
					}
				}
			}
		}
	} else if d.Message.Msgtype == "picture" {
		if c, ok := d.Message.Content.(map[string]interface{}); ok {
			d.ImageContent, err = d.GetImageContent(accessToken, c)
			if err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "get image content fail", "err", err)
				d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
				return false, err
			}
		}

	} else if d.Message.Msgtype == "audio" {
		if c, ok := d.Message.Content.(map[string]interface{}); ok {
			d.AudioContent, err = d.GetImageContent(accessToken, c)
			if err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "get image content fail", "err", err)
				d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
				return false, err
			}
			content, err = d.Robot.GetAudioContent(d.AudioContent)
			if err != nil {
				logger.Warn("generate text from audio failed", "err", err)
				d.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
				return false, err
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
			"duration":     strconv.Itoa(conf.VideoConfInfo.Duration),
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

	httpClient := utils.GetRobotProxyClient()

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
		ClientId:     tea.String(conf.BaseConfInfo.DingClientId),
		ClientSecret: tea.String(conf.BaseConfInfo.DingClientSecret),
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

	client := &http.Client{Timeout: 5 * time.Minute}
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
		logger.ErrorCtx(d.Robot.Ctx, "create ding client failed", "err", err)
		return nil, err
	}

	if dc, ok := c["downloadCode"].(string); ok {
		robotMessageFileDownloadRequest := &dingtalkrobot.RobotMessageFileDownloadRequest{
			DownloadCode: tea.String(dc),
			RobotCode:    tea.String(conf.BaseConfInfo.DingClientId),
		}

		resp, err := dingClient.RobotMessageFileDownloadWithOptions(robotMessageFileDownloadRequest, &dingtalkrobot.RobotMessageFileDownloadHeaders{
			XAcsDingtalkAccessToken: tea.String(accessToken),
		}, &teaUtil.RuntimeOptions{})
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "download file failed", "err", err)
			return nil, err
		}

		data, err := utils.DownloadFile(tea.StringValue(resp.Body.DownloadUrl))
		if err != nil {
			logger.ErrorCtx(d.Robot.Ctx, "download file failed", "err", err)
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
		logger.ErrorCtx(d.Robot.Ctx, "get access token fail", "err", err)
		return err
	}

	mediaId, err := d.UploadFileWithType(accessToken, "voice", "voice."+format, voiceContent)
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "upload file fail", "err", err)
		return err
	}

	_, err = d.VoiceReplyMarkdown(d.Robot.Ctx, mediaId, duration)
	if err != nil {
		logger.ErrorCtx(d.Robot.Ctx, "send voice fail", "err", err)
		return err
	}

	return nil
}
