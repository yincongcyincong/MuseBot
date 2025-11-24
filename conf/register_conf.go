package conf

import (
	"flag"
	"os"
	"strings"
)

type RegisterConf struct {
	Type         *string  `json:"type"`
	EtcdURLs     []string `json:"etcd_url"`
	EtcdUsername *string  `json:"etcd_username"`
	EtcdPassword *string  `json:"etcd_password"`
	
	etcdURLs *string
}

var (
	RegisterConfInfo = new(RegisterConf)
)

func InitRegisterConf() {
	RegisterConfInfo.Type = flag.String("register_type", "", "register type: etcd")
	RegisterConfInfo.etcdURLs = flag.String("etcd_urls", "http://127.0.0.1:2379", "etcd urls")
	RegisterConfInfo.EtcdUsername = flag.String("etcd_username", "", "etcd username")
	RegisterConfInfo.EtcdPassword = flag.String("etcd_password", "", "etcd password")
}

func EnvRegisterConf() {
	if os.Getenv("REGISTER_TYPE") != "" {
		*RegisterConfInfo.Type = os.Getenv("REGISTER_TYPE")
	}
	
	if os.Getenv("ETCD_URLS") != "" {
		RegisterConfInfo.EtcdURLs = strings.Split(os.Getenv("ETCD_URLS"), ",")
	}
	
	if os.Getenv("ETCD_USERNAME") != "" {
		*RegisterConfInfo.EtcdUsername = os.Getenv("ETCD_USERNAME")
	}
	
	if os.Getenv("ETCD_PASSWORD") != "" {
		*RegisterConfInfo.EtcdPassword = os.Getenv("ETCD_PASSWORD")
	}
	
	if *RegisterConfInfo.Type == "etcd" && RegisterConfInfo.etcdURLs != nil {
		RegisterConfInfo.EtcdURLs = strings.Split(*RegisterConfInfo.etcdURLs, ",")
	}
}
