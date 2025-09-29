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
	
	http.Handle("/", View())
	
	// User API
	http.HandleFunc("/user/create", controller.RequireLogin(controller.CreateUser))
	http.HandleFunc("/user/get", controller.RequireLogin(controller.GetUser))
	http.HandleFunc("/user/update", controller.RequireLogin(controller.UpdateUserPassword))
	http.HandleFunc("/user/delete", controller.RequireLogin(controller.DeleteUser))
	http.HandleFunc("/user/list", controller.RequireLogin(controller.ListUsers))
	
	// Bot API
	http.HandleFunc("/bot/dashboard", controller.RequireLogin(controller.Dashboard))
	http.HandleFunc("/bot/create", controller.RequireLogin(controller.CreateBot))
	http.HandleFunc("/bot/get", controller.RequireLogin(controller.GetBot))
	http.HandleFunc("/bot/restart", controller.RequireLogin(controller.RestartBot))
	http.HandleFunc("/bot/stop", controller.RequireLogin(controller.StopBot))
	http.HandleFunc("/bot/log", controller.RequireLogin(controller.Log))
	http.HandleFunc("/bot/update", controller.RequireLogin(controller.UpdateBotAddress))
	http.HandleFunc("/bot/delete", controller.RequireLogin(controller.SoftDeleteBot))
	http.HandleFunc("/bot/list", controller.RequireLogin(controller.ListBots))
	http.HandleFunc("/bot/conf/get", controller.RequireLogin(controller.GetBotConf))
	http.HandleFunc("/bot/conf/update", controller.RequireLogin(controller.UpdateBotConf))
	http.HandleFunc("/bot/command/get", controller.RequireLogin(controller.GetBotCommand))
	http.HandleFunc("/bot/record/list", controller.RequireLogin(controller.GetBotUserRecord))
	http.HandleFunc("/bot/user/list", controller.RequireLogin(controller.GetBotUser))
	http.HandleFunc("/bot/user/mode/update", controller.RequireLogin(controller.UpdateUserMode))
	http.HandleFunc("/bot/user/insert/records", controller.RequireLogin(controller.InsertUserRecord))
	http.HandleFunc("/bot/add/token", controller.RequireLogin(controller.AddUserToken))
	http.HandleFunc("/bot/online", controller.RequireLogin(controller.GetAllOnlineBot))
	http.HandleFunc("/bot/mcp/get", controller.RequireLogin(controller.GetBotMCPConf))
	http.HandleFunc("/bot/mcp/update", controller.RequireLogin(controller.UpdateBotMCPConf))
	http.HandleFunc("/bot/mcp/delete", controller.RequireLogin(controller.DeleteBotMCPConf))
	http.HandleFunc("/bot/mcp/disable", controller.RequireLogin(controller.DisableBotMCPConf))
	http.HandleFunc("/bot/mcp/prepare", controller.RequireLogin(controller.GetPrepareMCPServer))
	http.HandleFunc("/bot/mcp/sync", controller.RequireLogin(controller.SyncMCPServer))
	http.HandleFunc("/bot/communicate", controller.RequireLogin(controller.Communicate))
	http.HandleFunc("/bot/admin/chat", controller.RequireLogin(controller.GetBotAdminRecord))
	
	http.HandleFunc("/user/login", controller.UserLogin)
	http.HandleFunc("/user/me", controller.RequireLogin(controller.GetCurrentUserHandler))
	http.HandleFunc("/user/logout", controller.RequireLogin(controller.UserLogout))
	
	err := http.ListenAndServe(fmt.Sprintf(":%s", *conf.BaseConfInfo.AdminPort), nil)
	if err != nil {
		panic(err)
	}
}
