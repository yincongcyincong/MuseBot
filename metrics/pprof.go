package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"net/http"
	_ "net/http/pprof"
)

type PProfServer struct {
	Addr string
}

func InitPprof() {
	pprofServer := NewPProfServer(":36060")
	pprofServer.Start()
}

// NewPProfServer create http server, listen 36060 port.
func NewPProfServer(addr string) *PProfServer {
	if addr == "" {
		addr = ":36060"
	}
	return &PProfServer{
		Addr: addr,
	}
}

// Start start pprof server
func (p *PProfServer) Start() {
	go func() {
		logger.Info("Starting pprof server on", "addr", p.Addr)
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(p.Addr, nil)
		if err != nil {
			logger.Fatal("pprof server failed", "err", err)
		}
	}()
}
