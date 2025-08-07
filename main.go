//go:build !libtokenizers

package main

import (
	"os"
	"os/signal"
	"syscall"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/http"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/rag"
	"github.com/yincongcyincong/MuseBot/robot"
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
