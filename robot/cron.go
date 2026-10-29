package robot

import (
	"context"
	"github.com/yincongcyincong/MuseBot/conf"

	"encoding/json"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	serverModel "github.com/ArtisanCloud/PowerWeChat/v3/src/work/server/handlers/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack/slackevents"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

func InitCron() {
	time.Sleep(10 * time.Second)
	cronJobs, err := db.GetCronsByPage(1, 10000, "", "")
	if err != nil {
		logger.Error("get crons error", "err", err)
		return
	}

	Cron = cron.New(cron.WithSeconds())
	for _, c := range cronJobs {
		if c.CronSpec != "" && c.Status == 1 && c.Type != "" && c.Prompt != "" {
			cronID, err := Cron.AddFunc(c.CronSpec, func() {
				ExecCron(&c)
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

func ExecCron(c *db.Cron) {
	// one log id per cron fire, reused by every synthetic robot below so the
	// whole fire is traceable as one unit
	ctx := context.WithValue(context.WithValue(context.Background(),
		"bot_name", conf.BaseConfInfo.BotName), "log_id", uuid.New().String())
	logger.InfoCtx(ctx, "exec cron", "cron", c.CronName, "cronSpec",
		c.CronSpec, "type", c.Type, "targetId", c.TargetID, "groupId", c.GroupID)
	switch c.Type {
	case param.Wechat:
		ExecWechat(ctx, c)
	case param.Telegram:
		ExecTelegram(ctx, c)
	case param.ComWechat:
		ExecComWechat(ctx, c)
	case param.Lark:
		ExecLark(ctx, c)
	case param.PersonalQQ:
		ExecPersonalQQ(ctx, c)
	case param.Slack:
		ExecSlack(ctx, c)
	case param.Ding:
		ExecDing(ctx, c)
	}
}

func ExecDing(ctx context.Context, c *db.Cron) {
	if DingBotClient == nil {
		logger.ErrorCtx(ctx, "dingbot client is nil")
		return
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		t := &DingRobot{
			Message: &chatbot.BotCallbackDataModel{
				SenderId: c.CreateBy,
				Text: chatbot.BotCallbackDataTextModel{
					Content: c.Command + " " + c.Prompt,
				},
				Msgtype:        "text",
				SessionWebhook: targetId,
			},
			Client: DingBotClient,
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t),
			WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecLark(ctx context.Context, c *db.Cron) {
	if LarkBotClient == nil {
		logger.ErrorCtx(ctx, "larkbot client is nil")
		return
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		contentStr := MessageText{
			Text: c.Command + " " + c.Prompt,
		}
		contentByte, _ := json.Marshal(contentStr)
		content := string(contentByte)
		msgType := larkim.MsgTypeText
		t := &LarkRobot{
			Message: &larkim.P2MessageReceiveV1{
				Event: &larkim.P2MessageReceiveV1Data{
					Message: &larkim.EventMessage{
						Content:     &content,
						MessageType: &msgType,
						ChatId:      &targetId,
					},
					Sender: &larkim.EventSender{
						SenderId: &larkim.UserId{UserId: &c.CreateBy},
					},
				},
			},
			Client: LarkBotClient,
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t), WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecPersonalQQ(ctx context.Context, c *db.Cron) {
	for _, groupId := range strings.Split(c.GroupID, ",") {
		groupId = strings.TrimSpace(groupId)
		if groupId == "" {
			continue
		}
		t := &PersonalQQRobot{
			Msg: &QQMessage{
				GroupId:     groupId,
				MessageType: "group",
				UserID:      c.CreateBy,
				Message: []MessageItem{
					{
						Type: "text",
						Data: MessageItemData{
							Text: c.Command + " " + c.Prompt,
						},
					},
				},
			},
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t),
			WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		t := &PersonalQQRobot{
			Msg: &QQMessage{
				MessageType: "private",
				UserID:      targetId,
				Message: []MessageItem{
					{
						Type: "text",
						Data: MessageItemData{
							Text: c.Command + " " + c.Prompt,
						},
					},
				},
			},
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t),
			WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecSlack(ctx context.Context, c *db.Cron) {
	if SlackClient == nil {
		logger.ErrorCtx(ctx, "slack client is nil")
		return
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		t := &SlackRobot{
			Event: &slackevents.MessageEvent{
				User:    c.CreateBy,
				Channel: targetId,
				Text:    c.Command + " " + c.Prompt,
			},
			Client: SlackClient,
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t), WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecComWechat(ctx context.Context, c *db.Cron) {
	if ComWechatApp == nil {
		logger.WarnCtx(ctx, "com wechat app is nil")
		return
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		t := &ComWechatRobot{
			Event: models.CallbackMessageHeader{
				FromUserName: targetId,
				MsgType:      models.CALLBACK_MSG_TYPE_TEXT,
			},
			App: ComWechatApp,
			TextMsg: &serverModel.MessageText{
				Content: c.Command + " " + c.Prompt,
			},
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t), WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecTelegram(ctx context.Context, c *db.Cron) {
	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		if utils.ParseInt(targetId) == 0 {
			logger.WarnCtx(ctx, "cron target is not a valid telegram id, skip", "cron_id", c.ID, "targetId", targetId)
			continue
		}
		t := &TelegramRobot{
			Bot: TelegramBot,
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
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t), WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

	for _, groupId := range strings.Split(c.GroupID, ",") {
		groupId = strings.TrimSpace(groupId)
		if groupId == "" {
			continue
		}
		if utils.ParseInt(groupId) == 0 {
			logger.WarnCtx(ctx, "cron group is not a valid telegram id, skip", "cron_id", c.ID, "groupId", groupId)
			continue
		}
		t := &TelegramRobot{
			Bot: TelegramBot,
			Update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID: int64(utils.ParseInt(c.CreateBy)),
					},
					Chat: &tgbotapi.Chat{
						ID: int64(utils.ParseInt(groupId)),
					},
					Text: c.Command + " " + c.Prompt,
				},
			},
		}
		t.Robot = NewRobot(WithCtx(ctx), WithRobot(t), WithSkipCheck(true), WithUseRecord(false))
		t.Robot.Exec()
	}

}

func ExecWechat(ctx context.Context, c *db.Cron) {
	if OfficialAccountApp == nil {
		logger.WarnCtx(ctx, "official account app is nil", "type", c.Type, "userId", c.TargetID, "prompt", c.Prompt)
		return
	}

	for _, targetId := range strings.Split(c.TargetID, ",") {
		targetId = strings.TrimSpace(targetId)
		if targetId == "" {
			continue
		}
		w := &WechatRobot{
			Event: models.CallbackMessageHeader{
				FromUserName: targetId,
				MsgType:      models.CALLBACK_MSG_TYPE_TEXT,
			},
			App: OfficialAccountApp,
			TextMsg: &serverModel.MessageText{
				Content: c.Command + " " + c.Prompt,
			},
		}
		w.Robot = NewRobot(WithCtx(ctx), WithRobot(w), WithSkipCheck(true), WithUseRecord(false))
		w.Robot.Exec()
	}

}

func AddCron(cronInfo *db.Cron) error {
	cronJobId, err := Cron.AddFunc(cronInfo.CronSpec, func() {
		ExecCron(cronInfo)
	})
	if err != nil {
		return err
	}

	err = db.UpdateCronJobId(cronInfo.ID, int(cronJobId))
	if err != nil {
		return err
	}

	return nil

}
