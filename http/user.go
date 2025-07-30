package http

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type UserToken struct {
	UserID string `json:"user_id"`
	Token  int    `json:"token"`
}

func AddUserToken(w http.ResponseWriter, r *http.Request) {
	userToken := &UserToken{}
	err := utils.HandleJsonBody(r, userToken)
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	err = db.AddAvailToken(userToken.UserID, userToken.Token)
	if err != nil {
		logger.Error("add user token error", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	utils.Success(w, "success")
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	err := r.ParseForm()
	if err != nil {
		logger.Error("parse form error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("pageSize"))
	userId := r.FormValue("userId")
	
	users, err := db.GetUserByPage(page, pageSize, userId)
	if err != nil {
		logger.Error("get user error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	total, err := db.GetUserCount(userId)
	if err != nil {
		logger.Error("get user count error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBWriteFail, err)
		return
	}
	
	// 返回结果
	result := map[string]interface{}{
		"list":  users,
		"total": total,
	}
	
	utils.Success(w, result)
}

func GetRecords(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	query := r.URL.Query()
	page := utils.ParseInt(query.Get("page"))
	pageSize := utils.ParseInt(query.Get("pageSize"))
	isDeleted := -1
	if query.Get("isDeleted") != "" {
		isDeleted = utils.ParseInt(query.Get("isDeleted"))
	}
	userId := query.Get("userId")
	
	if page <= 0 {
		page = 1
	}
	
	if pageSize <= 0 {
		pageSize = 10
	}
	
	// 查询总数和数据
	total, err := db.GetRecordCount(userId, isDeleted, param.WEBRecordType)
	if err != nil {
		logger.Error("get record count error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	list, err := db.GetRecordList(userId, page, pageSize, isDeleted, param.WEBRecordType)
	if err != nil {
		logger.Error("get record list error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	result := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	
	utils.Success(w, result)
}
