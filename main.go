package main

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/robot"
)

func main() {
	conf.InitConf()
	if *conf.Mode == conf.ComplexMode {
		db.InitTable()
	}
	robot.StartListenRobot()
}
