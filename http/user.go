package http

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type UserToken struct {
	UserID int64 `json:"user_id"`
	Token  int   `json:"token"`
}

func AddUserToken(w http.ResponseWriter, r *http.Request) {
	userToken := &UserToken{}
	err := utils.HandleJsonBody(r, userToken)
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	err = db.AddToken(userToken.UserID, userToken.Token)
	if err != nil {
		logger.Error("add user token error", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	utils.Success(w, "success")
}
