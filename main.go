package main

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/command"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/robot"
)

func main() {
	logger.InitLogger()
	conf.InitConf()
	i18n.InitI18n()
	command.LoadCustomCommands()
	db.InitTable()
	db.UpdateUserTime()
	metrics.InitPprof()
	metrics.RegisterMetrics()
	robot.StartListenRobot()
}
