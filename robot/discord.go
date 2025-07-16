package robot

import (
	"encoding/json"
	"fmt"
	
	"github.com/bwmarrin/discordgo"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type DiscordRobot struct {
	session *discordgo.Session
	msg     *discordgo.MessageCreate
}

func StartDiscordRobot() {
	dg, err := discordgo.New("Bot " + *conf.BaseConfInfo.DiscordBotToken)
	if err != nil {
		logger.Fatal("create discord bot", "err", err)
	}
	
	if dg.State.User != nil {
		logger.Info("discordBot Info", dg.State.User.Username)
	}
	
	// 添加消息处理函数
	dg.AddHandler(messageCreate)
	
	// 打开连接
	err = dg.Open()
	if err != nil {
		logger.Fatal("connect fail", "err", err)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 忽略自己的消息
	d, _ := json.Marshal(m)
	fmt.Println(string(d))
	
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
	}
}
