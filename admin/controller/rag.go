package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	
	adminUtils "github.com/yincongcyincong/MuseBot/admin/utils"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type RagInfo struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

func ListRagFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/rag/list?page=%s&page_size=%s&name=%s", r.FormValue("page"), r.FormValue("pageSize"), r.FormValue("name")), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
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

func GetRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/rag/get?file_name="+r.FormValue("file_name"), bytes.NewBuffer(nil)))
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

func DeleteRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/rag/delete?file_name="+r.FormValue("file_name"), bytes.NewBuffer(nil)))
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

func CreateRagFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/rag/create", r.Body))
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
