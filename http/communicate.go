package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	
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
	
	//imageData := r.URL.Query().Get("image")
	
	realUserId := "-" + r.URL.Query().Get("userId")
	intUserId, _ := strconv.ParseInt(realUserId, 10, 64)
	
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	// 命令模式处理
	var err error
	command, args := parseCommand(prompt)
	switch command {
	case "/photo":
		err = handlePhoto(w, flusher, intUserId, realUserId, args)
		return
	case "/video":
		//handleVideo(w, flusher, intUserId, realUserId, args)
		return
	case "/mcp":
		//handleMCP(w, flusher, intUserId, realUserId, args)
		return
	default:
		err = handleChat(w, flusher, intUserId, realUserId, prompt)
	}
	
	if err != nil {
		logger.Warn("Error writing to SSE", "err", err)
	}
}

func handleChat(w http.ResponseWriter, flusher http.Flusher, intUserId int64, realUserId string, prompt string) error {
	// 默认：处理普通 LLM 流式回复
	var err error
	messageChan := make(chan string)
	l := llm.NewLLM(
		llm.WithChatId(intUserId),
		llm.WithUserId(realUserId),
		llm.WithMsgId(int(intUserId)),
		llm.WithHTTPChain(messageChan),
		llm.WithContent(prompt),
	)
	go func() {
		defer close(messageChan)
		err = l.CallLLM()
		if err != nil {
			logger.Warn("Error sending message", "err", err)
		}
	}()
	
	for msg := range messageChan {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
	
	return err
}

// parseCommand extracts command and arguments like /photo xxx
func parseCommand(prompt string) (command string, args string) {
	if len(prompt) == 0 || prompt[0] != '/' {
		return "", prompt
	}
	parts := strings.SplitN(prompt, " ", 2)
	command = parts[0]
	if len(parts) > 1 {
		args = parts[1]
	}
	return command, args
}

// handlePhoto generates and streams image in base64
func handlePhoto(w http.ResponseWriter, flusher http.Flusher, userId int64, realUserId, prompt string) error {
	// 生成图片并流式传输
	return nil
	
}
