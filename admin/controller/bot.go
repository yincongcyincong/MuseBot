package controller

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/checkpoint"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	adminUtils "github.com/yincongcyincong/telegram-deepseek-bot/admin/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Bot struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	CrtFile string `json:"crt_file"`
}

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
	
	err = db.CreateBot(b.Address, b.CrtFile)
	if err != nil {
		logger.Error("create bot error", "reason", "db fail", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
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
	
	err = db.UpdateBotAddress(b.ID, b.Address, b.CrtFile)
	if err != nil {
		logger.Error("update bot address error", "reason", "db fail", "err", err)
		utils.Failure(w, param.CodeDBWriteFail, param.MsgDBWriteFail, err)
		return
	}
	
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
	
	utils.Success(w, "bot deleted")
}

func ListBots(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePaginationParams(r)
	
	offset := (page - 1) * pageSize
	bots, total, err := db.ListBots(offset, pageSize)
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
				bot.Status = "offline"
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
	
	resp, err := adminUtils.GetCrtClient(botInfo.CrtFile).Get(strings.TrimSuffix(botInfo.Address, "/") + "/conf/get")
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

func UpdateBotConf(w http.ResponseWriter, r *http.Request) {
	botInfo, err := getBot(r)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("get bot conf error", "err", err)
		utils.Failure(w, param.CodeParamError, param.MsgParamError, err)
		return
	}
	
	req, err := http.NewRequest("POST", strings.TrimSuffix(botInfo.Address, "/")+"/conf/update", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := adminUtils.GetCrtClient(botInfo.CrtFile).Do(req)
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
	
	resp, err := adminUtils.GetCrtClient(botInfo.CrtFile).Get(strings.TrimSuffix(botInfo.Address, "/") + "/command/get")
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
