package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
)

func GetCrtClient(crtFile string) *http.Client {
	// 创建自定义 Transport，根据是否提供 crtFile 决定是否使用 TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: crtFile == "", // 如果没有证书，跳过验证（仅测试用，生产环境应避免）
		},
	}
	
	// 如果提供了证书文件，则加载证书
	if crtFile != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(crtFile))
		
		transport.TLSClientConfig = &tls.Config{
			RootCAs:            caCertPool, // 使用自定义 CA 证书
			InsecureSkipVerify: false,      // 必须验证证书
		}
	}
	
	// 创建带自定义 Transport 的 HTTP 客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
	
	return client
}
