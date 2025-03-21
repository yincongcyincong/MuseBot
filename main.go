package main

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/robot"
)

func main() {
	conf.InitConf()
	logger.InitLogger()
	db.InitTable()
	db.UpdateUserTime()
	metrics.InitPprof()
	metrics.RegisterMetrics()
	robot.StartListenRobot()
}
