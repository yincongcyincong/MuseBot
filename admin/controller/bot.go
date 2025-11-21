package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/admin/checkpoint"
	adminConf "github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/admin/db"
	adminUtils "github.com/yincongcyincong/MuseBot/admin/utils"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
)

type Bot struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
	CrtFile string `json:"crt_file"`
	KeyFile string `json:"key_file"`
	CaFile  string `json:"ca_file"`
	Command string `json:"command"`
	IsStart bool   `json:"is_start"`
}

type RegisterBot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	Status     string `json:"status"`
}

type GetBotConfRes struct {
	Data struct {
		Base  *conf.BaseConf  `json:"base"`
		Audio *conf.AudioConf `json:"audio"`
		LLM   *conf.LLMConf   `json:"llm"`
		Photo *conf.PhotoConf `json:"photo"`
		Video *conf.VideoConf `json:"video"`
	} `json:"data"`
}

var (
	SkipKey = map[string]bool{"bot": true}
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	day := r.URL.Query().Get("day")
	if day == "" {
		day = "7"
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/dashboard?day=%s", day), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func CreateBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var b Bot
	err := utils.HandleJsonBody(r, &b)
	if err != nil {
		logger.ErrorCtx(ctx, "create bot error", "bot", b)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	commands := adminUtils.ParseCommand(b.Command)
	if len(commands) == 0 || commands["bot_name"] == "" || commands["http_host"] == "" {
		logger.ErrorCtx(ctx, "create bot error", "commands", commands)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, errors.New("command is empty"))
		return
	}
	
	isRuning := false
	b.Name = commands["bot_name"]
	b.Address = utils.NormalizeHTTP(commands["http_host"])
	resp, err := adminUtils.GetCrtClient(&db.Bot{
		CaFile:  b.CaFile,
		CrtFile: b.CrtFile,
		KeyFile: b.KeyFile,
	}).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(b.Address, "/")+"/command/get?use_quota=1", bytes.NewBuffer(nil)))
	if err == nil {
		isRuning = true
		defer resp.Body.Close()
		bodyByte, err := io.ReadAll(resp.Body)
		httpRes := new(utils.Response)
		err = json.Unmarshal(bodyByte, httpRes)
		if err == nil {
			b.Command, _ = httpRes.Data.(string)
		}
	}
	
	err = db.CreateBot(b.Address, b.Name, b.CrtFile, b.KeyFile, b.CaFile, b.Command)
	if err != nil {
		logger.ErrorCtx(ctx, "create bot error", "reason", "db fail", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	go checkpoint.ScheduleBotChecks()
	
	if b.IsStart && !isRuning {
		err = adminUtils.StartDetachedProcess(b.Command)
		if err != nil {
			logger.ErrorCtx(ctx, "start bot error", "err", err)
			utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	}
	
	utils.Success(ctx, w, r, "bot created")
}

func RestartBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	params := r.URL.Query().Get("params")
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		"/restart?params="+url.QueryEscape(params), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func StopBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/stop", bytes.NewBuffer(nil)))
	utils.Success(ctx, w, r, "bot stopped")
}

func GetBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		logger.ErrorCtx(ctx, "get bot error", "id", idStr)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, errors.New("empty id"))
		return
	}
	
	bot, err := db.GetBotByID(idStr)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot error", "reason", "not found", "id", idStr, "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(ctx, w, r, bot)
}

func UpdateBotAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var b Bot
	err := utils.HandleJsonBody(r, &b)
	if err != nil {
		logger.ErrorCtx(ctx, "update bot error", "bot", b, "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	if b.Command == "" {
		resp, err := adminUtils.GetCrtClient(&db.Bot{
			CaFile:  b.CaFile,
			CrtFile: b.CrtFile,
			KeyFile: b.KeyFile,
		}).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(b.Address, "/")+"/command/get?use_quota=1", bytes.NewBuffer(nil)))
		if err == nil {
			defer resp.Body.Close()
			bodyByte, err := io.ReadAll(resp.Body)
			httpRes := new(utils.Response)
			err = json.Unmarshal(bodyByte, httpRes)
			if err == nil {
				b.Command, _ = httpRes.Data.(string)
			}
		}
	}
	
	commands := adminUtils.ParseCommand(b.Command)
	if len(commands) == 0 || commands["bot_name"] == "" || commands["http_host"] == "" {
		logger.ErrorCtx(ctx, "create bot error", "commands", commands)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, errors.New("command is empty"))
		return
	}
	
	botInfo, err := db.GetBotByID(strconv.Itoa(b.ID))
	if err != nil {
		logger.ErrorCtx(ctx, "update bot address error", "reason", "not found", "id", b.ID, "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	b.Address = utils.NormalizeHTTP(commands["http_host"])
	b.Name = commands["bot_name"]
	err = db.UpdateBotAddress(b.ID, b.Address, b.Name, b.CrtFile, b.KeyFile, b.CaFile, b.Command)
	if err != nil {
		logger.ErrorCtx(ctx, "update bot address error", "reason", "db fail", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	go checkpoint.ScheduleBotChecks()
	if botInfo.Address != b.Address || b.IsStart {
		adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
			strings.TrimSuffix(botInfo.Address, "/")+"/stop", bytes.NewBuffer(nil)))
	}
	
	if b.IsStart {
		err = adminUtils.StartDetachedProcess(b.Command)
		if err != nil {
			logger.ErrorCtx(ctx, "start bot error", "err", err)
			utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	}
	
	utils.Success(ctx, w, r, "bot address updated")
}

func SoftDeleteBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.ErrorCtx(ctx, "soft delete bot error", "reason", "invalid id", "id", idStr)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	err = db.SoftDeleteBot(id)
	if err != nil {
		logger.ErrorCtx(ctx, "soft delete bot error", "reason", "db fail", "id", id, "err", err)
		utils.Failure(ctx, w, r, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	checkpoint.BotMap.Delete(id)
	utils.Success(ctx, w, r, "bot deleted")
}

func ListBots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, pageSize := parsePaginationParams(r)
	
	if *adminConf.RegisterConfInfo.Type != "" {
		bots := make([]*RegisterBot, 0)
		
		var total = 0
		var index = 0
		start := (page - 1) * pageSize
		end := start + pageSize
		
		checkpoint.BotMap.Range(func(key, value any) bool {
			total++
			if index >= start && index < end {
				if bot, ok := value.(*checkpoint.BotStatus); ok {
					bots = append(bots, &RegisterBot{
						ID:         bot.Address,
						Name:       key.(string),
						Address:    bot.Address,
						Status:     bot.Status,
						CreateTime: bot.LastCheck.Unix(),
						UpdateTime: bot.LastCheck.Unix(),
					})
				}
			}
			index++
			return true
		})
		
		utils.Success(ctx, w, r, map[string]interface{}{
			"list":        bots,
			"total":       total,
			"is_register": true,
		})
		return
	}
	
	address := r.URL.Query().Get("address")
	
	offset := (page - 1) * pageSize
	bots, total, err := db.ListBots(offset, pageSize, address)
	if err != nil {
		logger.ErrorCtx(ctx, "list bots error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	for _, bot := range bots {
		statusInter, ok := checkpoint.BotMap.Load(bot.ID)
		if ok {
			status := statusInter.(*checkpoint.BotStatus)
			if status.LastCheck.Add(3 * time.Minute).After(time.Now()) {
				bot.Status = status.Status
			} else {
				bot.Status = checkpoint.OfflineStatus
			}
		}
	}
	
	utils.Success(ctx, w, r, map[string]interface{}{
		"list":        bots,
		"total":       total,
		"is_register": false,
	})
}

func GetBotConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/conf/get", bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	
	bodyByte, err := io.ReadAll(resp.Body)
	httpRes := new(GetBotConfRes)
	err = json.Unmarshal(bodyByte, httpRes)
	if err != nil {
		logger.ErrorCtx(ctx, "json umarshal error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	res := map[string]map[string]any{
		"base":  make(map[string]any),
		"audio": make(map[string]any),
		"llm":   make(map[string]any),
		"photo": make(map[string]any),
		"video": make(map[string]any),
	}
	for k, v := range CompareFlagsWithStructTags(httpRes.Data.Base) {
		res["base"][k] = v
	}
	for k, v := range CompareFlagsWithStructTags(httpRes.Data.Audio) {
		res["audio"][k] = v
	}
	for k, v := range CompareFlagsWithStructTags(httpRes.Data.LLM) {
		res["llm"][k] = v
	}
	for k, v := range CompareFlagsWithStructTags(httpRes.Data.Photo) {
		res["photo"][k] = v
	}
	for k, v := range CompareFlagsWithStructTags(httpRes.Data.Video) {
		res["video"][k] = v
	}
	
	utils.Success(ctx, w, r, res)
}

func AddUserToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+"/user/token/add", r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/user/list?page=%s&page_size=%s&user_id=%s", r.FormValue("page"), r.FormValue("pageSize"), r.FormValue("userId")), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotAdminRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		utils.Failure(ctx, w, r, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(ctx, w, r, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user record error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/record/list?page=%s&page_size=%s&user_id=%d&is_deleted=0&record_type=3",
			r.FormValue("page"), r.FormValue("pageSize"), userIDValue.(int)*-1), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user record error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotUserRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user record error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet, strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/record/list?page=%s&page_size=%s&user_id=%s&record_type=0,1,2,4", r.FormValue("page"),
			r.FormValue("pageSize"), r.FormValue("userId")), bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot user record error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetAllOnlineBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	res := make([]*checkpoint.BotStatus, 0)
	checkpoint.BotMap.Range(func(key any, value any) bool {
		status := value.(*checkpoint.BotStatus)
		if (*adminConf.RegisterConfInfo.Type != "" || status.LastCheck.Add(3*time.Minute).After(time.Now())) &&
			status.Status != checkpoint.OfflineStatus {
			res = append(res, status)
		}
		return true
	})
	
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})
	
	utils.Success(ctx, w, r, res)
}

func UpdateBotConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+"/conf/update", r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotCommand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/command/get", bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	realRsp := new(utils.Response)
	err = json.NewDecoder(resp.Body).Decode(realRsp)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	id := utils.ParseInt(r.URL.Query().Get("id"))
	err = db.UpdateBotCommand(id, realRsp.Data.(string))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	utils.Success(ctx, w, r, realRsp.Data)
	
}

func GetBotMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/get", bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func UpdateBotMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/update?name="+name, r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func DeleteBotMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/delete?name="+name, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "delete bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
}

func DisableBotMCPConf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	disable := r.URL.Query().Get("disable")
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/disable?disable="+disable+"&name="+name, bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "delete bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
}

func GetPrepareMCPServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/get", bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "read response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	getRes := struct {
		Data *mcpParam.McpClientGoConfig `json:"data"`
	}{}
	
	err = json.Unmarshal(byteBody, &getRes)
	if err != nil {
		logger.ErrorCtx(ctx, "unmarshal response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	res := &mcpParam.McpClientGoConfig{
		McpServers: make(map[string]*mcpParam.MCPConfig),
	}
	for name, config := range adminConf.MCPConf.McpServers {
		if _, ok := getRes.Data.McpServers[name]; !ok {
			res.McpServers[name] = config
		}
	}
	
	utils.Success(ctx, w, r, res)
}

func SyncMCPServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+"/mcp/sync", r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func Communicate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		utils.Failure(ctx, w, r, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(ctx, w, r, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	file, _, err := r.FormFile("file")
	var data []byte
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	} else {
		defer file.Close()
		
		data, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
			utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetSSERequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+
			fmt.Sprintf("/communicate?prompt=%s&user_id=-%d",
				url.QueryEscape(r.URL.Query().Get("prompt")), userIDValue), bytes.NewBuffer(data)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	reader := bufio.NewReader(resp.Body)
	
	for {
		line, err := reader.ReadString('\n')
		fmt.Fprint(w, line)
		flusher.Flush()
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading SSE:", err)
			}
			break
		}
	}
}

func Log(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	typ := r.URL.Query().Get("type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetSSERequest(ctx, http.MethodGet,
		strings.TrimSuffix(botInfo.Address, "/")+"/log?type="+typ,
		bytes.NewBuffer(nil)))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	reader := bufio.NewReader(resp.Body)
	
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// 每条日志前加 data:，后面要两个换行
			fmt.Fprintf(w, "data: %s\n\n", strings.TrimRight(line, "\n"))
			flusher.Flush()
		}
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading SSE:", err)
			}
			break
		}
	}
}

func getBot(r *http.Request) (*db.Bot, error) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		logger.ErrorCtx(ctx, "get bot error", "id", idStr)
		return nil, param.ErrParamError
	}
	
	if *adminConf.RegisterConfInfo.Type != "" {
		return &db.Bot{
			Address: idStr,
		}, nil
	}
	
	bot, err := db.GetBotByID(idStr)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot error", "id", idStr, "err", err)
		return nil, param.ErrDBQueryFail
	}
	
	return bot, nil
}

func CompareFlagsWithStructTags(cfg interface{}) map[string]any {
	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)
	
	// If it's a pointer, get the element it points to
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			logger.Warn("Input is a nil pointer")
			return nil
		}
		v = v.Elem()
		t = t.Elem()
	}
	
	if t.Kind() != reflect.Struct {
		logger.Warn("Input must be a struct or pointer to struct")
		return nil
	}
	
	res := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || SkipKey[jsonTag] || jsonTag == "-" {
			continue
		}
		
		structValue := ""
		switch jsonTag {
		case "allowed_user_ids", "allowed_group_ids", "admin_user_ids":
			structValue = utils.MapKeysToString(v.Field(i).Interface())
		default:
			structValue = utils.ValueToString(v.Field(i).Interface())
		}
		
		res[jsonTag] = structValue
	}
	
	return res
}

func InsertUserRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	botInfo, err := getBot(r)
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(GetRequest(ctx, http.MethodPost,
		strings.TrimSuffix(botInfo.Address, "/")+"/user/insert/record", r.Body))
	if err != nil {
		logger.ErrorCtx(ctx, "get bot conf error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "copy response body error", "err", err)
		utils.Failure(ctx, w, r, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetRequest(ctx context.Context, method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("LogId", ctx.Value("log_id").(string))
	return req
}

func GetSSERequest(ctx context.Context, method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "text/event-stream")
	req.Header.Set("LogId", ctx.Value("log_id").(string))
	req.Header.Set("Accept", "text/event-stream")
	return req
}
