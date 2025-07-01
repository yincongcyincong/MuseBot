package http

import (
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	
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

func GetCommand(w http.ResponseWriter, r *http.Request) {
	res := CompareFlagsWithStructTags(conf.BaseConfInfo)
	res += CompareFlagsWithStructTags(conf.AudioConfInfo)
	res += CompareFlagsWithStructTags(conf.LLMConfInfo)
	res += CompareFlagsWithStructTags(conf.PhotoConfInfo)
	res += CompareFlagsWithStructTags(conf.RagConfInfo)
	res += CompareFlagsWithStructTags(conf.VideoConfInfo)
	res += fmt.Sprintf("-mcp_conf_path=%s", *conf.McpConfPath)
	utils.Success(w, res)
}

func UpdateConf(w http.ResponseWriter, r *http.Request) {
	var updateConfParam UpdateConfParam
	if err := utils.HandleJsonBody(r, &updateConfParam); err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	handleSpecialData(&updateConfParam)
	
	var err error
	switch updateConfParam.Type {
	case "base":
		err = utils.SetStructFieldByJSONTag(conf.BaseConfInfo, updateConfParam.Key, updateConfParam.Value)
	case "audio":
		err = utils.SetStructFieldByJSONTag(conf.AudioConfInfo, updateConfParam.Key, updateConfParam.Value)
	case "llm":
		err = utils.SetStructFieldByJSONTag(conf.LLMConfInfo, updateConfParam.Key, updateConfParam.Value)
	case "photo":
		err = utils.SetStructFieldByJSONTag(conf.PhotoConfInfo, updateConfParam.Key, updateConfParam.Value)
	case "rag":
		err = utils.SetStructFieldByJSONTag(conf.RagConfInfo, updateConfParam.Key, updateConfParam.Value)
	case "video":
		err = utils.SetStructFieldByJSONTag(conf.VideoConfInfo, updateConfParam.Key, updateConfParam.Value)
	default:
		logger.Error("update conf error", "type", updateConfParam.Type)
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

func handleSpecialData(updateConfParam *UpdateConfParam) {
	switch updateConfParam.Key {
	case "allowed_telegram_user_ids", "allowed_telegram_group_ids", "admin_user_ids":
		ids := strings.Split(updateConfParam.Value.(string), ",")
		idMap := make(map[int64]bool)
		for _, idStr := range ids {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			idMap[int64(id)] = true
		}
		updateConfParam.Value = idMap
	case "stop":
		updateConfParam.Value = strings.Split(updateConfParam.Value.(string), ",")
	}
}

func CompareFlagsWithStructTags(cfg interface{}) string {
	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)
	
	// If it's a pointer, get the element it points to
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Println("Input is a nil pointer")
			return ""
		}
		v = v.Elem()
		t = t.Elem()
	}
	
	if t.Kind() != reflect.Struct {
		fmt.Println("Input must be a struct or pointer to struct")
		return ""
	}
	
	res := ""
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		
		flagValue := flag.Lookup(jsonTag)
		if flagValue == nil {
			logger.Warn("Flag not found", "jsonTag", jsonTag)
			continue
		}
		
		structValue := ""
		switch jsonTag {
		case "allowed_telegram_user_ids", "allowed_telegram_group_ids", "admin_user_ids":
			structValue = utils.MapKeysToString(v.Field(i).Interface())
		default:
			structValue = utils.ValueToString(v.Field(i).Interface())
		}
		
		if structValue != flagValue.DefValue {
			res += fmt.Sprintf("-%s=%s ", jsonTag, structValue)
		}
	}
	
	return res
}
