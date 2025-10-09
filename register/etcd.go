package register

import (
	"context"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitRegister() {
	if *conf.RegisterConfInfo.Type == "" {
		return
	}

	switch *conf.RegisterConfInfo.Type {
	case "etcd":
		StartEtcdRegister()
	}
}

func StartEtcdRegister() {
	if len(conf.RegisterConfInfo.EtcdURLs) == 0 {
		return
	}

	go func() {
		for {
			cli, err := clientv3.New(clientv3.Config{
				Endpoints:   conf.RegisterConfInfo.EtcdURLs,
				DialTimeout: 5 * time.Second,
				Username:    *conf.RegisterConfInfo.EtcdUsername,
				Password:    *conf.RegisterConfInfo.EtcdPassword,
			})
			if err != nil {
				logger.Error("register connect failed:", err)
				time.Sleep(3 * time.Second)
				continue
			}

			serviceKey := "/services/musebot/" +
				utils.MD5(*conf.BaseConfInfo.HTTPHost) +
				"/" + *conf.BaseConfInfo.BotName

			leaseResp, err := cli.Grant(context.Background(), 5)
			if err != nil {
				logger.Error("lease grant failed:", err)
				cli.Close()
				time.Sleep(3 * time.Second)
				continue
			}

			_, err = cli.Put(context.Background(), serviceKey, *conf.BaseConfInfo.HTTPHost, clientv3.WithLease(leaseResp.ID))
			if err != nil {
				logger.Error("register put fail:", err)
				cli.Close()
				time.Sleep(3 * time.Second)
				continue
			}

			ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
			if err != nil {
				logger.Error("keepalive start failed:", err)
				cli.Close()
				time.Sleep(3 * time.Second)
				continue
			}

			logger.Info("service registered", "serviceKey", serviceKey)

			keepAliveOK := true
			for ka := range ch {
				if ka == nil {
					logger.Error("keepalive channel closed, will retry register...")
					keepAliveOK = false
					break
				}
			}

			cli.Close()

			// 如果 keepalive 挂了，3 秒后重试注册
			if !keepAliveOK {
				time.Sleep(3 * time.Second)
			}
		}
	}()
}
