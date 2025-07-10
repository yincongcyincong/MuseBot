package controller

import (
	"net/http"
	
	"github.com/gorilla/sessions"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

var (
	sessionStore *sessions.CookieStore
	sessionName  = "telegram-deepseek-bot-session"
)

func InitSessionStore() {
	sessionStore = sessions.NewCookieStore([]byte(*conf.BaseConfInfo.SessionKey))
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.Error("create user error", "user", u)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	if u.Username == "" || u.Password == "" {
		logger.Error("login error", "reason", "empty username or password")
		utils.Failure(w, param.CodeParamError, param.MsgParamError, nil)
		return
	}
	
	user, err := db.GetUserByUsername(u.Username)
	if err != nil {
		logger.Error("login error", "reason", "user not found", "username", u.Username, "err", err)
		utils.Failure(w, param.CodeLoginFail, param.MsgLoginFail, nil)
		return
	}
	
	if user.Password != utils.MD5(u.Password) {
		logger.Error("login error", "reason", "wrong password", "username", u.Username)
		utils.Failure(w, param.CodeLoginFail, param.MsgLoginFail, nil)
		return
	}
	
	session, _ := sessionStore.Get(r, sessionName)
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	err = session.Save(r, w)
	if err != nil {
		logger.Error("save session error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	// 登录成功，返回用户信息（不含密码）
	resp := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
	}
	utils.Success(w, resp)
}

func RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
			return
		}
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	userName, ok := session.Values["username"]
	if !ok || userName == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	res := map[string]interface{}{
		"user_id":  userIDValue,
		"username": userName,
	}
	
	utils.Success(w, res)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		logger.Error("get session error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		logger.Error("save session error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(w, "success")
}
