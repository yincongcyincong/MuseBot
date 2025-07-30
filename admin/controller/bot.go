package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	
	mcpParam "github.com/yincongcyincong/mcp-client-go/clients/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/checkpoint"
	adminConf "github.com/yincongcyincong/telegram-deepseek-bot/admin/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	adminUtils "github.com/yincongcyincong/telegram-deepseek-bot/admin/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Bot struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	CrtFile string `json:"crt_file"`
	KeyFile string `json:"key_file"`
	CaFile  string `json:"ca_file"`
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

func CreateBot(w http.ResponseWriter, r *http.Request) {
	var b Bot
	err := utils.HandleJsonBody(r, &b)
	if err != nil {
		logger.Error("update bot error", "bot", b)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	if b.Address == "" {
		logger.Error("create bot error", "reason", "empty address or crt_file")
		utils.Failure(w, param.CodeParamError, param.MsgParamError, nil)
		return
	}
	
	err = db.CreateBot(b.Address, b.CrtFile, b.KeyFile, b.CaFile)
	if err != nil {
		logger.Error("create bot error", "reason", "db fail", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	go checkpoint.ScheduleBotChecks()
	
	utils.Success(w, "bot created")
}

func GetBot(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Error("get bot error", "id", idStr)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	bot, err := db.GetBotByID(id)
	if err != nil {
		logger.Error("get bot error", "reason", "not found", "id", id, "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	utils.Success(w, bot)
}

func UpdateBotAddress(w http.ResponseWriter, r *http.Request) {
	var b Bot
	err := utils.HandleJsonBody(r, &b)
	if err != nil {
		logger.Error("update bot error", "bot", b)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	if b.ID <= 0 || b.Address == "" {
		logger.Error("update bot address error", "reason", "invalid id or address", "id", b.ID)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, nil)
		return
	}
	
	err = db.UpdateBotAddress(b.ID, b.Address, b.CrtFile, b.KeyFile, b.CaFile)
	if err != nil {
		logger.Error("update bot address error", "reason", "db fail", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	go checkpoint.ScheduleBotChecks()
	
	utils.Success(w, "bot address updated")
}

func SoftDeleteBot(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Error("soft delete bot error", "reason", "invalid id", "id", idStr)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	err = db.SoftDeleteBot(id)
	if err != nil {
		logger.Error("soft delete bot error", "reason", "db fail", "id", id, "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
	go checkpoint.ScheduleBotChecks()
	
	utils.Success(w, "bot deleted")
}

func ListBots(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePaginationParams(r)
	
	address := r.URL.Query().Get("address")
	
	offset := (page - 1) * pageSize
	bots, total, err := db.ListBots(offset, pageSize, address)
	if err != nil {
		logger.Error("list bots error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
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
	
	utils.Success(w, map[string]interface{}{
		"list":  bots,
		"total": total,
	})
}

func GetBotConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/conf/get")
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	
	bodyByte, err := io.ReadAll(resp.Body)
	httpRes := new(GetBotConfRes)
	err = json.Unmarshal(bodyByte, httpRes)
	if err != nil {
		logger.Error("json umarshal error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
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
	
	utils.Success(w, res)
}

func AddUserToken(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	req, err := http.NewRequest("POST", strings.TrimSuffix(botInfo.Address, "/")+"/user/token/add", bytes.NewBuffer(body))
	if err != nil {
		logger.Error("Error creating request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(req)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotUser(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot user error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		logger.Error("parse form error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/user/list?page=%s&page_size=%s&user_id=%s", r.FormValue("page"), r.FormValue("pageSize"), r.FormValue("userId")))
	if err != nil {
		logger.Error("get bot user error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotAdminRecord(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot user record error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/record/list?page=%s&page_size=%s&user_id=%d&is_deleted=0",
			r.FormValue("page"), r.FormValue("pageSize"), userIDValue.(int)*-1))
	if err != nil {
		logger.Error("get bot user record error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotUserRecord(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot user record error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		logger.Error("parse form error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") +
		fmt.Sprintf("/record/list?page=%s&page_size=%s&user_id=%s", r.FormValue("page"), r.FormValue("pageSize"), r.FormValue("userId")))
	if err != nil {
		logger.Error("get bot user record error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetAllOnlineBot(w http.ResponseWriter, r *http.Request) {
	res := make([]*checkpoint.BotStatus, 0)
	checkpoint.BotMap.Range(func(key any, value any) bool {
		status := value.(*checkpoint.BotStatus)
		if status.LastCheck.Add(3*time.Minute).After(time.Now()) && status.Status != checkpoint.OfflineStatus {
			res = append(res, status)
		}
		return true
	})
	
	utils.Success(w, res)
}

func UpdateBotConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	req, err := http.NewRequest("POST", strings.TrimSuffix(botInfo.Address, "/")+"/conf/update", bytes.NewBuffer(body))
	if err != nil {
		logger.Error("Error creating request", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(req)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotCommand(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/command/get")
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func GetBotMCPConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/mcp/get")
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func UpdateBotMCPConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	req, err := http.NewRequest("POST", strings.TrimSuffix(botInfo.Address, "/")+"/mcp/update?name="+name, bytes.NewBuffer(body))
	if err != nil {
		logger.Error("Error creating request", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(req)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func DeleteBotMCPConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/mcp/delete?name=" + name)
	if err != nil {
		logger.Error("delete bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
}

func DisableBotMCPConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	name := r.URL.Query().Get("name")
	disable := r.URL.Query().Get("disable")
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/mcp/disable?disable=" + disable + "&name=" + name)
	if err != nil {
		logger.Error("delete bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
}

func GetPrepareMCPServer(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/mcp/get")
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	
	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("read response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	getRes := struct {
		Data *mcpParam.McpClientGoConfig `json:"data"`
	}{}
	
	err = json.Unmarshal(byteBody, &getRes)
	if err != nil {
		logger.Error("unmarshal response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
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
	
	utils.Success(w, res)
}

func SyncMCPServer(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	resp, err := adminUtils.GetCrtClient(botInfo).Get(strings.TrimSuffix(botInfo.Address, "/") + "/mcp/sync")
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("copy response body error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
}

func Communicate(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
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
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	userIDValue, ok := session.Values["user_id"]
	if !ok || userIDValue == nil {
		utils.Failure(w, param.CodeNotLogin, param.MsgNotLogin, nil)
		return
	}
	
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.Error("parse form error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	
	file, _, err := r.FormFile("file")
	var data []byte
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	} else {
		defer file.Close()
		
		data, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
			utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
			return
		}
	}
	
	req, err := http.NewRequest("POST", strings.TrimSuffix(botInfo.Address, "/")+
		fmt.Sprintf("/communicate?prompt=%s&userId=%d",
			url.QueryEscape(r.URL.Query().Get("prompt")), userIDValue), bytes.NewBuffer(data))
	if err != nil {
		logger.Error("Error creating request", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := adminUtils.GetCrtClient(botInfo).Do(req)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeServerFail, param.MsgServerFail, err)
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

func getBot(r *http.Request) (*db.Bot, error) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Error("get bot error", "id", idStr)
		return nil, param.ErrParamError
	}
	
	bot, err := db.GetBotByID(id)
	if err != nil {
		logger.Error("get bot error", "id", id, "err", err)
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
		if jsonTag == "" || SkipKey[jsonTag] {
			continue
		}
		
		structValue := ""
		switch jsonTag {
		case "allowed_telegram_user_ids", "allowed_telegram_group_ids", "admin_user_ids":
			structValue = utils.MapKeysToString(v.Field(i).Interface())
		default:
			structValue = utils.ValueToString(v.Field(i).Interface())
		}
		
		res[jsonTag] = structValue
	}
	
	return res
}
