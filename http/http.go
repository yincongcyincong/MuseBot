package http

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type HTTPServer struct {
	Addr string
}

func InitHTTP() {
	pprofServer := NewHTTPServer(fmt.Sprintf(":%d", *conf.BaseConfInfo.HTTPPort))
	pprofServer.Start()
}

// NewHTTPServer create http server, listen 36060 port.
func NewHTTPServer(addr string) *HTTPServer {
	if addr == "" {
		addr = ":36060"
	}
	return &HTTPServer{
		Addr: addr,
	}
}

// Start pprof server
func (p *HTTPServer) Start() {
	go func() {
		logger.Info("Starting pprof server on", "addr", p.Addr)
		http.Handle("/metrics", promhttp.Handler())
		
		http.HandleFunc("/user/token/add", AddUserToken)
		
		http.HandleFunc("/conf/update", UpdateConf)
		http.HandleFunc("/conf/get", GetConf)
		
		var err error
		if *conf.BaseConfInfo.CrtFile == "" || *conf.BaseConfInfo.KeyFile == "" {
			err = http.ListenAndServe(p.Addr, nil)
		} else {
			err = http.ListenAndServeTLS(p.Addr, *conf.BaseConfInfo.CrtFile, *conf.BaseConfInfo.KeyFile, nil)
		}
		if err != nil {
			logger.Fatal("pprof server failed", "err", err)
		}
	}()
}
