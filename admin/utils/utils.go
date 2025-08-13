package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/admin/db"
	"github.com/yincongcyincong/MuseBot/logger"
)

func GetCrtClient(bot *db.Bot) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   180 * time.Second,
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

func NormalizeAddress(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}
