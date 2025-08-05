package http

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

func PongHandler(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, "pong")
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	recordCount, err := db.GetRecordCount("", -1, "")
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	userCount, err := db.GetUserCount("")
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	day := utils.ParseInt(r.URL.Query().Get("day"))
	userDayCount, err := db.GetDailyNewUsers(day)
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	recordDayCount, err := db.GetDailyNewRecords(day)
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(w, map[string]interface{}{
		"record_count":     recordCount,
		"user_count":       userCount,
		"user_day_count":   userDayCount,
		"record_day_count": recordDayCount,
		"start_time":       conf.BaseConfInfo.StartTime,
	})
	
}
