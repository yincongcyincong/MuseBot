package http

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	
	"github.com/yincongcyincong/mcp-client-go/clients"
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
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
	
	utils.Success(w, res)
}

func GetMCPConf(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(*conf.McpConfPath)
	if err != nil {
		logger.Error("read mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	config := new(mcpParam.McpClientGoConfig)
	err = json.Unmarshal(data, config)
	if err != nil {
		logger.Error("unmarshal mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	utils.Success(w, config)
}

func UpdateMCPConf(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	config := new(mcpParam.MCPConfig)
	err := utils.HandleJsonBody(r, config)
	if err != nil {
		logger.Error("parse json body error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	mcpConfigs, err := getMCPConf()
	if err != nil {
		logger.Error("get mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	mcpConfigs.McpServers[name] = config
	
	err = updateMCPConfFile(mcpConfigs)
	if err != nil {
		logger.Error("update mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	go updateMCPConf(name, config)
	utils.Success(w, "")
}

func DeleteMCPConf(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	
	mcpConfigs, err := getMCPConf()
	if err != nil {
		logger.Error("get mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	if mcpConfigs.McpServers[name] != nil && !mcpConfigs.McpServers[name].Disabled {
		err = clients.RemoveMCPClient(name)
		if err != nil {
			logger.Error("remove mcp client error", "err", err)
			utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
			return
		}
	}
	
	delete(mcpConfigs.McpServers, name)
	conf.TaskTools.Delete(name)
	
	err = updateMCPConfFile(mcpConfigs)
	if err != nil {
		logger.Error("update mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	utils.Success(w, "")
}

func DisableMCPConf(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	disable := r.URL.Query().Get("disable")
	
	config, err := getMCPConf()
	if err != nil {
		logger.Error("get mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	if disable == "1" {
		for mcpName, client := range config.McpServers {
			if mcpName == name {
				client.Disabled = true
				conf.TaskTools.Delete(name)
				err = clients.RemoveMCPClient(name)
				if err != nil {
					logger.Error("remove mcp client error", "err", err)
					utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
					return
				}
			}
		}
	} else {
		for mcpName, client := range config.McpServers {
			if mcpName == name {
				client.Disabled = false
				go updateMCPConf(name, client)
			}
		}
	}
	
	err = updateMCPConfFile(config)
	if err != nil {
		logger.Error("update mcp conf error", "err", err)
		utils.Failure(w, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	utils.Success(w, "")
}

func SyncMCPConf(w http.ResponseWriter, r *http.Request) {
	clients.ClearAllMCPClient()
	conf.TaskTools.Clear()
	conf.InitTools()
	utils.Success(w, "")
}

func getMCPConf() (*mcpParam.McpClientGoConfig, error) {
	data, err := os.ReadFile(*conf.McpConfPath)
	if err != nil {
		logger.Error("read mcp conf error", "err", err)
		return nil, err
	}
	
	config := new(mcpParam.McpClientGoConfig)
	err = json.Unmarshal(data, config)
	if err != nil {
		logger.Error("unmarshal mcp conf error", "err", err)
		return nil, err
	}
	
	return config, nil
}

func updateMCPConfFile(config *mcpParam.McpClientGoConfig) error {
	file, err := os.OpenFile(*conf.McpConfPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("open mcp conf error", "err", err)
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(config); err != nil {
		logger.Error("encode mcp conf error", "err", err)
		return err
	}
	
	return nil
}

func updateMCPConf(name string, config *mcpParam.MCPConfig) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("update mcp conf error", "err", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	mcpCleintConf := clients.GetOneMCPClient(name, config)
	errs := clients.RegisterMCPClient(ctx, []*mcpParam.MCPClientConf{mcpCleintConf})
	if len(errs) > 0 {
		for mcpServer, err := range errs {
			logger.Error("register mcp client error", "server", mcpServer, "error", err)
		}
	}
	conf.InsertTools(name)
	return
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
