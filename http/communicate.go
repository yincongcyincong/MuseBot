package http

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/robot"
)

// Communicate handles the Server-Sent Events
func Communicate(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Error reading request body", "err", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	
	if prompt == "" && len(fileData) == 0 {
		http.Error(w, "Missing prompt parameter", http.StatusBadRequest)
		return
	}
	
	realUserId := r.URL.Query().Get("user_id")
	intUserId, _ := strconv.ParseInt(realUserId, 10, 64)
	
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	command, p := parseCommand(prompt)
	
	web := robot.NewWeb(command, intUserId, realUserId, p, prompt, fileData, w, flusher)
	web.Exec()
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
