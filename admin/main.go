package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	_ "net/http/pprof"
	"time"
	
	"github.com/google/uuid"
	"github.com/yincongcyincong/MuseBot/admin/checkpoint"
	"github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/admin/controller"
	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/logger"
)

func main() {
	logger.InitLogger()
	conf.InitConfig()
	controller.InitSessionStore()
	db.InitTable()
	checkpoint.InitStatusCheck()
	
	mux := http.NewServeMux()
	mux.Handle("/", View())
	
	// User API
	mux.HandleFunc("/user/create", controller.RequireLogin(controller.CreateUser))
	mux.HandleFunc("/user/get", controller.RequireLogin(controller.GetUser))
	mux.HandleFunc("/user/update", controller.RequireLogin(controller.UpdateUserPassword))
	mux.HandleFunc("/user/delete", controller.RequireLogin(controller.DeleteUser))
	mux.HandleFunc("/user/list", controller.RequireLogin(controller.ListUsers))
	
	// Bot API
	mux.HandleFunc("/bot/dashboard", controller.RequireLogin(controller.Dashboard))
	mux.HandleFunc("/bot/create", controller.RequireLogin(controller.CreateBot))
	mux.HandleFunc("/bot/get", controller.RequireLogin(controller.GetBot))
	mux.HandleFunc("/bot/restart", controller.RequireLogin(controller.RestartBot))
	mux.HandleFunc("/bot/stop", controller.RequireLogin(controller.StopBot))
	mux.HandleFunc("/bot/log", controller.RequireLogin(controller.Log))
	mux.HandleFunc("/bot/update", controller.RequireLogin(controller.UpdateBotAddress))
	mux.HandleFunc("/bot/delete", controller.RequireLogin(controller.SoftDeleteBot))
	mux.HandleFunc("/bot/list", controller.RequireLogin(controller.ListBots))
	mux.HandleFunc("/bot/conf/get", controller.RequireLogin(controller.GetBotConf))
	mux.HandleFunc("/bot/conf/update", controller.RequireLogin(controller.UpdateBotConf))
	mux.HandleFunc("/bot/command/get", controller.RequireLogin(controller.GetBotCommand))
	mux.HandleFunc("/bot/record/list", controller.RequireLogin(controller.GetBotUserRecord))
	mux.HandleFunc("/bot/user/list", controller.RequireLogin(controller.GetBotUser))
	mux.HandleFunc("/bot/user/mode/update", controller.RequireLogin(controller.UpdateUserMode))
	mux.HandleFunc("/bot/user/insert/records", controller.RequireLogin(controller.InsertUserRecord))
	mux.HandleFunc("/bot/add/token", controller.RequireLogin(controller.AddUserToken))
	mux.HandleFunc("/bot/online", controller.RequireLogin(controller.GetAllOnlineBot))
	mux.HandleFunc("/bot/mcp/get", controller.RequireLogin(controller.GetBotMCPConf))
	mux.HandleFunc("/bot/mcp/update", controller.RequireLogin(controller.UpdateBotMCPConf))
	mux.HandleFunc("/bot/mcp/delete", controller.RequireLogin(controller.DeleteBotMCPConf))
	mux.HandleFunc("/bot/mcp/disable", controller.RequireLogin(controller.DisableBotMCPConf))
	mux.HandleFunc("/bot/mcp/prepare", controller.RequireLogin(controller.GetPrepareMCPServer))
	mux.HandleFunc("/bot/mcp/sync", controller.RequireLogin(controller.SyncMCPServer))
	mux.HandleFunc("/bot/communicate", controller.RequireLogin(controller.Communicate))
	mux.HandleFunc("/bot/admin/chat", controller.RequireLogin(controller.GetBotAdminRecord))
	mux.HandleFunc("/bot/rag/list", controller.RequireLogin(controller.ListRagFiles))
	mux.HandleFunc("/bot/rag/delete", controller.RequireLogin(controller.DeleteRagFile))
	mux.HandleFunc("/bot/rag/create", controller.RequireLogin(controller.CreateRagFile))
	mux.HandleFunc("/bot/rag/get", controller.RequireLogin(controller.GetRagFile))
	mux.HandleFunc("/bot/cron/list", controller.RequireLogin(controller.ListCrons))
	mux.HandleFunc("/bot/cron/delete", controller.RequireLogin(controller.DeleteCron))
	mux.HandleFunc("/bot/cron/create", controller.RequireLogin(controller.CreateCron))
	mux.HandleFunc("/bot/cron/update/status", controller.RequireLogin(controller.UpdateCronStatus))
	mux.HandleFunc("/bot/cron/update", controller.RequireLogin(controller.UpdateCron))
	
	mux.HandleFunc("/user/login", controller.UserLogin)
	mux.HandleFunc("/user/me", controller.RequireLogin(controller.GetCurrentUserHandler))
	mux.HandleFunc("/user/logout", controller.RequireLogin(controller.UserLogout))
	
	wrappedMux := WithRequestContext(mux)
	
	err := http.ListenAndServe(fmt.Sprintf(":%s", conf.BaseConfInfo.AdminPort), wrappedMux)
	if err != nil {
		panic(err)
	}
}

//go:embed adminui/*
var staticFiles embed.FS

func View() http.HandlerFunc {
	distFS, _ := fs.Sub(staticFiles, "adminui")
	
	staticHandler := http.FileServer(http.FS(distFS))
	
	return func(w http.ResponseWriter, r *http.Request) {
		if fileExists(distFS, r.URL.Path[1:]) {
			staticHandler.ServeHTTP(w, r)
			return
		}
		
		fileBytes, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		
		reader := bytes.NewReader(fileBytes)
		http.ServeContent(w, r, "index.html", time.Now(), reader)
	}
}

func fileExists(fsys fs.FS, path string) bool {
	f, err := fsys.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

func WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logID := uuid.New().String()
		
		isSSE := r.Header.Get("Accept") == "text/event-stream"
		
		if !isSSE {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
		}
		
		// 通用的 context 值
		ctx = context.WithValue(ctx, "log_id", logID)
		ctx = context.WithValue(ctx, "start_time", time.Now())
		r = r.WithContext(ctx)
		
		logger.InfoCtx(ctx, "request start", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
