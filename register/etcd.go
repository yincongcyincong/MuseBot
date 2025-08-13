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
	
	leaseResp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		logger.Error("register init failed: ", err)
		return
	}
	
	serviceKey := "/services/musebot/" + utils.MD5(*conf.BaseConfInfo.HTTPHost) + "/" + *conf.BaseConfInfo.BotName
	_, err = cli.Put(context.Background(), serviceKey, *conf.BaseConfInfo.HTTPHost, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		logger.Error("register put fail: ", err)
		return
	}
	
	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		logger.Error("register keep alive failed: ", err)
		return
	}
	
	go func() {
		for {
			ka := <-ch
			if ka == nil {
				logger.Error("lease keepalive failed")
				return
			}
		}
	}()
}
