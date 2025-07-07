package controller

import (
	"net/http"
	"strconv"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.Error("create user error", "user", u)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.CreateUser(u.Username, u.Password)
	if err != nil {
		logger.Error("create user error", "user", u)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(w, "success")
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Error("get user error", "user", id)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	u, err := db.GetUserByID(id)
	if err != nil {
		logger.Error("get user error", "user", u)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	utils.Success(w, u)
}

func UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.Error("update user error", "user", u)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.UpdateUserPassword(u.ID, u.Password)
	if err != nil {
		logger.Error("update user error", "user", u)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(w, "success")
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Error("delete user error", "id", id)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.DeleteUser(id)
	if err != nil {
		logger.Error("delete user error", "id", id)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(w, "success")
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePaginationParams(r)
	
	username := r.URL.Query().Get("username")
	
	offset := (page - 1) * pageSize
	users, total, err := db.ListUsers(offset, pageSize, username)
	if err != nil {
		logger.Error("list users error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(w, map[string]interface{}{
		"list":  users,
		"total": total,
	})
}

func parsePaginationParams(r *http.Request) (page int, pageSize int) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	
	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return
}
