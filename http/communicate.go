package http

import (
	"io"
	"net/http"
	"strconv"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/tencent-connect/botgo/interaction/webhook"
	"github.com/tencent-connect/botgo/token"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/robot"
)

// Communicate handles the Server-Sent Events
func Communicate(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Error reading request body", "err", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	
	if prompt == "" && len(fileData) == 0 {
		http.Error(w, "Missing prompt parameter", http.StatusBadRequest)
		return
	}
	
	realUserId := r.URL.Query().Get("user_id")
	intUserId, _ := strconv.ParseInt(realUserId, 10, 64)
	
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	command, p := robot.ParseCommand(prompt)
	
	web := robot.NewWeb(command, intUserId, realUserId, p, prompt, fileData, w, flusher)
	web.Exec()
}

func ComWechatComm(w http.ResponseWriter, r *http.Request) {
	var rs *http.Response
	var err error
	if r.Method == http.MethodGet {
		rs, err = robot.ComWechatApp.Server.VerifyURL(r)
		if err != nil {
			logger.Error("verify url fail", "err", err)
		}
	} else {
		rs, err = robot.ComWechatApp.Server.Notify(r, func(event contract.EventInterface) interface{} {
			
			c := robot.NewComWechatRobot(event)
			c.Robot = robot.NewRobot(robot.WithRobot(c))
			c.Robot.Exec()
			return kernel.SUCCESS_EMPTY_RESPONSE
		})
		if err != nil {
			logger.Error("request notify fail", "err", err)
		}
	}
	
	if rs != nil {
		defer rs.Body.Close()
		data, err := io.ReadAll(rs.Body)
		if err != nil {
			logger.Error("read response body fail", "err", err)
		}
		_, err = w.Write(data)
		if err != nil {
			logger.Error("write response body fail", "err", err)
		}
		return
	}
	
	w.WriteHeader(http.StatusInternalServerError)
}

func WechatComm(w http.ResponseWriter, r *http.Request) {
	var rs *http.Response
	var err error
	var msgId string
	if r.Method == http.MethodGet {
		rs, err = robot.OfficialAccountApp.Server.VerifyURL(r)
		if err != nil {
			logger.Error("verify url fail", "err", err)
		}
	} else {
		rs, err = robot.OfficialAccountApp.Server.Notify(r, func(event contract.EventInterface) interface{} {
			c, isExec := robot.NewWechatRobot(event)
			if isExec {
				c.Robot.Exec()
			}
			_, msgId, _ = c.Robot.GetChatIdAndMsgIdAndUserID()
			return messages.NewText(c.GetLLMContent())
		})
		if err != nil {
			logger.Error("request notify fail", "err", err)
		}
	}
	
	if rs != nil {
		defer rs.Body.Close()
		data, err := io.ReadAll(rs.Body)
		if err != nil {
			logger.Error("read response body fail", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			logger.Error("write response body fail", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		robot.WechatMsgSent(msgId)
		return
	}
	
	w.WriteHeader(http.StatusInternalServerError)
}

func QQBotComm(w http.ResponseWriter, r *http.Request) {
	webhook.HTTPHandler(w, r, &token.QQBotCredentials{
		AppSecret: *conf.BaseConfInfo.QQAppSecret,
		AppID:     *conf.BaseConfInfo.QQAppID,
	})
}
