package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type QQMessage struct {
	Font          int           `json:"font"`
	GroupId       int64         `json:"group_id"`
	Message       []MessageItem `json:"message"`
	MessageFormat string        `json:"message_format"`
	MessageID     int64         `json:"message_id"`
	MessageSeq    int64         `json:"message_seq"`
	MessageType   string        `json:"message_type"`
	PostType      string        `json:"post_type"`
	Raw           RawMessage    `json:"raw"`
	RawMessage    string        `json:"raw_message"`
	RealID        int64         `json:"real_id"`
	RealSeq       string        `json:"real_seq"`
	SelfID        int64         `json:"self_id"`
	Sender        SenderInfo    `json:"sender"`
	SubType       string        `json:"sub_type"`
	TargetID      int64         `json:"target_id"`
	Time          int64         `json:"time"`
	UserID        int64         `json:"user_id"`
}

type MessageItem struct {
	Data struct {
		Text string `json:"text"`
		Url  string `json:"url"`
		QQ   string `json:"qq"`
	} `json:"data"`
	Type string `json:"type"`
}

type SenderInfo struct {
	Card     string `json:"card"`
	Nickname string `json:"nickname"`
	UserID   int64  `json:"user_id"`
}

// RawMessage 部分嵌套字段很多，以下是主要结构（可按需精简）
type RawMessage struct {
	ID             int64     `json:"id"`
	MsgID          string    `json:"msgId"`
	MsgSeq         string    `json:"msgSeq"`
	MsgTime        string    `json:"msgTime"`
	MsgType        int       `json:"msgType"`
	SubMsgType     int       `json:"subMsgType"`
	PeerUin        string    `json:"peerUin"`
	SenderUin      string    `json:"senderUin"`
	SenderUid      string    `json:"senderUid"`
	PeerUid        string    `json:"peerUid"`
	ClientSeq      string    `json:"clientSeq"`
	MsgRandom      string    `json:"msgRandom"`
	SendStatus     int       `json:"sendStatus"`
	SendType       int       `json:"sendType"`
	ChatType       int       `json:"chatType"`
	IsOnlineMsg    bool      `json:"isOnlineMsg"`
	IsImportMsg    bool      `json:"isImportMsg"`
	SourceType     int       `json:"sourceType"`
	FromAppID      string    `json:"fromAppid"`
	FromUid        string    `json:"fromUid"`
	PeerName       string    `json:"peerName"`
	SendNickName   string    `json:"sendNickName"`
	SendMemberName string    `json:"sendMemberName"`
	SendRemarkName string    `json:"sendRemarkName"`
	Elements       []Element `json:"elements"`
}

type Element struct {
	ElementID   string       `json:"elementId"`
	ElementType int          `json:"elementType"`
	TextElement *TextElement `json:"textElement,omitempty"`
}

type TextElement struct {
	Content string `json:"content"`
}

type PersonalQQRobot struct {
	Msg   *QQMessage
	Robot *RobotInfo
	Ctx   context.Context
	
	Command      string
	Prompt       string
	OriginPrompt string
	ImageContent []byte
	AudioContent []byte
}

func NewPersonalQQRobot(ctx context.Context, msgContent []byte) *PersonalQQRobot {
	msg := new(QQMessage)
	err := json.Unmarshal(msgContent, msg)
	if err != nil {
		logger.ErrorCtx(ctx, "Unmarshal QQMessage error", "error", err)
		return nil
	}
	
	q := &PersonalQQRobot{
		Msg: msg,
		Ctx: ctx,
	}
	
	q.Robot = NewRobot(WithRobot(q))
	return q
}

func (q *PersonalQQRobot) checkValid() bool {
	if q.Msg.Message == nil {
		return false
	}
	
	atBot, err := q.GetMessageContent()
	if err != nil {
		logger.ErrorCtx(q.Ctx, "get message content error", "err", err)
		return false
	}
	
	if !atBot && q.Msg.MessageType == "group" {
		return false
	}
	
	return true
}

func (q *PersonalQQRobot) getMsgContent() string {
	return q.Command
}

func (q *PersonalQQRobot) requestLLMAndResp(content string) {
	if !strings.Contains(content, "/") && q.Prompt == "" {
		q.Prompt = content
	}
	q.Robot.ExecCmd(content, q.sendChatMessage, nil, nil)
}

