package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
)

type HTTPServer struct {
	Addr string
}

func InitHTTP() {
	pprofServer := NewHTTPServer(fmt.Sprintf("%s", *conf.BaseConfInfo.HTTPHost))
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
		http.HandleFunc("/command/get", GetCommand)
		http.HandleFunc("/mcp/get", GetMCPConf)
		http.HandleFunc("/mcp/update", UpdateMCPConf)
		http.HandleFunc("/mcp/disable", DisableMCPConf)
		http.HandleFunc("/mcp/delete", DeleteMCPConf)
		http.HandleFunc("/mcp/sync", SyncMCPConf)
		
		http.HandleFunc("/user/list", GetUsers)
		http.HandleFunc("/user/update/mode", UpdateMode)
		http.HandleFunc("/record/list", GetRecords)
		
		http.HandleFunc("/pong", PongHandler)
		http.HandleFunc("/dashboard", DashboardHandler)
		
		http.HandleFunc("/communicate", Communicate)
		http.HandleFunc("/com/wechat", ComWechatComm)
		http.HandleFunc("/qq", QQBotComm)
		
		var err error
		if conf.BaseConfInfo.CrtFile == nil || conf.BaseConfInfo.KeyFile == nil ||
			*conf.BaseConfInfo.CrtFile == "" || *conf.BaseConfInfo.KeyFile == "" {
			err = http.ListenAndServe(p.Addr, nil)
		} else {
			err = runTLSServer()
		}
		if err != nil {
			logger.Fatal("pprof server failed", "err", err)
		}
	}()
}

func runTLSServer() error {
	caCert, err := ioutil.ReadFile(*conf.BaseConfInfo.CaFile)
	if err != nil {
		return err
	}
	
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	cert, err := tls.LoadX509KeyPair(*conf.BaseConfInfo.CrtFile, *conf.BaseConfInfo.KeyFile)
	if err != nil {
		return err
	}
	
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}
	
	server := &http.Server{
		Addr:      fmt.Sprintf("%s", *conf.BaseConfInfo.HTTPHost),
		TLSConfig: tlsConfig,
	}
	
	err = server.ListenAndServeTLS("", "") // cert/key 已通过 TLSConfig 提供
	if err != nil {
		return err
	}
	
	return nil
}
