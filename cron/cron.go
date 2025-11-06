package cron

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

type Cron struct {
	Cron    string `json:"cron"`
	Command string `json:"command"`
	Type    string `json:"type"`
	UserIds string `json:"user_ids"`
	Prompt  string `json:"prompt"`
}

var (
	Crons = make([]*Cron, 0)
)

func InitCron() {
	file, err := os.Open("./conf/cron/cron.json")
	if err != nil {
		logger.Error("open cron.json error", err)
		return
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		logger.Error("read command.json error", err)
		return
	}
	
	err = json.Unmarshal(data, &Crons)
	if err != nil {
		logger.Error("parse command.json error", err)
	}
	
	cr := cron.New(cron.WithSeconds())
	
	for _, c := range Crons {
		if c.Cron != "" && c.Type != "" && c.UserIds != "" {
			_, err = cr.AddFunc(c.Cron, func() {
				c.Exec()
			})
			if err != nil {
				logger.Error("crontab parse error", "err", err)
			}
		}
	}
	
	cr.Start()
}

func (c *Cron) Exec() {
	userIds := strings.Split(c.UserIds, ",")
	for _, userId := range userIds {
		switch c.Type {
		case param.Wechat:
			c.ExecWechat(userId)
		case param.Telegram:
			c.ExecTelegram(userId)
		}
	}
	
}

func (c *Cron) ExecTelegram(userId string) {
	t := &robot.TelegramRobot{
		Bot: robot.TelegramBot,
		Update: tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{
					ID: int64(utils.ParseInt(userId)),
				},
				Chat: &tgbotapi.Chat{
					ID: int64(utils.ParseInt(userId)),
				},
				Text: c.Command + " " + c.Prompt,
			},
		},
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

func (c *Cron) ExecWechat(userId string) {
	if robot.OfficialAccountApp == nil {
		logger.Warn("official account app is nil", "type", c.Type, "userId", userId, "prompt", c.Prompt)
		return
	}
	w := &robot.WechatRobot{
		Event: models.CallbackMessageHeader{
			ToUserName: userId,
		},
		App: robot.OfficialAccountApp,
		TextMsg: &serverModel.MessageText{
			Content: c.Command + " " + c.Prompt,
		},
	}
	w.Robot = robot.NewRobot(robot.WithRobot(w), robot.WithTencentRobot(w),
		robot.WithSkipCheck(true))
	w.Robot.Exec()
}
