package robot

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type Web struct {
	Command      string
	UserId       int64
	RealUserId   string
	Prompt       string
	ImageContent []byte
	AudioContent []byte
	cs           *param.ContextState
	
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
		W:              w,
		Flusher:        flusher,
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
	
	web.Robot = NewRobot(WithRobot(web), WithSkipCheck(true))
	return web
}

func (web *Web) getMsgContent() string {
	return web.Command
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
	base64Content := base64.StdEncoding.EncodeToString(imageContent)
	format := utils.DetectImageFormat(imageContent)
	dataURI := fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
	
	web.sendMedia(imageContent, format, "image")
	
	originImageURI := ""
	if len(web.ImageContent) > 0 {
		base64Content = base64.StdEncoding.EncodeToString(web.ImageContent)
		format = utils.DetectImageFormat(web.ImageContent)
		originImageURI = fmt.Sprintf("data:image/%s;base64,%s", format, base64Content)
	}
	
	// save message record
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     dataURI,
		Content:    originImageURI,
		Token:      0,
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
		Mode:       utils.GetImgType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})
	
	// save data record
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     dataURI,
		Content:    originImageURI,
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
	
	base64Content := base64.StdEncoding.EncodeToString(videoContent)
	format := utils.DetectVideoMimeType(videoContent)
	dataURI := fmt.Sprintf("data:video/%s;base64,%s", format, base64Content)
	
	web.sendMedia(videoContent, format, "video")
	
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     dataURI,
		Token:      0,
		IsDeleted:  0,
		RecordType: param.WEBRecordType,
		Mode:       utils.GetVideoType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})
	
	db.InsertRecordInfo(web.Robot.Ctx, &db.Record{
		UserId:     web.RealUserId,
		Question:   web.OriginalPrompt,
		Answer:     dataURI,
		Token:      totalToken,
		IsDeleted:  0,
		RecordType: param.VideoRecordType,
		Mode:       utils.GetVideoType(db.GetCtxUserInfo(web.Robot.Ctx).LLMConfigRaw),
	})
	
}

func (web *Web) sendChatMessage() {
	
	prompt := web.Prompt
	
	messageChan := make(chan string)
	l := llm.NewLLM(
		llm.WithChatId(web.RealUserId),
		llm.WithUserId(web.RealUserId),
		llm.WithMsgId(web.RealUserId),
		llm.WithHTTPMsgChan(messageChan),
		llm.WithContent(prompt),
		llm.WithPerMsgLen(1000000),
		llm.WithCS(web.cs),
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
	if web.ImageContent != nil {
		originDataURI = fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(web.ImageContent), web.ImageContent)
	} else if web.AudioContent != nil {
		originDataURI = fmt.Sprintf("data:audio/%s;base64,%s", utils.DetectAudioFormat(web.AudioContent), web.AudioContent)
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
	_, err := web.W.Write([]byte(msgContent))
	if err != nil {
		logger.ErrorCtx(web.Robot.Ctx, "send message fail", "err", err)
	}
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

func (web *Web) setPrompt(prompt string) {
	web.Prompt = prompt
}

func (web *Web) getAudio() []byte {
	return web.AudioContent
}

func (web *Web) getImage() []byte {
	return web.ImageContent
}

func (web *Web) setImage(image []byte) {
	web.ImageContent = image
}

func (web *Web) checkValid() bool {
	var err error
	web.Prompt, err = web.Robot.GetAudioContent(web.AudioContent)
	if err != nil {
		logger.WarnCtx(web.Robot.Ctx, "generate text from audio failed", "err", err)
		return false
	}
	return true
}

func (web *Web) requestLLM(content string) {
	web.Robot.ExecCmd(content, web.sendChatMessage, nil, nil)
}

func (web *Web) getPrompt() string {
	return web.Prompt
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

func (web *Web) setCommand(command string) {
	web.Command = command
}

func (web *Web) getCommand() string {
	return web.Command
}

func (web *Web) getUserName() string {
	return web.RealUserId
}

func (web *Web) executeLLM() {

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
