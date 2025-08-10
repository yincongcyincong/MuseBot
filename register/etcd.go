package register

import (
	"context"
	"strconv"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitRegister() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.RegisterConfInfo.EtcdURLs,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Error("register init failed: ", err)
		return
	}
	defer cli.Close()
	
	leaseResp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		logger.Error("register init failed: ", err)
		return
	}
	
	serviceKey := "/services/musebot/" + strconv.FormatInt(conf.BaseConfInfo.StartTime, 10) + utils.MD5(*conf.BaseConfInfo.HTTPHost)
	_, err = cli.Put(context.Background(), serviceKey, *conf.BaseConfInfo.HTTPHost, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		logger.Error("register init failed: ", err)
		return
	}
	
	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		logger.Error("register init failed: ", err)
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
