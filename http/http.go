package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
	
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
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
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		
		mux.HandleFunc("/user/token/add", AddUserToken)
		
		mux.HandleFunc("/conf/update", UpdateConf)
		mux.HandleFunc("/conf/get", GetConf)
		mux.HandleFunc("/command/get", GetCommand)
		mux.HandleFunc("/restart", Restart)
		mux.HandleFunc("/stop", Stop)
		mux.HandleFunc("/log", Log)
		
		mux.HandleFunc("/mcp/get", GetMCPConf)
		mux.HandleFunc("/mcp/update", UpdateMCPConf)
		mux.HandleFunc("/mcp/disable", DisableMCPConf)
		mux.HandleFunc("/mcp/delete", DeleteMCPConf)
		mux.HandleFunc("/mcp/sync", SyncMCPConf)
		
		mux.HandleFunc("/user/list", GetUsers)
		mux.HandleFunc("/user/update/mode", UpdateMode)
		mux.HandleFunc("/user/insert/record", InsertUserRecords)
		mux.HandleFunc("/record/list", GetRecords)
		
		mux.HandleFunc("/rag/list", GetRagFile)
		mux.HandleFunc("/rag/delete", DeleteRagFile)
		mux.HandleFunc("/rag/create", CreateRagFile)
		mux.HandleFunc("/rag/get", GetRagFileContent)
		mux.HandleFunc("/rag/clear", ClearAllVectorData)
		
		mux.HandleFunc("/pong", PongHandler)
		mux.HandleFunc("/dashboard", DashboardHandler)
		
		mux.HandleFunc("/communicate", Communicate)
		mux.HandleFunc("/com/wechat", ComWechatComm)
		mux.HandleFunc("/wechat", WechatComm)
		mux.HandleFunc("/qq", QQBotComm)
		
		wrappedMux := WithRequestContext(mux)
		
		var err error
		if conf.BaseConfInfo.CrtFile == nil || conf.BaseConfInfo.KeyFile == nil ||
			*conf.BaseConfInfo.CrtFile == "" || *conf.BaseConfInfo.KeyFile == "" {
			err = http.ListenAndServe(p.Addr, wrappedMux)
		} else {
			err = runTLSServer(wrappedMux)
		}
		if err != nil {
			logger.Fatal("pprof server failed", "err", err)
		}
	}()
}

func runTLSServer(wrappedMux http.Handler) error {
	caCert, err := os.ReadFile(*conf.BaseConfInfo.CaFile)
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
		Handler:   wrappedMux,
	}
	
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		return err
	}
	
	return nil
}

func WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		isSSE := r.Header.Get("Accept") == "text/event-stream"
		
		var cancel context.CancelFunc
		if !isSSE {
			ctx, cancel = context.WithTimeout(ctx, 15*time.Minute)
			defer cancel()
		}
		
		logID := r.Header.Get("LogId")
		if logID == "" {
			logID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, "log_id", logID)
		
		if conf.BaseConfInfo.BotName != nil {
			ctx = context.WithValue(ctx, "bot_name", *conf.BaseConfInfo.BotName)
		}
		
		ctx = context.WithValue(ctx, "start_time", time.Now())
		
		r = r.WithContext(ctx)
		
		logger.InfoCtx(ctx, "request start", "path", r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		metrics.HTTPRequestCount.WithLabelValues(r.URL.Path).Inc()
	})
}
