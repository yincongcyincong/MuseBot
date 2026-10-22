package robot

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type Web struct {
	UserId     int64
	RealUserId string

	cs *param.ContextState

	OriginalPrompt string

	W       http.ResponseWriter
	Flusher http.Flusher

	Robot *RobotInfo

	baseRobot
}

func NewWeb(command string, userId int64, realUserId, prompt, originalPrompt, fileName string, bodyData []byte, w http.ResponseWriter, flusher http.Flusher) *Web {
	metrics.AppRequestCount.WithLabelValues("web").Inc()
	web := &Web{
		UserId:         userId,
		RealUserId:     realUserId,
		OriginalPrompt: originalPrompt,
		baseRobot: baseRobot{
			Command: command,
			Prompt:  prompt,
		},
		W:       w,
		Flusher: flusher,
		cs: &param.ContextState{
			UseRecord: true,
		},
	}

	if utils.DetectImageFormat(bodyData) != "unknown" {
		web.ImageContent = bodyData
	}

	if utils.DetectAudioFormat(bodyData) != "unknown" {
		web.AudioContent = bodyData
	}

	if web.ImageContent == nil && web.AudioContent == nil &&
		utils.DetectDocumentType(fileName, bodyData) != "" {
		web.DocContent = &param.DocInfo{FileName: fileName, Content: bodyData}
	}

	web.Robot = NewRobot(WithRobot(web), WithSkipCheck(true))
	return web
}

func (web *Web) getMsgContent() string {
	return web.Prompt
}

func (web *Web) sendImg() {

	prompt := strings.TrimSpace(web.Prompt)
	if prompt == "" {
		logger.WarnCtx(web.Robot.Ctx, "prompt is empty")
		web.SendMsg(i18n.GetMessage("photo_empty_content", nil))
		return
	}

	lastImageContent := web.ImageContent
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
	format := utils.DetectImageFormat(imageContent)
	web.sendMedia(imageContent, format, "image")

	answerPath := web.Robot.saveMediaContent(imageContent, param.ImageRecordType)
	originImagePath := ""
	if len(web.ImageContent) > 0 {
		originImagePath = web.Robot.saveMediaContent(web.ImageContent, param.ImageRecordType)
	}

	// save message record
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     answerPath,
		Content:    originImagePath,
		Token:      0,
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
		Mode:       utils.GetImgType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})

	// save data record
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     answerPath,
		Content:    originImagePath,
		Token:      totalToken,
		IsDeleted:  0,
		RecordType: param.ImageRecordType,
		Mode:       utils.GetImgType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})

}

func (web *Web) sendVideo() {
	prompt := strings.TrimSpace(web.Prompt)
	if prompt == "" {
		logger.WarnCtx(web.Robot.Ctx, "prompt is empty")
		web.SendMsg(i18n.GetMessage("video_empty_content", nil))
		return
	}

	videoContent, totalToken, err := web.Robot.CreateVideo(prompt, web.ImageContent)
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "generate video fail", "err", err)
		web.SendMsg(err.Error())
		return
	}

	format := utils.DetectVideoMimeType(videoContent)
	web.sendMedia(videoContent, format, "video")

	answerPath := web.Robot.saveMediaContent(videoContent, param.VideoRecordType)

	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     answerPath,
		Token:      0,
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
		Mode:       utils.GetVideoType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})

	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     answerPath,
		Token:      totalToken,
		IsDeleted:  0,
		RecordType: param.VideoRecordType,
		Mode:       utils.GetVideoType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})

}

func (web *Web) sendChatMessage() {

	web.Robot.TalkingPreCheck(func() {
		web.executeLLM()
	})

}

