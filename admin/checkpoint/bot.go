package checkpoint

import (
	"context"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/admin/utils"
	"github.com/yincongcyincong/MuseBot/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	OnlineStatus  = "online"
	OfflineStatus = "offline"
)

var BotMap sync.Map

type BotStatus struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	LastCheck time.Time `json:"-"`
}

func InitStatusCheck() {
	if *conf.RegisterConfInfo.Type != "" {
		go func() {
			InitRegister()
		}()
		
	} else {
		go func() {
			ManualCheckPoint()
		}()
	}
	
}

func ManualCheckPoint() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("InitStatusCheck panic", "err", err, "stack", string(debug.Stack()))
		}
	}()
	ScheduleBotChecks()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C: // 每 60 秒触发一次
			go ScheduleBotChecks()
		}
	}
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
	
	batchCount := 60
	batchSize := (len(bots) + batchCount - 1) / batchCount
	
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
			
			timer := time.NewTimer(time.Duration(batchIndex) * time.Second)
			<-timer.C
			
			for _, b := range batch {
				status := checkBotStatus(b)
				BotMap.Store(b.ID, &BotStatus{
					Id:        strconv.Itoa(b.ID),
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

func InitRegister() {
	switch *conf.RegisterConfInfo.Type {
	case "etcd":
		InitEtcdRegister()
	}
}

func InitEtcdRegister() {
	if len(conf.RegisterConfInfo.EtcdURLs) == 0 {
		return
	}
	
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.RegisterConfInfo.EtcdURLs,
		DialTimeout: 5 * time.Second,
		Username:    *conf.RegisterConfInfo.EtcdUsername,
		Password:    *conf.RegisterConfInfo.EtcdPassword,
	})
	if err != nil {
		logger.Error("register init failed: ", err)
		return
	}
	defer cli.Close()
	
	ctx := context.Background()
	prefix := "/services/musebot/"
	
	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error("register get failed: ", err)
		return
	}
	
	for _, kv := range resp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		name := parts[len(parts)-1]
		
		id := utils.NormalizeAddress(string(kv.Value))
		
		BotMap.Store(name, &BotStatus{
			Id:        id,
			Name:      name,
			Address:   id,
			Status:    OnlineStatus,
			LastCheck: time.Now(),
		})
	}
	
	rch := cli.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	
	for wresp := range rch {
		for _, ev := range wresp.Events {
			key := string(ev.Kv.Key)
			val := string(ev.Kv.Value)
			
			switch ev.Type {
			case clientv3.EventTypePut:
				parts := strings.Split(key, "/")
				name := parts[len(parts)-1]
				
				id := utils.NormalizeAddress(val)
				BotMap.Store(name, &BotStatus{
					Id:        id,
					Name:      name,
					Address:   id,
					Status:    OnlineStatus,
					LastCheck: time.Now(),
				})
			case clientv3.EventTypeDelete:
				BotMap.Delete(key)
			}
		}
	}
}
