package http

import (
	"fmt"
	"net/http"
	"strconv"
	
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
	
	realUserId := "-" + r.URL.Query().Get("userId")
	intUserId, _ := strconv.ParseInt(realUserId, 10, 64)
	
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
	l := llm.NewLLM(llm.WithChatId(intUserId), llm.WithUserId(realUserId), llm.WithMsgId(int(intUserId)),
		llm.WithHTTPChain(messageChan), llm.WithContent(prompt))
	go func() {
		defer close(messageChan)
		err = l.CallLLM()
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
