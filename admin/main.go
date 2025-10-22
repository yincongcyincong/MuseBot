package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	
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
	
	mux.HandleFunc("/user/login", controller.UserLogin)
	mux.HandleFunc("/user/me", controller.RequireLogin(controller.GetCurrentUserHandler))
	mux.HandleFunc("/user/logout", controller.RequireLogin(controller.UserLogout))
	
	wrappedMux := WithRequestContext(mux)
	
	err := http.ListenAndServe(fmt.Sprintf(":%s", *conf.BaseConfInfo.AdminPort), wrappedMux)
	if err != nil {
		panic(err)
	}
}
