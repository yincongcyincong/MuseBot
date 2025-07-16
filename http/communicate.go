package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

// Communicate handles the Server-Sent Events
func Communicate(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(w, "Missing prompt parameter", http.StatusBadRequest)
		return
	}
	
	userIdStr := r.URL.Query().Get("userId")
	userId, _ := strconv.Atoi(userIdStr)
	realUserId := userId * -1
	
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	// 获取 http.Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	var err error
	messageChan := make(chan string)
	l := llm.NewLLM(llm.WithChatId(int64(realUserId)), llm.WithUserId(int64(realUserId)), llm.WithMsgId(realUserId),
		llm.WithHTTPChain(messageChan), llm.WithContent(prompt))
	l.LLMClient.GetMessages(int64(realUserId), prompt)
	go func() {
		defer close(messageChan)
		err = l.LLMClient.Send(context.Background(), l)
		if err != nil {
			logger.Warn("Error sending message", "err", err)
		}
	}()
	
	for msg := range messageChan {
		_, err = fmt.Fprint(w, msg)
		if err != nil {
			logger.Warn("Error writing to SSE", "err", err)
		}
		flusher.Flush()
	}
	
	if err != nil {
		logger.Warn("Error writing to SSE", "err", err)
	}
}

func getMockTelegramBot(realUserId int) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: realUserId,
		
		Message: &tgbotapi.Message{
			MessageID: realUserId,
			Chat: &tgbotapi.Chat{
				ID: int64(realUserId),
			},
			From: &tgbotapi.User{
				ID: int64(realUserId),
			},
		},
	}
}
