package controller

import (
	"net/http"
	
	"github.com/gorilla/sessions"
	"github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	sessionStore *sessions.CookieStore
	sessionName  = "MuseBot-session"
)

func InitSessionStore() {
	sessionStore = sessions.NewCookieStore([]byte(*conf.BaseConfInfo.SessionKey))
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.ErrorCtx(ctx, "create user error", "user", u)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	if u.Username == "" || u.Password == "" {
		logger.ErrorCtx(ctx, "login error", "reason", "empty username or password")
		utils.Failure(w, param.CodeParamError, param.MsgParamError, nil)
		return
	}
	
	user, err := db.GetUserByUsername(u.Username)
	if err != nil {
		logger.ErrorCtx(ctx, "login error", "reason", "user not found", "username", u.Username, "err", err)
		utils.Failure(w, param.CodeLoginFail, param.MsgLoginFail, nil)
		return
	}
	
	if user.Password != utils.MD5(u.Password) {
		logger.ErrorCtx(ctx, "login error", "reason", "wrong password", "username", u.Username)
		utils.Failure(w, param.CodeLoginFail, param.MsgLoginFail, nil)
		return
	}
	
	session, _ := sessionStore.Get(r, sessionName)
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	err = session.Save(r, w)
	if err != nil {
		logger.ErrorCtx(ctx, "save session error", "err", err)
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
		ctx := r.Context()
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
			logger.ErrorCtx(ctx, "get session error", "err", err)
			return
		}
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
			logger.ErrorCtx(ctx, "get session error", "err", err)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		logger.ErrorCtx(ctx, "get session error", "err", err)
		return
	}
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		logger.ErrorCtx(ctx, "get session error", "userId", userIDValue)
		return
	}
	
	userName, ok := session.Values["username"]
	if !ok || userName == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		logger.ErrorCtx(ctx, "get session error", "userName", userName)
		return
	}
	
	res := map[string]interface{}{
		"user_id":  userIDValue,
		"username": userName,
	}
	
	utils.Success(w, res)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		logger.ErrorCtx(ctx, "get session error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		logger.ErrorCtx(ctx, "save session error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(w, "success")
}
