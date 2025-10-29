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
	"os"
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
	"github.com/yincongcyincong/MuseBot/db"
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
	event         *dto.WSPayload
	Robot         *RobotInfo
	QQApi         openapi.OpenAPI
	QQTokenSource oauth2.TokenSource
	RobotInfo     *dto.User
	
	C2CMessage     *dto.WSC2CMessageData
	GroupAtMessage *dto.WSGroupATMessageData
	ATMessage      *dto.WSATMessageData
	
	Command      string
	Prompt       string
	BotName      string
	OriginPrompt string
	ImageContent []byte
}

func StartQQRobot(ctx context.Context) {
	var err error
	botgo.SetLogger(logger.QQLogger)
	QQTokenSource = token.NewQQBotTokenSource(
		&token.QQBotCredentials{
			AppID:     *conf.BaseConfInfo.QQAppID,
			AppSecret: *conf.BaseConfInfo.QQAppSecret,
		})
	if err = token.StartRefreshAccessToken(ctx, QQTokenSource); err != nil {
		logger.Error("start refresh access token error", "err", err)
		return
	}
	
	QQApi = botgo.NewOpenAPI(*conf.BaseConfInfo.QQAppID, QQTokenSource).WithTimeout(5 * time.Second).SetDebug(false)
	QQRobotInfo, err = QQApi.Me(ctx)
	if err != nil {
		logger.Error("get me error", "err", err)
		return
	}
	logger.Info("QQRobot Info", "username", QQRobotInfo.Username)
	
	event.RegisterHandlers(
		event.GroupATMessageEventHandler(QQGroupATMessageEventHandler),
		event.ATMessageEventHandler(QQATMessageEventHandler),
		event.C2CMessageEventHandler(C2CMessageEventHandler),
	)
}

func C2CMessageEventHandler(event *dto.WSPayload, message *dto.WSC2CMessageData) error {
	d := NewQQRobot(event, message, nil, nil)
	d.Robot = NewRobot(WithRobot(d), WithTencentRobot(d))
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

func QQGroupATMessageEventHandler(event *dto.WSPayload, atMessage *dto.WSGroupATMessageData) error {
	d := NewQQRobot(event, nil, atMessage, nil)
	d.Robot = NewRobot(WithRobot(d), WithTencentRobot(d))
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
	d := NewQQRobot(event, nil, nil, atMessage)
	d.Robot = NewRobot(WithRobot(d), WithTencentRobot(d))
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

func NewQQRobot(event *dto.WSPayload, c2cMessage *dto.WSC2CMessageData,
	atGroupMessage *dto.WSGroupATMessageData, atMessage *dto.WSATMessageData) *QQRobot {
	metrics.AppRequestCount.WithLabelValues("qq").Inc()
	return &QQRobot{
		QQApi:          QQApi,
		QQTokenSource:  QQTokenSource,
		RobotInfo:      QQRobotInfo,
		C2CMessage:     c2cMessage,
		ATMessage:      atMessage,
		GroupAtMessage: atGroupMessage,
		event:          event,
		BotName:        BotName,
	}
}

func (q *QQRobot) checkValid() bool {
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
		_, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
			if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
				if msgInfo.Status == msgChangePhoto || msgInfo.Status == msgRecognizePhoto {
					logger.Info("ComWechatRobot handle photo msg", "msgId", msgId, "userId", userId)
					q.passiveExecCmd()
					return false
				}
			}
		}
	}
	
	if q.GetAttachment() != nil && strings.Contains(q.GetAttachment().ContentType, "voice") {
		_, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
			if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
				if msgInfo.Status == msgSaveVoice {
					logger.Info("ComWechatRobot handle voice msg", "msgId", msgId, "userId", userId)
					q.passiveExecCmd()
					return false
				}
			}
		}
	}
	
	return true
}

func (q *QQRobot) getMsgContent() string {
	return q.Command
}

