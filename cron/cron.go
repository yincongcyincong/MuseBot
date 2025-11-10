package cron

import (
	"encoding/json"
	"strings"
	"time"
	
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack/slackevents"
	"github.com/tencent-connect/botgo/dto"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	Cron = cron.New(cron.WithSeconds())
)

func InitCron() {
	time.Sleep(10 * time.Second)
	cronJobs, err := db.GetCronsByPage(1, 10000, "")
	if err != nil {
		logger.Error("get crons error", "err", err)
		return
	}
	
	for _, c := range cronJobs {
		if c.CronSpec != "" && c.Status == 1 && c.Type != "" && c.TargetID != "" && c.Prompt != "" {
			cronID, err := Cron.AddFunc(c.CronSpec, func() {
				Exec(&c)
			})
			if err != nil {
				logger.Error("crontab parse error", "err", err)
				continue
			}
			
			err = db.UpdateCronJobId(c.ID, int(cronID))
			if err != nil {
				logger.Error("update cron job id error", "err", err)
			}
		}
	}
	
	Cron.Start()
}

func Exec(c *db.Cron) {
	switch c.Type {
	case param.Wechat:
		ExecWechat(c)
	case param.Telegram:
		ExecTelegram(c)
	case param.ComWechat:
		ExecComWechat(c)
	case param.Ding:
		ExecDing(c)
	case param.Lark:
		ExecLark(c)
	case param.PersonalQQ:
		ExecPersonalQQ(c)
	case param.Slack:
		ExecSlack(c)
	}
}

func ExecDing(c *db.Cron) {
	if robot.DingBotClient == nil {
		logger.Error("dingbot client is nil")
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.DingRobot{
			Message: &chatbot.BotCallbackDataModel{
				SenderId: targetId,
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
	
	for _, targetId := range strings.Split(c.GroupID, ",") {
		t := &robot.DingRobot{
			Message: &chatbot.BotCallbackDataModel{
				SenderId: targetId,
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
	
}

func ExecLark(c *db.Cron) {
	if robot.LarkBotClient == nil {
		logger.Error("larkbot client is nil")
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
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
						ChatId:      &targetId,
					},
					Sender: &larkim.EventSender{
						SenderId: &larkim.UserId{UserId: &targetId},
					},
				},
			},
			Client: robot.LarkBotClient,
		}
		t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
		t.Robot.Exec()
	}
	
}

func ExecPersonalQQ(c *db.Cron) {
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.PersonalQQRobot{
			Msg: &robot.QQMessage{
				GroupId:     int64(utils.ParseInt(targetId)),
				MessageType: "group",
				UserID:      int64(utils.ParseInt(targetId)),
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
	
	for _, targetId := range strings.Split(c.GroupID, ",") {
		t := &robot.PersonalQQRobot{
			Msg: &robot.QQMessage{
				MessageType: "private",
				UserID:      int64(utils.ParseInt(targetId)),
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
	
}

func ExecSlack(c *db.Cron) {
	if robot.SlackClient == nil {
		logger.Error("slack client is nil")
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.SlackRobot{
			Event: &slackevents.MessageEvent{
				User:    targetId,
				Channel: targetId,
				Text:    c.Command + " " + c.Prompt,
			},
			Client: robot.SlackClient,
		}
		t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
		t.Robot.Exec()
	}
	
}

// fail
func ExecQQ(c *db.Cron) {
	if robot.QQApi == nil {
		logger.Warn("qq api is nil")
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.QQRobot{
			RobotInfo: robot.QQRobotInfo,
			C2CMessage: &dto.WSC2CMessageData{
				Author: &dto.User{
					ID: targetId,
				},
				Content: c.Command + " " + c.Prompt,
			},
		}
		t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
		t.Robot.Exec()
	}
	
}

func ExecComWechat(c *db.Cron) {
	if robot.ComWechatApp == nil {
		logger.Warn("com wechat app is nil")
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.ComWechatRobot{
			Event: models.CallbackMessageHeader{
				FromUserName: targetId,
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
	
}

func ExecTelegram(c *db.Cron) {
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.TelegramRobot{
			Bot: robot.TelegramBot,
			Update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID: int64(utils.ParseInt(targetId)),
					},
					Chat: &tgbotapi.Chat{
						ID: int64(utils.ParseInt(targetId)),
					},
					Text: c.Command + " " + c.Prompt,
				},
			},
		}
		t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
		t.Robot.Exec()
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &robot.TelegramRobot{
			Bot: robot.TelegramBot,
			Update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID: int64(utils.ParseInt(targetId)),
					},
					Chat: &tgbotapi.Chat{
						ID: int64(utils.ParseInt(targetId)),
					},
					Text: c.Command + " " + c.Prompt,
				},
			},
		}
		t.Robot = robot.NewRobot(robot.WithRobot(t), robot.WithSkipCheck(true))
		t.Robot.Exec()
	}
	
}

func ExecWechat(c *db.Cron) {
	if robot.OfficialAccountApp == nil {
		logger.Warn("official account app is nil", "type", c.Type, "userId", c.TargetID, "prompt", c.Prompt)
		return
	}
	
	for _, targetId := range strings.Split(c.TargetID, ",") {
		w := &robot.WechatRobot{
			Event: models.CallbackMessageHeader{
				FromUserName: targetId,
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
	
}
