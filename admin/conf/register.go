package conf

import (
	"flag"
	"os"
	"strings"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type RegisterConf struct {
	Type         string   `json:"type"`
	EtcdURLs     []string `json:"etcd_url"`
	EtcdUsername string   `json:"etcd_username"`
	EtcdPassword string   `json:"etcd_password"`
	
	etcdURLs string
}

var (
	RegisterConfInfo = new(RegisterConf)
)

func InitRegisterConf() {
	flag.StringVar(&RegisterConfInfo.Type, "register_type", "", "register type: etcd")
	flag.StringVar(&RegisterConfInfo.etcdURLs, "etcd_urls", "", "etcd urls")
	flag.StringVar(&RegisterConfInfo.EtcdUsername, "etcd_username", "", "etcd username")
	flag.StringVar(&RegisterConfInfo.EtcdPassword, "etcd_password", "", "etcd password")
	
}

func EnvRegisterConf() {
	if os.Getenv("REGISTER_TYPE") != "" {
		RegisterConfInfo.Type = os.Getenv("REGISTER_TYPE")
	}
	
	if os.Getenv("ETCD_URLS") != "" {
		RegisterConfInfo.EtcdURLs = strings.Split(os.Getenv("ETCD_URLS"), ",")
	}
	
	if os.Getenv("ETCD_USERNAME") != "" {
		RegisterConfInfo.EtcdUsername = os.Getenv("ETCD_USERNAME")
	}
	
	if os.Getenv("ETCD_PASSWORD") != "" {
		RegisterConfInfo.EtcdPassword = os.Getenv("ETCD_PASSWORD")
	}
	
	if RegisterConfInfo.Type == "etcd" && RegisterConfInfo.etcdURLs != "" {
		RegisterConfInfo.EtcdURLs = strings.Split(RegisterConfInfo.etcdURLs, ",")
	}
	
	logger.Info("REGISTER_CONF", "Type", RegisterConfInfo.Type)
	logger.Info("REGISTER_CONF", "EtcdURLs", RegisterConfInfo.EtcdURLs)
	logger.Info("REGISTER_CONF", "EtcdUsername", RegisterConfInfo.EtcdUsername)
	logger.Info("REGISTER_CONF", "EtcdPassword", RegisterConfInfo.EtcdPassword)
	
}
