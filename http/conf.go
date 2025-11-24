package http

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
)

type UpdateConfParam struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func GetCommand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	useQuota := r.URL.Query().Get("use_quota") == "1"
	res := CompareFlagsWithStructTags(conf.BaseConfInfo, useQuota)
	res += CompareFlagsWithStructTags(conf.AudioConfInfo, useQuota)
	res += CompareFlagsWithStructTags(conf.LLMConfInfo, useQuota)
	res += CompareFlagsWithStructTags(conf.PhotoConfInfo, useQuota)
	res += CompareFlagsWithStructTags(conf.RagConfInfo, useQuota)
	res += CompareFlagsWithStructTags(conf.VideoConfInfo, useQuota)
	
	flagValue := flag.Lookup("mcp_conf_path")
	if flagValue.DefValue != *conf.ToolsConfInfo.McpConfPath {
		res += fmt.Sprintf("-mcp_conf_path=%s", *conf.ToolsConfInfo.McpConfPath)
	}
	utils.Success(ctx, w, r, res)
}

func UpdateConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var updateConfParam UpdateConfParam
	if err := utils.HandleJsonBody(r, &updateConfParam); err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
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
		logger.ErrorCtx(ctx, "update conf error", "type", updateConfParam.Type)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	if err != nil {
		logger.ErrorCtx(ctx, "update conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	utils.Success(ctx, w, r, "")
}

func GetConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	res := map[string]interface{}{}
	res["base"] = conf.BaseConfInfo
	res["audio"] = conf.AudioConfInfo
	res["llm"] = conf.LLMConfInfo
	res["photo"] = conf.PhotoConfInfo
	res["rag"] = conf.RagConfInfo
	res["video"] = conf.VideoConfInfo
	
	utils.Success(ctx, w, r, res)
}

func GetMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, err := os.ReadFile(*conf.ToolsConfInfo.McpConfPath)
	if err != nil {
		logger.ErrorCtx(ctx, "read mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	config := new(mcpParam.McpClientGoConfig)
	err = json.Unmarshal(data, config)
	if err != nil {
		logger.ErrorCtx(ctx, "unmarshal mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	utils.Success(ctx, w, r, config)
}

func UpdateMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")
	config := new(mcpParam.MCPConfig)
	err := utils.HandleJsonBody(r, config)
	if err != nil {
		logger.ErrorCtx(ctx, "parse json body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	mcpConfigs, err := getMCPConf(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	mcpConfigs.McpServers[name] = config
	
	mcpClientConf := clients.GetOneMCPClient(name, config)
	if mcpClientConf == nil {
		logger.ErrorCtx(ctx, "get mcp client error")
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	err = updateMCPConfFile(ctx, mcpConfigs)
	if err != nil {
		logger.ErrorCtx(ctx, "update mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	go updateMCPConf(ctx, name, mcpClientConf)
	utils.Success(ctx, w, r, "")
}

func DeleteMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")
	
	mcpConfigs, err := getMCPConf(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	if mcpConfigs.McpServers[name] != nil {
		err = clients.RemoveMCPClient(name)
		if err != nil {
			logger.ErrorCtx(ctx, "remove mcp client error", "err", err)
		}
	}
	
	delete(mcpConfigs.McpServers, name)
	conf.TaskTools.Delete(name)
	
	err = updateMCPConfFile(ctx, mcpConfigs)
	if err != nil {
		logger.ErrorCtx(ctx, "update mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	utils.Success(ctx, w, r, "")
}

func DisableMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")
	disable := r.URL.Query().Get("disable")
	
	config, err := getMCPConf(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "get mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	if disable == "1" {
		for mcpName, client := range config.McpServers {
			if mcpName == name {
				client.Disabled = true
				conf.TaskTools.Delete(name)
				err = clients.RemoveMCPClient(name)
				if err != nil {
					logger.ErrorCtx(ctx, "remove mcp client error", "err", err)
				}
			}
		}
	} else {
		for mcpName, client := range config.McpServers {
			if mcpName == name {
				client.Disabled = false
				mcpClientConf := clients.GetOneMCPClient(name, client)
				if mcpClientConf == nil {
					logger.ErrorCtx(ctx, "get mcp client error")
					utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
					return
				}
				go updateMCPConf(ctx, name, mcpClientConf)
			}
		}
	}
	
	err = updateMCPConfFile(ctx, config)
	if err != nil {
		logger.ErrorCtx(ctx, "update mcp conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeConfigError, param.MsgConfigError, err)
		return
	}
	
	utils.Success(ctx, w, r, "")
}

func SyncMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	clients.ClearAllMCPClient()
	conf.TaskTools.Clear()
	conf.InitTools()
	utils.Success(ctx, w, r, "")
}

func getMCPConf(ctx context.Context) (*mcpParam.McpClientGoConfig, error) {
	data, err := os.ReadFile(*conf.ToolsConfInfo.McpConfPath)
	if err != nil {
		logger.ErrorCtx(ctx, "read mcp conf error", "err", err)
		return nil, err
	}
	
	config := new(mcpParam.McpClientGoConfig)
	err = json.Unmarshal(data, config)
	if err != nil {
		logger.ErrorCtx(ctx, "unmarshal mcp conf error", "err", err)
		return nil, err
	}
	
	return config, nil
}

func updateMCPConfFile(ctx context.Context, config *mcpParam.McpClientGoConfig) error {
	file, err := os.OpenFile(*conf.ToolsConfInfo.McpConfPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.ErrorCtx(ctx, "open mcp conf error", "err", err)
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(config); err != nil {
		logger.ErrorCtx(ctx, "encode mcp conf error", "err", err)
		return err
	}
	
	return nil
}

func updateMCPConf(ctx context.Context, name string, mcpClientConf *mcpParam.MCPClientConf) {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(ctx, "update mcp conf error", "err", err, "stack", string(debug.Stack()))
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	errs := clients.RegisterMCPClient(ctx, []*mcpParam.MCPClientConf{mcpClientConf})
	if len(errs) > 0 {
		for mcpServer, err := range errs {
			logger.ErrorCtx(ctx, "register mcp client error", "server", mcpServer, "error", err)
		}
	}
	conf.InsertTools(name)
	return
}

func handleSpecialData(updateConfParam *UpdateConfParam) {
	switch updateConfParam.Key {
	case "allowed_user_ids", "allowed_group_ids", "admin_user_ids":
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

func CompareFlagsWithStructTags(cfg interface{}, useQuota bool) string {
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
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		
		flagValue := flag.Lookup(jsonTag)
		if flagValue == nil {
			logger.Warn("Flag not found", "jsonTag", jsonTag)
			continue
		}
		
		structValue := ""
		switch jsonTag {
		case "allowed_user_ids", "allowed_group_ids", "admin_user_ids":
			structValue = utils.MapKeysToString(v.Field(i).Interface())
		default:
			structValue = utils.ValueToString(v.Field(i).Interface())
		}
		
		if structValue != flagValue.DefValue || jsonTag == "bot_name" || jsonTag == "http_host" {
			if useQuota {
				res += fmt.Sprintf("-%s='%s'\n", jsonTag, structValue)
			} else {
				res += fmt.Sprintf("-%s=%s\n", jsonTag, structValue)
			}
			
		}
	}
	
	return res
}
