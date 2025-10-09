package main

import (
	"context"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/robot"
)

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	os.Exit(code)
}

func setup() {
	conf.InitConf()
	db.InitTable()
	i18n.InitI18n()
}

func TestSendTelegramMsg(t *testing.T) {
	messageChan := make(chan *param.MsgInfo)

	go func() {
		bot := robot.CreateBot(context.Background())
		t := robot.NewTelegramRobot(tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From: &tgbotapi.User{
					ID: 5542540980,
				},
				Chat: &tgbotapi.Chat{
					ID: 5542540980,
				},
			},
		}, bot)
		t.Robot = robot.NewRobot(robot.WithRobot(t))
		t.Robot.HandleUpdate(&robot.MsgChan{
			NormalMessageChan: messageChan,
		}, "")
	}()

	*conf.BaseConfInfo.Type = param.DeepSeek

	callLLM := llm.NewLLM(llm.WithChatId("1"), llm.WithMsgId("2"), llm.WithUserId("3"),
		llm.WithMessageChan(messageChan), llm.WithContent("hi"))
	callLLM.LLMClient.GetModel(callLLM)
	callLLM.GetMessages("3", "hi")
	err := callLLM.LLMClient.Send(context.Background(), callLLM)
	assert.Equal(t, nil, err)
}
