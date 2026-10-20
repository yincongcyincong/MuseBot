package http

import (
	"encoding/base64"
	"net/http"

	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type UserToken struct {
	UserID string `json:"user_id"`
	Token  int    `json:"token"`
}

func AddUserToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userToken := &UserToken{}
	err := utils.HandleJsonBody(r, userToken)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	err = db.AddAvailToken(userToken.UserID, userToken.Token)
	if err != nil {
		logger.ErrorCtx(ctx, "add user token error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	utils.Success(ctx, w, r, "success")
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 解析参数
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("page_size"))
	userId := r.FormValue("user_id")

	users, err := db.GetUserByPage(page, pageSize, userId)
	if err != nil {
		logger.ErrorCtx(ctx, "get user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	total, err := db.GetUserCount(userId)
	if err != nil {
		logger.ErrorCtx(ctx, "get user count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBWriteFail, err)
		return
	}

	// 返回结果
	result := map[string]interface{}{
		"list":  users,
		"total": total,
	}

	utils.Success(ctx, w, r, result)
}

func GetRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 获取参数
	query := r.URL.Query()
	page := utils.ParseInt(query.Get("page"))
	pageSize := utils.ParseInt(query.Get("page_size"))
	isDeleted := -1
	if query.Get("is_deleted") != "" {
		isDeleted = utils.ParseInt(query.Get("is_deleted"))
	}
	userId := query.Get("user_id")
	recordTypeStr := query.Get("record_type")

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数和数据
	total, err := db.GetRecordCount(userId, isDeleted, recordTypeStr)
	if err != nil {
		logger.ErrorCtx(ctx, "get record count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	list, err := db.GetRecordList(userId, page, pageSize, isDeleted, recordTypeStr)
	if err != nil {
		logger.ErrorCtx(ctx, "get record list error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	// media is stored as local file paths since the media-storage refactor;
	// inline it back as data URIs so the admin ui keeps rendering history
	for i := range list {
		list[i].Answer = mediaPathToDataURI(list[i].Answer)
		list[i].Content = mediaPathToDataURI(list[i].Content)
	}

	result := map[string]interface{}{
		"list":  list,
		"total": total,
	}

	utils.Success(ctx, w, r, result)
}

// mediaPathToDataURI inlines a stored media file path as a data URI for the
// admin ui. Non-media values (legacy base64 rows, plain text answers) are
// returned unchanged; unreadable files fall back to the raw path so a
// cleaned-up file never breaks the whole record list.
func mediaPathToDataURI(p string) string {
	if !utils.IsMediaPath(p) {
		return p
	}

	data, err := utils.ReadMediaFile(p)
	if err != nil {
		return p
	}

	mimeType := getMimeType(p)
	if mimeType == "application/octet-stream" {
		// keep the legacy guess: unknown media in records is mostly images
		mimeType = "image/png"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func InsertUserRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userRecords := &db.UserRecords{}
	err := utils.HandleJsonBody(r, userRecords)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	err = db.InsertUserRecords(userRecords)
	if err != nil {
		logger.ErrorCtx(ctx, "change user mode error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}

	for _, aq := range userRecords.Records {
		db.InsertMsgRecord(ctx, userRecords.UserId, aq, false)
	}

	utils.Success(ctx, w, r, "success")

}
