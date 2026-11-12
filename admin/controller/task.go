package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	adminUtils "github.com/yincongcyincong/MuseBot/admin/utils"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// ListTaskRuns 转发分页查询任务执行列表的请求（运行中置顶 + 历史快照）
func ListTaskRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	// 解析请求参数 (page, pageSize)
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}

	// 构造目标 URL：/task/list?page=...&page_size=...
	// 前端发送的是 page_size（与主服务 /task/list 一致），不是 pageSize
	targetURL := strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/task/list?page=%s&page_size=%s",
			url.QueryEscape(r.FormValue("page")), url.QueryEscape(r.FormValue("page_size")))

	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, targetURL, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "request task list error", "err", err)
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

// GetTaskRun 转发查询单次任务执行完整快照的请求（流程图、步骤、事件、最终结果）
func GetTaskRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	runId := r.FormValue("run_id")
	if runId == "" {
		utils.Failure(ctx, w, r, param.CodeParamError, "run_id is required", nil)
		return
	}

	// 构造目标 URL：/task/get?run_id=...
	targetURL := strings.TrimSuffix(botInfo.Address, "/") + "/task/get?run_id=" + url.QueryEscape(runId)

	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, targetURL, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "request task run error", "err", err)
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
