package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	LogId   string      `json:"log_id"`
}

func Failure(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, message string, data interface{}) {
	_, logId := logger.GetBotNameAndLogId(ctx)
	
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
		LogId:   logId,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	json.NewEncoder(w).Encode(resp)
	metrics.HTTPResponseCount.WithLabelValues(r.URL.Path, strconv.Itoa(code)).Inc()
}

func Success(ctx context.Context, w http.ResponseWriter, r *http.Request, data interface{}) {
	_, logId := logger.GetBotNameAndLogId(ctx)
	
	resp := Response{
		Code:    param.CodeSuccess,
		Message: param.MsgSuccess,
		Data:    data,
		LogId:   logId,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	json.NewEncoder(w).Encode(resp)
	metrics.HTTPResponseCount.WithLabelValues(r.URL.Path, strconv.Itoa(param.CodeSuccess)).Inc()
}

func HandleJsonBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}
	
	return nil
}
