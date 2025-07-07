package checkpoint

import (
	"net/http"
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
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	LastCheck time.Time `json:"-"`
}

func InitStatusCheck() {
	go func() {
		scheduleBotChecks()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C: // 每 60 秒触发一次
				scheduleBotChecks()
			}
		}
	}()
}

// 健康检查
func checkBotStatus(address string, crtFile string) string {
	// 发送请求
	resp, err := utils.GetCrtClient(crtFile).Get(strings.TrimSuffix(address, "/") + "/pong")
	if err != nil {
		logger.Warn("checkpoint request fail", "err", err, "address", address)
		return "offline" // 请求失败
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		logger.Warn("checkpoint request fail", "resp", resp, "address", address)
		return OfflineStatus
	}
	
	return OnlineStatus
}

// scheduleBotChecks 分批调度，每批 10 秒执行一次
func scheduleBotChecks() {
	bots, _, err := db.ListBots(0, 10000, "")
	if err != nil {
		panic(err)
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
			defer wg.Done()
			// 延迟 10 * batchIndex 秒启动
			timer := time.NewTimer(time.Duration(batchIndex) * time.Second)
			<-timer.C
			
			for _, b := range batch {
				status := checkBotStatus(b.Address, b.CrtFile)
				BotMap.Store(b.ID, &BotStatus{
					Id:        b.ID,
					Address:   b.Address,
					Status:    status,
					LastCheck: time.Now(),
				})
			}
		}(batch, batchIndex)
	}
	
	wg.Wait()
}
