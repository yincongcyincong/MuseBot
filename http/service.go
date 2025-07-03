package http

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

func PongHandler(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, "pong")
}