func (q *QQRobot) requestLLMAndResp(content string) {
	if !strings.Contains(content, "/") && !strings.Contains(content, "$") && q.Prompt == "" {
		q.Prompt = content
	}
	q.Robot.ExecCmd(content, q.sendChatMessage, nil, nil)
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
		
		if len(lastImageContent) == 0 && strings.Contains(q.Command, "edit_photo") {
			lastImageContent, err = q.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		imageContent, totalToken, err := q.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectImageFormat(imageContent)
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		
		data, err := q.UploadFile(base64Content, 1)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = q.PostRichMediaMessage(data)
		if err != nil {
			logger.Warn("post rich media msg fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
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
			Mode:       utils.GetImgType(db.GetCtxUserInfo(q.Robot.Ctx).LLMConfigRaw),
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
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
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
		
		videoContent, totalToken, err := q.Robot.CreateVideo(prompt, imageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectVideoMimeType(videoContent)
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		data, err := q.UploadFile(base64Content, 2)
		if err != nil {
			logger.Warn("upload file fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		err = q.PostRichMediaMessage(data)
		if err != nil {
			logger.Warn("post rich media msg fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		originImageURI := ""
		if len(imageContent) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(imageContent)
			format = utils.DetectImageFormat(imageContent)
			originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
		}
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   q.Prompt,
			Answer:     dataURI,
			Token:      totalToken,
			Content:    originImageURI,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       utils.GetVideoType(db.GetCtxUserInfo(q.Robot.Ctx).LLMConfigRaw),
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
	var msgChan *MsgChan
	if q.C2CMessage != nil {
		msgChan = &MsgChan{
			StrMessageChan: make(chan string),
		}
	} else {
		msgChan = &MsgChan{
			NormalMessageChan: make(chan *param.MsgInfo),
		}
	}
	go q.Robot.ExecChain(q.Prompt, msgChan)
	
	// send response message
	go q.Robot.HandleUpdate(msgChan, "silk")
}

func (q *QQRobot) executeLLM() {
	var msgChan *MsgChan
	if q.C2CMessage != nil && *conf.AudioConfInfo.TTSType == "" {
		msgChan = &MsgChan{
			StrMessageChan: make(chan string),
		}
	} else {
		msgChan = &MsgChan{
			NormalMessageChan: make(chan *param.MsgInfo),
		}
	}
	
	go q.Robot.HandleUpdate(msgChan, "silk")
	
	go q.Robot.ExecLLM(q.Prompt, msgChan)
	
}

func (q *QQRobot) getContent(content string) (string, error) {
	attachment := q.GetAttachment()
	if attachment == nil {
		return content, nil
	}
	
	switch {
	case strings.Contains(attachment.ContentType, "image"):
		data, err := utils.DownloadFile(attachment.URL)
		if err != nil {
			logger.Error("get image content fail", "err", err)
			return "", err
		}
		
		content, err = q.Robot.GetImageContent(data, content)
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
		
		data, err = utils.SilkToWav(data)
		if err != nil {
			logger.Error("silk to wav fail", "err", err)
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

func (q *QQRobot) PostRichMediaMessage(data []byte) error {
	_, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	var err error
	if q.C2CMessage != nil {
		_, err = q.QQApi.PostC2CMessage(q.Robot.Ctx, q.C2CMessage.Author.ID, &dto.MessageToCreate{
			MsgType: dto.RichMediaMsg,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
		})
	}
	
	if q.ATMessage != nil {
		_, err = q.QQApi.PostMessage(q.Robot.Ctx, q.ATMessage.GuildID, &dto.MessageToCreate{
			MsgType: dto.RichMediaMsg,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
		})
	}
	
	if q.GroupAtMessage != nil {
		_, err = q.QQApi.PostGroupMessage(q.Robot.Ctx, q.GroupAtMessage.GroupID, &dto.MessageToCreate{
			MsgType: dto.RichMediaMsg,
			MsgID:   msgId,
			Media: &dto.MediaInfo{
				FileInfo: data,
			},
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

func (q *QQRobot) passiveExecCmd() {
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		if !*conf.BaseConfInfo.WechatActive {
			logger.Warn("only wechat_active is true can generate image")
			q.Robot.SendMsg(chatId, "only wechat_active is true can generate image", msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		attachment := q.GetAttachment()
		
		if attachment == nil {
			status := msgChangePhoto
			switch q.Command {
			case "/change_photo", "change_photo", "$change_photo":
				status = msgChangePhoto
			case "/rec_photo", "rec_photo", "$rec_photo":
				status = msgRecognizePhoto
			case "/save_voice", "save_voice", "$save_voice":
				status = msgSaveVoice
			}
			TencentMsgMap.Store(userId, &TencentWechatMessage{
				Msg:       q.Prompt,
				Status:    status,
				StartTime: time.Now(),
			})
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_pre_prompt_success", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		if strings.Contains(attachment.ContentType, "image") {
			if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
				if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
					switch msgInfo.Status {
					case msgChangePhoto:
						data, err := utils.DownloadFile(attachment.URL)
						if err != nil {
							logger.Error("get image content fail", "err", err)
							q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						q.Prompt = msgInfo.Msg
						q.ImageContent = data
						q.sendImg()
					case msgRecognizePhoto:
						q.Prompt = msgInfo.Msg
						q.executeLLM()
					}
					
					TencentMsgMap.Delete(userId)
				}
			}
		}
		
		if strings.Contains(attachment.ContentType, "voice") {
			if msgInfoInter, ok := TencentMsgMap.Load(userId); ok {
				if msgInfo, ok := msgInfoInter.(*TencentWechatMessage); ok {
					switch msgInfo.Status {
					case msgSaveVoice:
						data, err := utils.DownloadFile(attachment.URL)
						if err != nil {
							logger.Error("get image content fail", "err", err)
							q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						
						data, err = utils.SilkToMp3(data)
						if err != nil {
							logger.Error("silk to wav fail", "err", err)
							q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						
						fileName := utils.GetAbsPath("data/" + utils.RandomFilename(utils.DetectAudioFormat(data)))
						err = os.WriteFile(fileName, data, 0666)
						if err != nil {
							logger.Error("save image fail", "err", err)
							q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
							return
						}
						
						q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "save_audio_success", map[string]interface{}{
							"filename": fileName,
						}), msgId, tgbotapi.ModeMarkdown, nil)
					}
					
					TencentMsgMap.Delete(userId)
				}
			}
		}
	})
}

func (q *QQRobot) sendText(messageChan *MsgChan) {
	chatId, msgId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	if messageChan.NormalMessageChan != nil {
		var msg *param.MsgInfo
		for msg = range messageChan.NormalMessageChan {
			if msg.Finished {
				q.Robot.SendMsg(chatId, msg.Content, msgId, "", nil)
			}
		}
		
		if msg != nil {
			q.Robot.SendMsg(chatId, msg.Content, msgId, "", nil)
		}
	} else {
		var id string
		var err error
		idx := int32(0)
		
		for msg := range messageChan.StrMessageChan {
			id, err = q.PostStreamMessage(1, idx, id, msg)
			if err != nil {
				logger.Warn("send stream msg fail", "err", err)
			}
			idx++
		}
		
		_, err = q.PostStreamMessage(10, idx, id, " ")
		if err != nil {
			logger.Warn("send stream msg fail", "err", err)
		}
	}
}

func (q *QQRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	base64Content := base64.StdEncoding.EncodeToString(voiceContent)
	data, err := q.UploadFile(base64Content, 3)
	if err != nil {
		logger.Warn("upload file fail", "err", err)
		return err
	}
	
	return q.PostRichMediaMessage(data)
}
