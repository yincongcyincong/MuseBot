package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"golang.org/x/oauth2"
)

var (
	QQApi         openapi.OpenAPI
	QQTokenSource oauth2.TokenSource
	QQRobotInfo   *dto.User
)

type QQRobot struct {
	Robot         *RobotInfo
	QQApi         openapi.OpenAPI
	QQTokenSource oauth2.TokenSource
	RobotInfo     *dto.User

	C2CMessage     *dto.WSC2CMessageData
	GroupAtMessage *dto.WSGroupATMessageData
	ATMessage      *dto.WSATMessageData

	BotName      string
	OriginPrompt string
	baseRobot
}

func StartQQRobot(ctx context.Context) {
	var err error
	botgo.SetLogger(logger.QQLogger)
	QQTokenSource = token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     conf.BaseConfInfo.QQAppID,
			AppSecret: conf.BaseConfInfo.QQAppSecret,
		})
	if err = token.StartRefreshAccessToken(ctx, QQTokenSource); err != nil {
		logger.ErrorCtx(ctx, "start refresh access token error", "err", err)
		return
	}

	QQApi = botgo.NewOpenAPI(conf.BaseConfInfo.QQAppID, QQTokenSource).WithTimeout(5 * time.Second).SetDebug(false)
	QQRobotInfo, err = QQApi.Me(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "get me error", "err", err)
		return
	}
	logger.InfoCtx(ctx, "QQRobot Info", "username", QQRobotInfo.Username)

	event.RegisterHandlers(
		event.GroupATMessageEventHandler(QQGroupATMessageEventHandler),
		event.ATMessageEventHandler(QQATMessageEventHandler),
		event.C2CMessageEventHandler(C2CMessageEventHandler),
	)
}

func C2CMessageEventHandler(event *dto.WSPayload, message *dto.WSC2CMessageData) error {
	d := NewQQRobot(message, nil, nil)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "QQ exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()

		d.Robot.Exec()
	}()
	return nil
}

func QQGroupATMessageEventHandler(event *dto.WSPayload, atMessage *dto.WSGroupATMessageData) error {
	d := NewQQRobot(nil, atMessage, nil)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "QQ exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()

		d.Robot.Exec()
	}()

	return nil
}

func QQATMessageEventHandler(event *dto.WSPayload, atMessage *dto.WSATMessageData) error {
	d := NewQQRobot(nil, nil, atMessage)
	d.Robot = NewRobot(WithRobot(d))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCtx(d.Robot.Ctx, "QQ exec panic", "err", err, "stack", string(debug.Stack()))
			}
		}()

		d.Robot.Exec()
	}()

	return nil
}

func NewQQRobot(c2cMessage *dto.WSC2CMessageData,
	atGroupMessage *dto.WSGroupATMessageData, atMessage *dto.WSATMessageData) *QQRobot {
	metrics.AppRequestCount.WithLabelValues("qq").Inc()
	qr := &QQRobot{
		QQApi:          QQApi,
		QQTokenSource:  QQTokenSource,
		RobotInfo:      QQRobotInfo,
		C2CMessage:     c2cMessage,
		ATMessage:      atMessage,
		GroupAtMessage: atGroupMessage,
		BotName:        BotName,
	}

	if c2cMessage != nil {
		qr.UserName = c2cMessage.Author.Username
	}
	if atGroupMessage != nil {
		qr.UserName = atGroupMessage.Author.Username
	}
	if atMessage != nil {
		qr.UserName = atMessage.Author.Username
	}

	return qr
}

