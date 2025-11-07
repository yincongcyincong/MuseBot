package cron

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack/slackevents"
	"github.com/tencent-connect/botgo/dto"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

type Cron struct {
	Cron      string `json:"cron"`
	Command   string `json:"command"`
	Type      string `json:"type"`
	TargetIds string `json:"target_ids"`
	Prompt    string `json:"prompt"`
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
	cr.Stop()
	
	for _, c := range Crons {
		if c.Cron != "" && c.Type != "" && c.TargetIds != "" {
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
	targetIds := strings.Split(c.TargetIds, ",")
	for _, targetId := range targetIds {
		switch c.Type {
		case param.Wechat:
			c.ExecWechat(targetId)
		case param.Telegram:
			c.ExecTelegram(targetId)
		case param.ComWechat:
			c.ExecComWechat(targetId)
		case param.Ding:
			c.ExecDing(targetId)
		case param.Lark:
			c.ExecLark(targetId)
		case param.PersonalQQ:
			c.ExecPersonalQQ(targetId)
		case param.Slack:
			c.ExecSlack(targetId)
		}
	}
}

func (c *Cron) ExecDing(userId string) {
	if robot.DingBotClient == nil {
		logger.Error("dingbot client is nil")
		return
	}
	
	t := &robot.DingRobot{
		Message: &chatbot.BotCallbackDataModel{
			SenderId: userId,
			Text: chatbot.BotCallbackDataTextModel{
				Content: c.Prompt,
			},
			Msgtype: "text",
		},
		Client: robot.DingBotClient,
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

func (c *Cron) ExecLark(userId string) {
	if robot.LarkBotClient == nil {
		logger.Error("larkbot client is nil")
		return
	}
	contentStr := robot.MessageText{
		Text: c.Command + " " + c.Prompt,
	}
	contentByte, _ := json.Marshal(contentStr)
	content := string(contentByte)
	msgType := larkim.MsgTypeText
	t := &robot.LarkRobot{
		Message: &larkim.P2MessageReceiveV1{
			Event: &larkim.P2MessageReceiveV1Data{
				Message: &larkim.EventMessage{
					Content:     &content,
					MessageType: &msgType,
					ChatId:      &userId,
				},
				Sender: &larkim.EventSender{
					SenderId: &larkim.UserId{UserId: &userId},
				},
			},
		},
		Client: robot.LarkBotClient,
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

func (c *Cron) ExecPersonalQQ(userId string) {
	t := &robot.PersonalQQRobot{
		Msg: &robot.QQMessage{
			MessageType: "private",
			UserID:      int64(utils.ParseInt(userId)),
			Message: []robot.MessageItem{
				{
					Type: "text",
					Data: robot.MessageItemData{
						Text: c.Command + " " + c.Prompt,
					},
				},
			},
		},
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

// fail
func (c *Cron) ExecSlack(userId string) {
	if robot.SlackClient == nil {
		logger.Error("slack client is nil")
		return
	}
	
	t := &robot.SlackRobot{
		Event: &slackevents.MessageEvent{
			User: userId,
			Text: c.Command + " " + c.Prompt,
		},
		Client: robot.SlackClient,
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

// fail
func (c *Cron) ExecQQ(userId string) {
	if robot.QQApi == nil {
		logger.Warn("qq api is nil")
		return
	}
	
	t := &robot.QQRobot{
		RobotInfo: robot.QQRobotInfo,
		C2CMessage: &dto.WSC2CMessageData{
			Author: &dto.User{
				ID: userId,
			},
			Content: c.Command + " " + c.Prompt,
		},
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
}

// fail
func (c *Cron) ExecComWechat(userId string) {
	if robot.ComWechatApp == nil {
		logger.Warn("com wechat app is nil")
		return
	}
	
	t := &robot.ComWechatRobot{
		Event: models.CallbackMessageHeader{
			FromUserName: userId,
			MsgType:      models.CALLBACK_MSG_TYPE_TEXT,
		},
		App: robot.ComWechatApp,
		TextMsg: &serverModel.MessageText{
			Content: c.Command + " " + c.Prompt,
		},
	}
	t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
	t.Robot.Exec()
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
			FromUserName: userId,
			MsgType:      models.CALLBACK_MSG_TYPE_TEXT,
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
