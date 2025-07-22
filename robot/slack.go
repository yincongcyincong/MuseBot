package robot

import (
	"context"
	"runtime/debug"
	"strings"
	"time"
	
	"github.com/slack-go/slack"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/rag"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type SlackRobot struct {
	Client *slack.Client
	Event  *slack.MessageEvent
	Robot  *RobotInfo
}

func StartSlackBot() {
	api := slack.New(*conf.BaseConfInfo.SlackBotToken)
	logger.Info("Slack bot started")
	
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if ev.User == "" || ev.BotID != "" {
				continue
			}
			r := &SlackRobot{Client: api, Event: ev}
			r.Robot = NewRobot(WithRobot(r))
			r.Robot.Exec()
		}
	}
}

func (r *SlackRobot) Exec() {
	text := strings.TrimSpace(r.Event.Text)
	if text == "" {
		return
	}
	r.requestDeepseekAndResp(text)
}

func (r *SlackRobot) requestDeepseekAndResp(content string) {
	chatID, _, userID := r.Robot.GetChatIdAndMsgIdAndUserID()
	
	if r.Robot.checkUserTokenExceed(chatID, 0, userID) {
		logger.Warn("user token exceed", "userID", userID)
		return
	}
	
	if conf.RagConfInfo.Store != nil {
		r.executeChain(content)
	} else {
		r.executeLLM(content)
	}
}

func (r *SlackRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	chatID, _, userID := r.Robot.GetChatIdAndMsgIdAndUserID()
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic", "err", err, "stack", string(debug.Stack()))
			}
			utils.DecreaseUserChat(userID)
		}()
		
		if utils.CheckUserChatExceed(userID) {
			r.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil), 0, "", nil)
			return
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		text := content
		dpLLM := rag.NewRag(
			llm.WithMessageChan(messageChan),
			llm.WithContent(content),
			llm.WithChatId(chatID),
			llm.WithUserId(userID),
		)
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err := chains.Run(ctx, qaChain, text)
		if err != nil {
			r.Robot.SendMsg(chatID, err.Error(), 0, "", nil)
		}
	}()
	
	go r.handleUpdate(messageChan)
}

func (r *SlackRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	go r.callLLM(content, messageChan)
	go r.handleUpdate(messageChan)
}

func (r *SlackRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	for msg := range messageChan {
		if msg.Content == "" {
			msg.Content = "get nothing from deepseek!"
		}
	}
}

func (r *SlackRobot) callLLM(content string, messageChan chan *param.MsgInfo) {
	chatID, _, userID := r.Robot.GetChatIdAndMsgIdAndUserID()
	
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic", "err", err, "stack", string(debug.Stack()))
		}
		utils.DecreaseUserChat(userID)
	}()
	
	if utils.CheckUserChatExceed(userID) {
		r.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil), 0, "", nil)
		return
	}
	
	l := llm.NewLLM(
		llm.WithMessageChan(messageChan),
		llm.WithContent(content),
		llm.WithChatId(chatID),
		llm.WithUserId(userID),
		llm.WithTaskTools(&conf.AgentInfo{
			DeepseekTool:    conf.DeepseekTools,
			VolTool:         conf.VolTools,
			OpenAITools:     conf.OpenAITools,
			GeminiTools:     conf.GeminiTools,
			OpenRouterTools: conf.OpenRouterTools,
		}),
	)
	
	err := l.CallLLM()
	if err != nil {
		logger.Error("callLLM error", "err", err)
		r.Robot.SendMsg(chatID, err.Error(), 0, "", nil)
	}
}