func (q *QQRobot) checkValid() bool {
	chatId, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	var err error
	if q.C2CMessage != nil {
		q.Command, q.Prompt = ParseCommand(q.C2CMessage.Content)
	}
	if q.ATMessage != nil {
		q.Command, q.Prompt = ParseCommand(q.ATMessage.Content)
	}
	if q.GroupAtMessage != nil {
		q.Command, q.Prompt = ParseCommand(q.GroupAtMessage.Content)
	}

	if q.GetAttachment() != nil && strings.Contains(q.GetAttachment().ContentType, "image") {
		q.ImageContent, err = utils.DownloadFile(q.GetAttachment().URL)
		if err != nil {
			q.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
			logger.ErrorCtx(q.Robot.Ctx, "download image fail", "err", err)
			return false
		}
	}

	if q.GetAttachment() != nil && strings.Contains(q.GetAttachment().ContentType, "voice") {
		q.AudioContent, err = utils.DownloadFile(q.GetAttachment().URL)
		if err != nil {
			q.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
			logger.ErrorCtx(q.Robot.Ctx, "get image content fail", "err", err)
			return false
		}
		data, err := utils.SilkToWav(q.AudioContent)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "silk to wav fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}

		q.Prompt, err = q.Robot.GetAudioContent(data)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "generate text from audio failed", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return false
		}
	}

	if q.GetAttachment() != nil && q.GetAttachment().ContentType != "" &&
		!strings.Contains(q.GetAttachment().ContentType, "image") &&
		!strings.Contains(q.GetAttachment().ContentType, "voice") &&
		!strings.Contains(q.GetAttachment().ContentType, "video") {
		// any other attachment (pdf, txt, csv, html, and text-like files
		// with other extensions) is a document
		docContent, err := utils.DownloadFile(q.GetAttachment().URL)
		if err != nil {
			q.Robot.SendMsg(chatId, "get media fail", msgId, tgbotapi.ModeMarkdown, nil)
			logger.ErrorCtx(q.Robot.Ctx, "download document fail", "err", err)
			return false
		}
		if utils.DetectDocumentType(q.GetAttachment().FileName, docContent) == "" {
			logger.WarnCtx(q.Robot.Ctx, "unsupported document type", "fileName", q.GetAttachment().FileName)
			q.Robot.SendMsg(chatId, i18n.GetMessage("doc_unsupported_type",
				map[string]interface{}{"file_name": q.GetAttachment().FileName}), msgId, "", nil)
			return true
		}

		q.DocContent = &param.DocInfo{FileName: q.GetAttachment().FileName, Content: docContent}
	}

	return true
}

func (q *QQRobot) getMsgContent() string {
	return q.Command
}

func (q *QQRobot) requestLLM(content string) {
	if !strings.Contains(content, "/") && !strings.Contains(content, "$") && q.Prompt == "" {
		q.Prompt = content
	}
	q.Robot.ExecCmd(content, q.sendChatMessage, nil, nil)
}

func (q *QQRobot) sendImg() {
	q.Robot.TalkingPreCheck(func() {
		q.Robot.generateMedia(q, mediaImage)
	})
}

func (q *QQRobot) sendMedia(media []byte, contentType, sType string) error {
	if sType == "image" {
		base64Content := base64.StdEncoding.EncodeToString(media)
		data, err := q.UploadFile(base64Content, 1)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "upload file fail", "err", err)
			return err
		}

		err = q.PostRichMediaMessage(data, "")
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "post rich media msg fail", "err", err)
			return err
		}
	} else {
		base64Content := base64.StdEncoding.EncodeToString(media)
		data, err := q.UploadFile(base64Content, 2)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "upload file fail", "err", err)
			return err
		}

		err = q.PostRichMediaMessage(data, "")
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "post rich media msg fail", "err", err)
			return err
		}
	}

	return nil
}

// sendVideo keeps its own body: unlike the other platforms it re-downloads the
// message attachment as the image-to-video source instead of reusing
// ImageContent.
func (q *QQRobot) sendVideo() {
	// 检查 prompt
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()

		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			logger.ErrorCtx(q.Robot.Ctx, "prompt is empty")
			q.Robot.SendMsg(chatId, i18n.GetMessage("video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}

		var imageContent []byte
		var err error
		attachment := q.GetAttachment()
		if attachment != nil {
			imageContent, err = utils.DownloadFile(attachment.URL)
			if err != nil {
				logger.ErrorCtx(q.Robot.Ctx, "download image fail", "err", err)
				q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return
			}
		}

		videoContent, totalToken, err := q.Robot.CreateVideo(prompt, imageContent)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "generate video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}

		err = q.sendMedia(videoContent, utils.DetectVideoMimeType(videoContent), "video")
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "send video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}

		q.Robot.saveRecord(videoContent, imageContent, param.VideoRecordType, totalToken)
	})

}

func (q *QQRobot) sendChatMessage() {
	q.Robot.TalkingPreCheck(func() {
		q.executeLLM()
	})

}

func (q *QQRobot) executeLLM() {
	q.Robot.runLLM(q, "silk", true)
}

type UploadFileRequest struct {
	FileType   int    `json:"file_type"`
	URL        string `json:"url"`
	SrvSendMsg bool   `json:"srv_send_msg"`
	FileData   string `json:"file_data"`
}

type FileData struct {
	FileUUID string `json:"file_uuid"`
	FileInfo []byte `json:"file_info"`
	TTL      int    `json:"ttl"`
	ID       string `json:"id"`
}