func (web *Web) sendTextStream(messageChan *MsgChan) {
	totalContent := ""
	for msg := range messageChan.StrMessageChan {
		fmt.Fprintf(web.W, "%s", msg)
		totalContent += msg
		web.Flusher.Flush()
	}

	originDataURI := ""
	if web.ImageContent != nil {
		originDataURI = web.Robot.saveMediaContent(web.ImageContent, param.ImageRecordType)
	} else if web.AudioContent != nil {
		originDataURI = fmt.Sprintf("data:audio/%s;base64,%s", utils.DetectAudioFormat(web.AudioContent),
			base64.StdEncoding.EncodeToString(web.AudioContent))
	}

	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     totalContent,
		Content:    originDataURI,
		Token:      0, // llm already calculate it
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
	})
}

func (web *Web) SendMsg(msgContent string) {
	fmt.Fprint(web.W, msgContent)
	web.Flusher.Flush()
}

func (web *Web) InsertRecord() {
	id, err := db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.Prompt,
		RecordType: param.TextRecordType,
	})
	if err != nil {
		logger.ErrorCtx(web.Robot.Ctx, "insert record fail", "err", err)
		return
	}

	web.cs.RecordID = id
}

func (web *Web) AddUserInfo() bool {
	userInfo, err := db.GetUserByID(web.RealUserId)
	if err != nil {
		logger.ErrorCtx(web.Robot.Ctx, "addUserInfo GetUserByID err", "err", err)
		return false
	}

	if userInfo == nil || userInfo.ID == 0 {
		_, err = db.InsertUser(web.RealUserId, utils.GetDefaultLLMConfig())
		if err != nil {
			logger.ErrorCtx(web.Robot.Ctx, "insert user fail", "userID", web.RealUserId, "err", err)
			return false
		}

		userInfo, err = db.GetUserByID(web.RealUserId)
		if err != nil || userInfo == nil {
			logger.ErrorCtx(web.Robot.Ctx, "addUserInfo GetUserByID err", "err", err)
			return false
		}
	}

	web.Robot.Ctx = context.WithValue(web.Robot.Ctx, "user_info", userInfo)

	return true
}

func (web *Web) checkValid() bool {
	var err error
	if web.AudioContent != nil {
		web.OriginalPrompt, err = web.Robot.GetAudioContent(web.AudioContent)
		if err != nil {
			logger.ErrorCtx(web.Robot.Ctx, "generate text from audio failed", "err", err)
			return false
		}
	}

	web.Command, web.Prompt = ParseCommand(web.OriginalPrompt)
	return true
}

func (web *Web) requestLLM(content string) {
	web.Robot.ExecCmd(web.Command, web.sendChatMessage, nil, nil)
}

func (web *Web) getPerMsgLen() int {
	return 1000000
}

func (web *Web) sendVoiceContent(voiceContent []byte, duration int) error {
	fmt.Fprintf(web.W, "%s", fmt.Sprintf("data:audio/%s;base64,%s", utils.DetectAudioFormat(voiceContent),
		base64.StdEncoding.EncodeToString(voiceContent)))
	web.Flusher.Flush()
	return nil
}

func (web *Web) getUserName() string {
	return web.RealUserId
}

func (web *Web) executeLLM() {
	var msgChan *MsgChan
	if conf.BaseConfInfo.IsStreaming {
		msgChan = &MsgChan{
			StrMessageChan: make(chan string),
		}
	} else {
		msgChan = &MsgChan{
			NormalMessageChan: make(chan *param.MsgInfo),
		}
	}

	go web.Robot.ExecLLM(web.Prompt, msgChan)

	web.Robot.HandleUpdate(msgChan, "mp3")

}

func (web *Web) sendMedia(media []byte, contentType, sType string) error {
	if sType == "image" {
		fmt.Fprintf(web.W, "%s", fmt.Sprintf("data:image/%s;base64,%s", contentType,
			base64.StdEncoding.EncodeToString(media)))
	} else {
		fmt.Fprintf(web.W, "%s", fmt.Sprintf("data:video/%s;base64,%s", contentType,
			base64.StdEncoding.EncodeToString(media)))
	}

	web.Flusher.Flush()

	return nil
}
