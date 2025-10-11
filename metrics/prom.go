package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	APIRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_request_total",
			Help: "Total number of API requests, labeled by model name.",
		},
		[]string{"model"},
	)
	
	APIRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Histogram of API request durations in seconds, labeled by model name.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model"},
	)
	
	AppRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_request_total",
			Help: "Total number of API requests per app.",
		},
		[]string{"app"},
	)
)

// RegisterMetrics 注册指标
func RegisterMetrics() {
	prometheus.MustRegister(APIRequestCount)
	prometheus.MustRegister(APIRequestDuration)
	prometheus.MustRegister(AppRequestCount)
}
