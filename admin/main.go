package main

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/controller"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

func main() {
	logger.InitLogger()
	conf.InitConfig()
	controller.InitSessionStore()
	db.InitTable()
	
	http.Handle("/", View())
	
	// User API
	http.HandleFunc("/user/create", controller.CreateUser)
	http.HandleFunc("/user/get", controller.RequireLogin(controller.GetUser))
	http.HandleFunc("/user/update", controller.UpdateUserPassword)
	http.HandleFunc("/user/delete", controller.DeleteUser)
	http.HandleFunc("/user/list", controller.ListUsers)
	
	// Bot API
	http.HandleFunc("/bot/create", controller.CreateBot)
	http.HandleFunc("/bot/get", controller.GetBot)
	http.HandleFunc("/bot/update", controller.UpdateBotAddress)
	http.HandleFunc("/bot/delete", controller.SoftDeleteBot)
	http.HandleFunc("/bot/list", controller.ListBots)
	
	http.HandleFunc("/user/login", controller.UserLogin)
	http.HandleFunc("/user/me", controller.GetCurrentUserHandler)
	http.HandleFunc("/user/logout", controller.RequireLogin(controller.UserLogout))
	
	err := http.ListenAndServe(":18080", nil)
	if err != nil {
		panic(err)
	}
}
