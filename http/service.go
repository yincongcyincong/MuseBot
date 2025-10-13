package http

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	
	"github.com/hpcloud/tail"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

func PongHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	utils.Success(ctx, w, r, "pong")
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recordCount, err := db.GetRecordCount("", -1, "")
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	userCount, err := db.GetUserCount("")
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	day := utils.ParseInt(r.URL.Query().Get("day"))
	userDayCount, err := db.GetDailyNewUsers(day)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	recordDayCount, err := db.GetDailyNewRecords(day)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(ctx, w, r, map[string]interface{}{
		"record_count":     recordCount,
		"user_count":       userCount,
		"user_day_count":   userDayCount,
		"record_day_count": recordDayCount,
		"start_time":       conf.BaseConfInfo.StartTime,
	})
	
}

func Restart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := r.URL.Query().Get("params")
	if params == "" {
		logger.ErrorCtx(ctx, "get param error", "param", params)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, "")
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
			cmd := exec.Command(execPath, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = env
			cmd.Dir = filepath.Dir(execPath)
			
			if err := cmd.Start(); err != nil {
				logger.ErrorCtx(ctx, "restart fail", "err", err)
				return
			}
			os.Exit(0)
		} else {
			if err := syscall.Exec(execPath, append([]string{execPath}, args...), env); err != nil {
				logger.ErrorCtx(ctx, "restart fail", "err", err)
				return
			}
		}
	}()
	
	utils.Success(ctx, w, r, "")
}

func Log(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	
	t, _ := tail.TailFile(utils.GetAbsPath("log/muse_bot.log"), tail.Config{
		Follow:    true,
		ReOpen:    true, // 日志切割后自动重新打开
		MustExist: true,
		Poll:      true,
	})
	
	flusher := w.(http.Flusher)
	
	// 用 slice 维护最近 1000 行
	const maxLines = 1000
	var buffer []string
	
	for line := range t.Lines {
		select {
		case <-r.Context().Done():
			return
		default:
			// 存入 buffer
			if len(buffer) >= maxLines {
				// 丢掉最旧的一条
				buffer = buffer[1:]
			}
			buffer = append(buffer, line.Text)
			
			// 只输出 buffer 的最后一条（避免每次都全量输出）
			fmt.Fprintln(w, line.Text)
			flusher.Flush()
		}
	}
}

func Stop(w http.ResponseWriter, r *http.Request) {
	robot.RobotControl.Cancel()
	time.Sleep(100 * time.Millisecond)
	os.Exit(0)
}
