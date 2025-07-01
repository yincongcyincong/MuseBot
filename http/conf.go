package http

import (
	"net/http"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type UpdateConfParam struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func UpdateConf(w http.ResponseWriter, r *http.Request) {
	var updateConfparam UpdateConfParam
	if err := utils.HandleJsonBody(r, &updateConfparam); err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	var err error
	switch updateConfparam.Type {
	case "base":
		err = utils.SetStructFieldByJSONTag(conf.BaseConfInfo, updateConfparam.Key, updateConfparam.Value)
	case "audio":
		err = utils.SetStructFieldByJSONTag(conf.AudioConfInfo, updateConfparam.Key, updateConfparam.Value)
	case "llm":
		err = utils.SetStructFieldByJSONTag(conf.LLMConfInfo, updateConfparam.Key, updateConfparam.Value)
	case "photo":
		err = utils.SetStructFieldByJSONTag(conf.PhotoConfInfo, updateConfparam.Key, updateConfparam.Value)
	case "rag":
		err = utils.SetStructFieldByJSONTag(conf.RagConfInfo, updateConfparam.Key, updateConfparam.Value)
	case "video":
		err = utils.SetStructFieldByJSONTag(conf.VideoConfInfo, updateConfparam.Key, updateConfparam.Value)
	default:
		logger.Error("update conf error", "type", updateConfparam.Type)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	if err != nil {
		logger.Error("update conf error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	utils.Success(w, "")
}

func GetConf(w http.ResponseWriter, r *http.Request) {
	res := map[string]interface{}{}
	res["base"] = conf.BaseConfInfo
	res["audio"] = conf.AudioConfInfo
	res["llm"] = conf.LLMConfInfo
	res["photo"] = conf.PhotoConfInfo
	res["rag"] = conf.RagConfInfo
	res["video"] = conf.VideoConfInfo
	res["tools"] = conf.TaskTools
	
	utils.Success(w, res)
}
