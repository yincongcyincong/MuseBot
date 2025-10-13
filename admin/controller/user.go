package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	
	"github.com/yincongcyincong/MuseBot/admin/db"
	adminUtils "github.com/yincongcyincong/MuseBot/admin/utils"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.ErrorCtx(ctx, "create user error", "user", u)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.CreateUser(u.Username, u.Password)
	if err != nil {
		logger.ErrorCtx(ctx, "create user error", "user", u)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(ctx, w, r, "success")
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.ErrorCtx(ctx, "get user error", "user", id)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	u, err := db.GetUserByID(id)
	if err != nil {
		logger.ErrorCtx(ctx, "get user error", "user", u)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	utils.Success(ctx, w, r, u)
}

func UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var u User
	err := utils.HandleJsonBody(r, &u)
	if err != nil {
		logger.ErrorCtx(ctx, "update user error", "user", u)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.UpdateUserPassword(u.ID, u.Password)
	if err != nil {
		logger.ErrorCtx(ctx, "update user error", "user", u)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(ctx, w, r, "success")
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.ErrorCtx(ctx, "delete user error", "id", id)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.DeleteUser(id)
	if err != nil {
		logger.ErrorCtx(ctx, "delete user error", "id", id)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	utils.Success(ctx, w, r, "success")
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, pageSize := parsePaginationParams(r)
	
	username := r.URL.Query().Get("username")
	
	offset := (page - 1) * pageSize
	users, total, err := db.ListUsers(offset, pageSize, username)
	if err != nil {
		logger.ErrorCtx(ctx, "list users error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(ctx, w, r, map[string]interface{}{
		"list":  users,
		"total": total,
	})
}

func UpdateUserMode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	userId := r.URL.Query().Get("userId")
	mode := r.URL.Query().Get("mode")
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/user/mode/update?userId=%s&mode=%s", userId, mode))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
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
