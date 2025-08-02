package checkpoint

import (
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/utils"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

const (
	OnlineStatus  = "online"
	OfflineStatus = "offline"
)

var BotMap sync.Map

type BotStatus struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	LastCheck time.Time `json:"-"`
}

func InitStatusCheck() {
	go func() {
		ScheduleBotChecks()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C: // 每 60 秒触发一次
				go ScheduleBotChecks()
			}
		}
	}()
}

// 健康检查
func checkBotStatus(bot *db.Bot) string {
	// 发送请求
	resp, err := utils.GetCrtClient(bot).Get(strings.TrimSuffix(bot.Address, "/") + "/pong")
	if err != nil {
		logger.Warn("checkpoint request fail", "err", err, "address", bot.Address)
		return "offline" // 请求失败
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		logger.Warn("checkpoint request fail", "resp", resp, "address", bot.Address)
		return OfflineStatus
	}
	
	return OnlineStatus
}

// ScheduleBotChecks 分批调度，每批 10 秒执行一次
func ScheduleBotChecks() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("ScheduleBotChecks panic", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	bots, _, err := db.ListBots(0, 10000, "")
	if err != nil {
		logger.Error("ScheduleBotChecks list bots fail", "err", err)
		return
	}
	
	batchCount := 60                                       // 60 秒内分 60 批执行
	batchSize := (len(bots) + batchCount - 1) / batchCount // 向上取整
	
	var wg sync.WaitGroup
	for i := 0; i < batchCount; i++ {
		start := i * batchSize
		end := (i + 1) * batchSize
		if end > len(bots) {
			end = len(bots)
		}
		if start >= len(bots) {
			break
		}
		
		// 复制下标范围，防止闭包变量覆盖
		batch := bots[start:end]
		batchIndex := i
		
		wg.Add(1)
		go func(batch []*db.Bot, batchIndex int) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("ScheduleBotChecks panic", "err", err, "stack", string(debug.Stack()))
				}
				wg.Done()
			}()
			// 延迟 10 * batchIndex 秒启动
			timer := time.NewTimer(time.Duration(batchIndex) * time.Second)
			<-timer.C
			
			for _, b := range batch {
				status := checkBotStatus(b)
				BotMap.Store(b.ID, &BotStatus{
					Id:        b.ID,
					Name:      b.Name,
					Address:   b.Address,
					Status:    status,
					LastCheck: time.Now(),
				})
			}
		}(batch, batchIndex)
	}
	
	wg.Wait()
}
