package robot

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

const (
	TelegramRobotType = "telegram"
	DiscordRobotType  = "discord"
)

type Robot struct {
	botType string
	
	telegramRobot *TelegramRobot
	discordRobot  *DiscordRobot
	
	MsgChan chan *param.MsgInfo
}

type botOption func(r *Robot)

func NewRobot(options ...botOption) *Robot {
	r := new(Robot)
	for _, o := range options {
		o(r)
	}
	return r
}

func (r *Robot) Check() {

}

func WithBotType(botType string) func(*Robot) {
	return func(r *Robot) {
		r.botType = botType
	}
}

func StartRobot() {
	if *conf.BaseConfInfo.TelegramBotToken != "" {
		go func() {
			StartTelegramRobot()
		}()
	}
	
	if *conf.BaseConfInfo.DiscordBotToken != "" {
		go func() {
			StartDiscordRobot()
		}()
	}
}
