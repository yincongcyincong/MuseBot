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
	
	HTTPRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path"},
	)
	
	HTTPResponseCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_total",
			Help: "Total number of HTTP responses.",
		},
		[]string{"path", "code"},
	)
	
	HTTPResponseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_duration_seconds",
			Help:    "Histogram of HTTP response durations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "code"},
	)
	
	MCPRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_request_total",
			Help: "Total number of MCP requests.",
		},
		[]string{"mcp_service", "mcp_func"},
	)
	
	MCPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_duration_seconds",
			Help:    "Histogram of MCP request durations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"mcp_service", "mcp_func"},
	)
)

// RegisterMetrics 注册指标
func RegisterMetrics() {
	prometheus.MustRegister(APIRequestCount)
	prometheus.MustRegister(APIRequestDuration)
	prometheus.MustRegister(AppRequestCount)
	prometheus.MustRegister(HTTPRequestCount)
	prometheus.MustRegister(MCPRequestCount)
	prometheus.MustRegister(HTTPResponseCount)
	prometheus.MustRegister(HTTPResponseDuration)
	prometheus.MustRegister(MCPRequestDuration)
}