func (q *PersonalQQRobot) sendImg() {
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		var err error
		var lastImageContent = q.ImageContent
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
		
		_, err = q.SendMsg("", imageContent, nil, nil)
		if err != nil {
			logger.Warn("send image fail", "err", err)
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

func (q *PersonalQQRobot) sendVideo() {
	// 检查 prompt
	q.Robot.TalkingPreCheck(func() {
		chatId, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			logger.Warn("prompt is empty")
			q.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		videoContent, totalToken, err := q.Robot.CreateVideo(prompt, q.ImageContent)
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		_, err = q.SendMsg("", nil, videoContent, nil)
		if err != nil {
			logger.Warn("send video fail", "err", err)
			q.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
		format := utils.DetectVideoMimeType(videoContent)
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
		
		originImageURI := ""
		if len(q.ImageContent) > 0 {
			base64Content = base64.StdEncoding.EncodeToString(q.ImageContent)
			format = utils.DetectImageFormat(q.ImageContent)
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

func (q *PersonalQQRobot) sendChatMessage() {
	q.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			q.executeChain()
		} else {
			q.executeLLM()
		}
	})
	
}

func (q *PersonalQQRobot) executeChain() {
	var msgChan *MsgChan
	if *conf.AudioConfInfo.TTSType == "" {
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
	go q.Robot.HandleUpdate(msgChan, "mp3")
}

func (q *PersonalQQRobot) executeLLM() {
	msgChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	
	go q.Robot.HandleUpdate(msgChan, "mp3")
	
	go q.Robot.ExecLLM(q.Prompt, msgChan)
	
}

func (q *PersonalQQRobot) getContent(content string) (string, error) {
	
	var err error
	if q.ImageContent != nil {
		content, err = q.Robot.GetImageContent(q.ImageContent, content)
		if err != nil {
			logger.Warn("generate text from audio failed", "err", err)
			return "", err
		}
	}
	
	if q.AudioContent != nil {
		content, err = q.Robot.GetAudioContent(q.AudioContent)
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

func (q *PersonalQQRobot) getPrompt() string {
	return q.Prompt
}

func (q *PersonalQQRobot) getPerMsgLen() int {
	return 1800
}

func (q *PersonalQQRobot) sendText(messageChan *MsgChan) {
	chatId, messageId, _ := q.Robot.GetChatIdAndMsgIdAndUserID()
	
	var msg *param.MsgInfo
	for msg = range messageChan.NormalMessageChan {
		if msg.Finished {
			q.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
		}
	}
	
	if msg == nil || len(msg.Content) == 0 {
		msg = new(param.MsgInfo)
		return
	}
	
	q.Robot.SendMsg(chatId, msg.Content, messageId, "", nil)
}

type OneBotResult struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

func (q *PersonalQQRobot) SendMsg(txt string, image []byte, video []byte, voice []byte) (string, error) {
	_, msgId, userId := q.Robot.GetChatIdAndMsgIdAndUserID()
	
	msgArray := []map[string]interface{}{}
	
	if txt != "" {
		msgArray = append(msgArray, map[string]interface{}{
			"type": "reply",
			"data": map[string]string{"id": msgId},
		})
		msgArray = append(msgArray, map[string]interface{}{
			"type": "text",
			"data": map[string]string{"text": txt},
		})
	}
	
	if len(image) > 0 {
		msgArray = append(msgArray, map[string]interface{}{
			"type": "reply",
			"data": map[string]string{"id": msgId},
		})
		
		encoded := "base64://" + base64.StdEncoding.EncodeToString(image)
		msgArray = append(msgArray, map[string]interface{}{
			"type": "image",
			"data": map[string]string{"file": encoded},
		})
	}
	
	if len(video) > 0 {
		encoded := "base64://" + base64.StdEncoding.EncodeToString(video)
		msgArray = append(msgArray, map[string]interface{}{
			"type": "video",
			"data": map[string]string{"file": encoded},
		})
	}
	
	if len(voice) > 0 {
		encoded := "base64://" + base64.StdEncoding.EncodeToString(voice)
		msgArray = append(msgArray, map[string]interface{}{
			"type": "record",
			"data": map[string]string{"file": encoded},
		})
	}
	
	if len(msgArray) == 0 {
		return "", fmt.Errorf("no content")
	}
	
	payload := map[string]interface{}{
		"message": msgArray,
	}
	path := "/send_private_msg"
	if q.Msg.MessageType == "group" {
		path = "/send_group_msg"
		payload["group_id"] = q.Msg.GroupId
	} else {
		payload["user_id"] = userId
	}
	
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", strings.TrimRight(*conf.BaseConfInfo.QQOneBotHttpServer, "/")+
		path, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if *conf.BaseConfInfo.QQOneBotSendToken != "" {
		req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.QQOneBotSendToken)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorCtx(q.Ctx, "send message failed", "err", err, "req", payload)
		return "", err
	}
	defer resp.Body.Close()
	
	// 解析返回值
	result := new(OneBotResult)
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.ErrorCtx(q.Ctx, "send message failed", "err", err, "req", payload)
		return "", err
	}
	
	if result.Status != "ok" && result.MessageID == "" {
		logger.ErrorCtx(q.Ctx, "send message failed", "err", err, "result", result)
		return "", fmt.Errorf("send fail: %v", result)
	}
	
	return result.MessageID, nil
}

func (q *PersonalQQRobot) GetMessageContent() (bool, error) {
	var err error
	prompt := ""
	isAt := false
	for _, v := range q.Msg.Message {
		switch v.Type {
		case "text":
			prompt += v.Data.Text
		case "image":
			q.ImageContent, err = utils.DownloadFile(v.Data.Url)
			if err != nil {
				return false, err
			}
		case "record":
			q.AudioContent, err = utils.DownloadFile(v.Data.Url)
			if err != nil {
				return false, err
			}
		case "at":
			isAt = v.Data.QQ == strconv.Itoa(int(q.Msg.SelfID))
		}
	}
	q.Command, q.Prompt = ParseCommand(prompt)
	if q.Command == "" && q.Prompt == "" && q.ImageContent == nil && q.AudioContent == nil {
		return false, fmt.Errorf("no content")
	}
	
	return isAt, nil
}

func (q *PersonalQQRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	_, err := q.SendMsg("", nil, nil, voiceContent)
	return err
}