func (q *QQRobot) UploadFile(imageContent string, fileType int) ([]byte, error) {
	chatId, _, _ := q.Robot.GetChatIdAndMsgIdAndUserID()

	apiURL := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/files", chatId)
	if q.ATMessage != nil {
		apiURL = fmt.Sprintf("https://api.sgroup.qq.com/v2/groups/%s/files", chatId)
	}
	if q.GroupAtMessage != nil {
		apiURL = fmt.Sprintf("https://api.sgroup.qq.com/v2/groups/%s/files", chatId)
	}

	reqBody := UploadFileRequest{
		FileType:   fileType,
		FileData:   imageContent,
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
	req.Header.Set("Authorization", "QQBot "+tokenInfo.AccessToken) // 如果需要鉴权的话

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	fileData := new(FileData)
	err = json.Unmarshal(respData, fileData)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	if len(fileData.FileInfo) == 0 {
		return nil, errors.New("file info is empty" + string(respData))
	}

	return fileData.FileInfo, err
}

func (q *QQRobot) GetAttachment() *dto.MessageAttachment {
	if q.C2CMessage != nil && len(q.C2CMessage.Attachments) != 0 {
		return q.C2CMessage.Attachments[0]
	}

	if q.ATMessage != nil && len(q.ATMessage.Attachments) != 0 {
		return q.ATMessage.Attachments[0]
	}

	if q.GroupAtMessage != nil && len(q.GroupAtMessage.Attachments) != 0 {
		return q.GroupAtMessage.Attachments[0]
	}

	return nil
}

func (q *QQRobot) PostRichMediaMessage(data []byte, content string) error {
	_, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	msgType := dto.TextMsg
	if len(data) > 0 {
		msgType = dto.RichMediaMsg
	}

	var err error
	if q.C2CMessage != nil {
		_, err = q.QQApi.PostC2CMessage(q.Robot.Ctx, q.C2CMessage.Author.ID, &dto.MessageToCreate{
			MsgType: msgType,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
			Content: content,
		})
	}

	if q.ATMessage != nil {
		_, err = q.QQApi.PostMessage(q.Robot.Ctx, q.ATMessage.GuildID, &dto.MessageToCreate{
			MsgType: msgType,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
			Content: content,
		})
	}

	if q.GroupAtMessage != nil {
		_, err = q.QQApi.PostGroupMessage(q.Robot.Ctx, q.GroupAtMessage.GroupID, &dto.MessageToCreate{
			MsgType: msgType,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
			Content: content,
		})
	}

	return err

}

func (q *QQRobot) PostStreamMessage(state, idx int32, id, content string) (string, error) {
	_, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	msg := &dto.MessageToCreate{
		MsgType: dto.TextMsg,
		MsgID:   msgId,
		Content: content,
		MsgSeq:  crc32.ChecksumIEEE([]byte(content)),
		Stream: &dto.Stream{
			State: state,
			Index: idx,
			ID:    id,
		},
	}

	if q.C2CMessage != nil {
		resp, err := q.QQApi.PostC2CMessage(q.Robot.Ctx, q.C2CMessage.Author.ID, msg)
		if err != nil {
			return "", err
		}
		return resp.ID, err
	}

	if q.GroupAtMessage != nil {
		resp, err := q.QQApi.PostGroupMessage(q.Robot.Ctx, q.GroupAtMessage.GroupID, msg)
		if err != nil {
			return "", err
		}
		return resp.ID, err
	}

	if q.ATMessage != nil {
		resp, err := q.QQApi.PostGroupMessage(q.Robot.Ctx, q.ATMessage.GuildID, msg)
		if err != nil {
			return "", err
		}
		return resp.ID, err
	}

	return "", errors.New("don't get message")

}

func (q *QQRobot) getPerMsgLen() int {
	return 1800
}

func (q *QQRobot) sendTextStream(messageChan *MsgChan) {
	var id string
	var err error
	idx := int32(0)

	for msg := range messageChan.StrMessageChan {
		id, err = q.PostStreamMessage(1, idx, id, msg)
		if err != nil {
			logger.ErrorCtx(q.Robot.Ctx, "send stream msg fail", "err", err)
		}
		idx++
	}

	_, err = q.PostStreamMessage(10, idx, id, " ")
	if err != nil {
		logger.ErrorCtx(q.Robot.Ctx, "send stream msg fail", "err", err)
	}
}

func (q *QQRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	base64Content := base64.StdEncoding.EncodeToString(voiceContent)
	data, err := q.UploadFile(base64Content, 3)
	if err != nil {
		logger.ErrorCtx(q.Robot.Ctx, "upload file fail", "err", err)
		return err
	}

	return q.PostRichMediaMessage(data, "")
}

func init() {
	RegisterChannel(Channel{
		Name: param.QQ,
		MissingCfg: func() []string {
			var missing []string
			missing = required(missing, "QQ_APP_ID", conf.BaseConfInfo.QQAppID)
			return required(missing, "QQ_APP_SECRET", conf.BaseConfInfo.QQAppSecret)
		},
		Start: StartQQRobot,
	})
}
