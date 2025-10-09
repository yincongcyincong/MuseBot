package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/yincongcyincong/MuseBot/param"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Failure(w http.ResponseWriter, code int, message string, data interface{}) {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 可选，默认就是 200

	json.NewEncoder(w).Encode(resp)
}

func Success(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Code:    param.CodeSuccess,
		Message: param.MsgSuccess,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 可选，默认就是 200

	json.NewEncoder(w).Encode(resp)
}

func HandleJsonBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// 解析 JSON 数据到结构体
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}
