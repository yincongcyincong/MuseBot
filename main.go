//go:build !libtokenizers

package main

import (
	"os"
	"os/signal"
	"syscall"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/http"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/rag"
	"github.com/yincongcyincong/telegram-deepseek-bot/robot"
)

func main() {
	logger.InitLogger()
	conf.InitConf()
	i18n.InitI18n()
	db.InitTable()
	db.UpdateUserTime()
	conf.InitTools()
	rag.InitRag()
	http.InitHTTP()
	metrics.RegisterMetrics()
	robot.StartRobot()
	
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
