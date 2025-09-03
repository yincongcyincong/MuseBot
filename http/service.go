package http

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
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

func Restart(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query().Get("params")
	if params == "" {
		logger.Error("get param error", "param", params)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, "")
		return
	}
	
	lines := strings.Split(params, "\n")
	
	execPath, _ := os.Executable()
	
	args := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			args = append(args, line)
		}
	}
	
	env := os.Environ()
	
	go func() {
		if runtime.GOOS == "windows" {
			cmd := exec.Command(execPath, append([]string{execPath}, args...)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = env
			cmd.Dir = filepath.Dir(execPath)
			
			if err := cmd.Start(); err != nil {
				logger.Error("restart fail", "err", err)
				return
			}
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		} else {
			if err := syscall.Exec(execPath, append([]string{execPath}, args...), env); err != nil {
				logger.Error("restart fail", "err", err)
				return
			}
		}
	}()
	
	utils.Success(w, "")
}
