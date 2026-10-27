package robot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestBuiltinChannelsRegistered(t *testing.T) {
	got := map[string]bool{}
	for _, ch := range channels {
		got[ch.Name] = true
	}
	for _, name := range []string{
		param.Telegram, param.Discord, param.Lark, param.Slack,
		param.Ding, param.ComWechat, param.QQ, param.Wechat,
	} {
		assert.True(t, got[name], "missing channel %s", name)
	}
}

func TestChannelMissingCfgReportsKeysOnEmptyConfig(t *testing.T) {
	pointers := map[string]*string{
		"TELEGRAM_BOT_TOKEN":          &conf.BaseConfInfo.TelegramBotToken,
		"DISCORD_BOT_TOKEN":           &conf.BaseConfInfo.DiscordBotToken,
		"LARK_APP_ID":                 &conf.BaseConfInfo.LarkAPPID,
		"LARK_APP_SECRET":             &conf.BaseConfInfo.LarkAppSecret,
		"SLACK_BOT_TOKEN":             &conf.BaseConfInfo.SlackBotToken,
		"SLACK_APP_TOKEN":             &conf.BaseConfInfo.SlackAppToken,
		"DING_CLIENT_ID":              &conf.BaseConfInfo.DingClientId,
		"DING_CLIENT_SECRET":          &conf.BaseConfInfo.DingClientSecret,
		"COM_WECHAT_SECRET":           &conf.BaseConfInfo.ComWechatSecret,
		"COM_WECHAT_AGENT_ID":         &conf.BaseConfInfo.ComWechatAgentID,
		"COM_WECHAT_ENCODING_AES_KEY": &conf.BaseConfInfo.ComWechatEncodingAESKey,
		"QQ_APP_ID":                   &conf.BaseConfInfo.QQAppID,
		"QQ_APP_SECRET":               &conf.BaseConfInfo.QQAppSecret,
		"WECHAT_APP_ID":               &conf.BaseConfInfo.WechatAppID,
		"WECHAT_APP_SECRET":           &conf.BaseConfInfo.WechatAppSecret,
	}
	previous := make(map[string]string, len(pointers))
	for key, p := range pointers {
		previous[key] = *p
		*p = ""
	}
	t.Cleanup(func() {
		for key, p := range pointers {
			*p = previous[key]
		}
	})

	for _, ch := range channels {
		missing := ch.MissingCfg()
		assert.NotEmpty(t, missing, "channel %s should report missing keys on empty config", ch.Name)
	}

	for _, ch := range channels {
		if ch.Name == param.Telegram {
			assert.Equal(t, []string{"TELEGRAM_BOT_TOKEN"}, ch.MissingCfg())
		}
	}
}

func TestChannelMissingCfgEmptyWhenConfigured(t *testing.T) {
	previous := conf.BaseConfInfo.TelegramBotToken
	conf.BaseConfInfo.TelegramBotToken = "token"
	t.Cleanup(func() { conf.BaseConfInfo.TelegramBotToken = previous })

	for _, ch := range channels {
		if ch.Name == param.Telegram {
			assert.Empty(t, ch.MissingCfg())
		}
	}
}

func TestRegisterChannelRejectsInvalid(t *testing.T) {
	assert.Panics(t, func() { RegisterChannel(Channel{Name: ""}) })
	assert.Panics(t, func() { RegisterChannel(Channel{Name: "x"}) })
	assert.Panics(t, func() {
		RegisterChannel(Channel{Name: "x", MissingCfg: func() []string { return nil }})
	})
	assert.Panics(t, func() {
		RegisterChannel(Channel{Name: "x", Start: func(ctx context.Context) {}})
	})

	t.Cleanup(func() { channels = channels[:len(channels)-1] })
	assert.NotPanics(t, func() {
		RegisterChannel(Channel{Name: "x", MissingCfg: func() []string { return nil },
			Start: func(ctx context.Context) {}})
	})
}
