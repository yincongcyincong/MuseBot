package http

import (
	"net/http"
	
	cronC "github.com/robfig/cron/v3"
	"github.com/yincongcyincong/MuseBot/cron"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// ... (省略原有导入和代码)

// CronRequest 定义了创建和更新定时任务的请求体结构
type CronRequest struct {
	ID       int64  `json:"id"`
	CronName string `json:"cron_name"`
	CronSpec string `json:"cron_spec"`
	TargetID string `json:"target_id"`
	GroupID  string `json:"group_id"`
	Command  string `json:"command"`
	Prompt   string `json:"prompt"`
	Type     string `json:"type"`
}

func CreateCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &CronRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	// 简单参数校验
	if req.CronName == "" || req.CronSpec == "" || req.TargetID == "" {
		utils.Failure(ctx, w, r, param.CodeParamError, "CronName, CronSpec, and TargetID are required", nil)
		return
	}
	
	id, err := db.InsertCron(req.CronName, req.CronSpec, req.TargetID, req.GroupID, req.Command, req.Prompt, req.Type)
	if err != nil {
		logger.ErrorCtx(ctx, "insert cron error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	cronInfo, err := db.GetCronByID(id)
	if err != nil {
		logger.ErrorCtx(ctx, "get cron by id error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	err = AddCron(cronInfo)
	if err != nil {
		logger.ErrorCtx(ctx, "add cron error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(ctx, w, r, map[string]int64{"id": id})
}

func UpdateCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &CronRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	if req.ID == 0 {
		utils.Failure(ctx, w, r, param.CodeParamError, "ID is required for update", nil)
		return
	}
	
	// 1. 更新数据库
	err = db.UpdateCron(req.ID, req.CronName, req.CronSpec, req.TargetID, req.GroupID, req.Command, req.Prompt, req.Type)
	if err != nil {
		logger.ErrorCtx(ctx, "update cron error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	cronInfo, err := db.GetCronByID(req.ID)
	if err != nil {
		logger.ErrorCtx(ctx, "get cron by id error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	cron.Cron.Remove(cronC.EntryID(cronInfo.CronJobId))
	err = AddCron(cronInfo)
	if err != nil {
		logger.ErrorCtx(ctx, "add cron error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(ctx, w, r, nil)
}

func UpdateCronStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 假设请求体包含 ID 和 Status
	type StatusRequest struct {
		ID     int64 `json:"id"`
		Status int   `json:"status"` // 0:disable 1:enable
	}
	req := &StatusRequest{}
	err := utils.HandleJsonBody(r, req)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	if req.ID == 0 || (req.Status != 0 && req.Status != 1) {
		utils.Failure(ctx, w, r, param.CodeParamError, "Invalid ID or Status", nil)
		return
	}
	
	// 1. 查询当前任务获取 Job ID
	cronTask, err := db.GetCronByID(req.ID)
	if err != nil {
		logger.ErrorCtx(ctx, "get cron task error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	if cronTask.Status == req.Status {
		utils.Success(ctx, w, r, nil)
		return
	}
	
	// 2. 更新数据库状态
	err = db.UpdateCronStatus(req.ID, req.Status)
	if err != nil {
		logger.ErrorCtx(ctx, "update cron status error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	if req.Status == 0 {
		cron.Cron.Remove(cronC.EntryID(cronTask.CronJobId))
	} else {
		err = AddCron(cronTask)
		if err != nil {
			logger.ErrorCtx(ctx, "add cron error", "err", err)
			utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	}
	
	utils.Success(ctx, w, r, nil)
}

func DeleteCron(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	id := utils.ParseInt(r.FormValue("id"))
	if id == 0 {
		utils.Failure(ctx, w, r, param.CodeParamError, "ID is required for delete", nil)
		return
	}
	
	// 1. 查询当前任务获取 Job ID
	cronTask, err := db.GetCronByID(int64(id))
	if err != nil {
		logger.ErrorCtx(ctx, "get cron task error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 2. 软删除数据库记录
	err = db.DeleteCronByID(int64(id))
	if err != nil {
		logger.ErrorCtx(ctx, "delete cron error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	cron.Cron.Remove(cronC.EntryID(cronTask.CronJobId))
	
	utils.Success(ctx, w, r, nil)
}

func GetCronByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	id := utils.ParseInt(r.FormValue("id"))
	if id == 0 {
		utils.Failure(ctx, w, r, param.CodeParamError, "ID is required", nil)
		return
	}
	
	cronTask, err := db.GetCronByID(int64(id))
	if err != nil {
		logger.ErrorCtx(ctx, "get cron task error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(ctx, w, r, cronTask)
}

func GetCrons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("page_size"))
	name := r.FormValue("name")
	
	// 1. 查询列表数据
	cronTasks, err := db.GetCronsByPage(page, pageSize, name)
	if err != nil {
		logger.ErrorCtx(ctx, "get cron tasks error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	// 2. 查询总数
	total, err := db.GetCronsCount(name)
	if err != nil {
		logger.ErrorCtx(ctx, "get cron tasks count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	result := map[string]interface{}{
		"list":  cronTasks,
		"total": total,
	}
	utils.Success(ctx, w, r, result)
}

func AddCron(cronInfo *db.Cron) error {
	cronJobId, err := cron.Cron.AddFunc(cronInfo.CronSpec, func() {
		cron.Exec(cronInfo)
	})
	if err != nil {
		return err
	}
	
	err = db.UpdateCronJobId(cronInfo.ID, int(cronJobId))
	if err != nil {
		return err
	}
	
	return nil
	
}
