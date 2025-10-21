package http

import (
	"fmt"
	"io"
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
	typ := r.URL.Query().Get("type")
	
	maxLines := 5000
	if typ != "" {
		maxLines = 100000
	}
	
	filePath := utils.GetAbsPath("log/muse_bot.log")
	startFrom, err := utils.GetTailStartOffset(filePath, maxLines)
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}
	
	t, err := tail.TailFile(filePath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Poll:      true,
		Location:  &tail.SeekInfo{Offset: startFrom, Whence: io.SeekStart},
	})
	if err != nil {
		http.Error(w, "Failed to tail log file", http.StatusInternalServerError)
		return
	}
	
	flusher := w.(http.Flusher)
	
	for line := range t.Lines {
		select {
		case <-r.Context().Done():
			return
		default:
			if typ != "" && !strings.Contains(line.Text, typ) {
				continue
			}
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
