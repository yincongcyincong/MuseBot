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

// ListCrons 转发分页查询 Cron 任务列表的请求
func ListCrons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 解析请求参数 (page, pageSize, name)
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	// 构造目标 URL：/cron/list?page=...&page_size=...&name=...
	targetURL := strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/cron/list?page=%s&page_size=%s&name=%s", r.FormValue("page"), r.FormValue("pageSize"), r.FormValue("name"))
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, targetURL, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "request cron list error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	// 直接将 Bot 响应体复制给客户端
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

// CreateCron 转发创建 Cron 任务的请求。请求体是 JSON。
func CreateCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 构造目标 URL：/cron/create
	targetURL := strings.TrimSuffix(botInfo.Address, "/") + "/cron/create"
	
	// 使用 r.Body 作为请求体（原样转发客户端的 JSON POST 数据）
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost, targetURL, r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "request create cron error", "err", err)
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

// UpdateCron 转发更新 Cron 任务主要信息的请求。请求体是 JSON。
func UpdateCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 构造目标 URL：/cron/update
	targetURL := strings.TrimSuffix(botInfo.Address, "/") + "/cron/update"
	
	// 使用 r.Body 作为请求体（原样转发客户端的 JSON POST 数据）
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost, targetURL, r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "request update cron error", "err", err)
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

// UpdateCronStatus 转发启用/禁用 Cron 任务状态的请求。请求体是包含 ID 和 Status 的 JSON。
func UpdateCronStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 构造目标 URL：/cron/update_status
	targetURL := strings.TrimSuffix(botInfo.Address, "/") + "/cron/update_status"
	
	// 使用 r.Body 作为请求体（原样转发客户端的 JSON POST 数据）
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost, targetURL, r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "request update cron status error", "err", err)
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

// DeleteCron 转发软删除 Cron 任务的请求。参数通过 URL Query 传递 ID。
func DeleteCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 解析 ID 参数
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	id := r.FormValue("cron_id")
	
	// 构造目标 URL：/cron/delete?id=...
	targetURL := strings.TrimSuffix(botInfo.Address, "/") + "/cron/delete?id=" + id
	
	// 转发 GET 请求（与您示例中的 DeleteRagFile 保持一致）
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, targetURL, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "request delete cron error", "err", err)
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
