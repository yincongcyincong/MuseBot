package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/admin/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

func GetCrtClient(bot *db.Bot) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
	
	if bot.KeyFile != "" && bot.CrtFile != "" && bot.CaFile != "" {
		clientCert, err := tls.X509KeyPair([]byte(bot.CrtFile), []byte(bot.KeyFile))
		if err != nil {
			logger.Error("Failed to load client cert/key", "err", err)
			return client
		}
		
		// Load CA cert from memory into cert pool
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM([]byte(bot.CaFile)); !ok {
			logger.Error("Failed to append CA certificate to pool")
			return client
		}
		
		// TLS config with mTLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      caCertPool,
		}
		transport.TLSClientConfig = tlsConfig
	}
	
	return client
}
