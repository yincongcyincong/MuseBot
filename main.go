package main

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/robot"
)

func main() {
	conf.InitConf()

	robot.StartListenRobot()

}
