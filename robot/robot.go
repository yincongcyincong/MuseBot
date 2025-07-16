package robot

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Robot struct {
	TelegramRobot *TelegramRobot
	DiscordRobot  *DiscordRobot
	
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

func (r *Robot) GetChatIdAndMsgIdAndUserID() (int64, int, int64) {
	chatId := int64(0)
	msgId := 0
	userId := int64(0)
	
	switch {
	case r.TelegramRobot != nil:
		if r.TelegramRobot.Update.Message != nil {
			chatId = r.TelegramRobot.Update.Message.Chat.ID
			userId = r.TelegramRobot.Update.Message.From.ID
			msgId = r.TelegramRobot.Update.Message.MessageID
		}
		if r.TelegramRobot.Update.CallbackQuery != nil {
			chatId = r.TelegramRobot.Update.CallbackQuery.Message.Chat.ID
			userId = r.TelegramRobot.Update.CallbackQuery.From.ID
			msgId = r.TelegramRobot.Update.CallbackQuery.Message.MessageID
		}
	case r.DiscordRobot != nil:
		if r.DiscordRobot.msg != nil {
			chatId, _ = strconv.ParseInt(r.DiscordRobot.msg.ChannelID, 10, 64)
			userId, _ = strconv.ParseInt(r.DiscordRobot.msg.Author.ID, 10, 64)
			msgId, _ = strconv.Atoi(r.DiscordRobot.msg.Message.ID)
		}
	}
	
	return chatId, msgId, userId
}

func (r *Robot) GetContent(content string) (string, error) {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	// check user chat exceed max count
	if CheckUserChatExceed(userId) {
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return "", errors.New("token exceed")
	}
	
	if content == "" && r.TelegramRobot.Update.Message.Voice != nil && *conf.AudioConfInfo.AudioAppID != "" {
		audioContent := r.GetAudioContent()
		if audioContent == nil {
			logger.Warn("audio url empty")
			return "", errors.New("audio url empty")
		}
		content = utils.FileRecognize(audioContent)
	}
	
	if content == "" && r.TelegramRobot.Update.Message.Photo != nil {
		imageContent, err := utils.GetImageContent(r.GetPhotoContent())
		if err != nil {
			logger.Warn("get image content err", "err", err)
			return "", err
		}
		content = imageContent
	}
	
	if content == "" {
		logger.Warn("content empty")
		return "", errors.New("content empty")
	}
	
	text := strings.ReplaceAll(content, "@"+r.TelegramRobot.Bot.Self.UserName, "")
	return text, nil
}

func (r *Robot) GetAudioContent() []byte {
	if r.TelegramRobot.Update.Message == nil || r.TelegramRobot.Update.Message.Voice == nil {
		return nil
	}
	
	fileID := r.TelegramRobot.Update.Message.Voice.FileID
	file, err := r.TelegramRobot.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}
	
	// 构造下载 URL
	downloadURL := file.Link(r.TelegramRobot.Bot.Token)
	
	transport := &http.Transport{}
	
	if *conf.BaseConfInfo.TelegramProxy != "" {
		proxy, err := url.Parse(*conf.BaseConfInfo.TelegramProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // 设置超时
	}
	
	// 通过代理下载
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	voice, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	return voice
}

func (r *Robot) GetPhotoContent() []byte {
	if r.TelegramRobot.Update.Message == nil || r.TelegramRobot.Update.Message.Photo == nil {
		return nil
	}
	
	var photo tgbotapi.PhotoSize
	for i := len(r.TelegramRobot.Update.Message.Photo) - 1; i >= 0; i-- {
		if r.TelegramRobot.Update.Message.Photo[i].FileSize < 8*1024*1024 {
			photo = r.TelegramRobot.Update.Message.Photo[i]
			break
		}
	}
	
	fileID := photo.FileID
	file, err := r.TelegramRobot.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}
	
	downloadURL := file.Link(r.TelegramRobot.Bot.Token)
	
	client := utils.GetTelegramProxyClient()
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	photoContent, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	
	return photoContent
}

func (r *Robot) SendMsg(chatId int64, msgContent string, replyToMessageID int,
	parseMode string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) int {
	switch {
	case r.TelegramRobot != nil:
		msg := tgbotapi.NewMessage(chatId, msgContent)
		msg.ParseMode = parseMode
		msg.ReplyMarkup = inlineKeyboard
		msg.ReplyToMessageID = replyToMessageID
		msgInfo, err := r.TelegramRobot.Bot.Send(msg)
		if err != nil {
			logger.Warn("send clear message fail", "err", err)
			return 0
		}
		return msgInfo.MessageID
	case r.DiscordRobot != nil:
		//SendMsgToDiscord(chatId, msgContent, r.DiscordRobot, replyToMessageID, parseMode)
	}
	
	return 0
}

func WithTelegramRobot(TelegramBot *TelegramRobot) func(*Robot) {
	return func(r *Robot) {
		r.TelegramRobot = TelegramBot
	}
}

func WithDiscordRobot(discordBot *DiscordRobot) func(*Robot) {
	return func(r *Robot) {
		r.DiscordRobot = discordBot
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
