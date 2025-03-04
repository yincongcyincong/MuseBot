package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// 定义所有监控指标
var (
	TotalUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_total_users",
			Help: "Total number of unique users.",
		},
	)

	TotalRecords = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_total_records",
			Help: "Total number of records.",
		},
	)

	ConversationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_conversation_duration_seconds",
			Help:    "Duration of conversations in seconds.",
			Buckets: prometheus.DefBuckets, // 默认桶: 0.005, 0.01, 0.025, 0.05, ..., 10, 30, 60
		},
	)

	ImageDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "generate_image_duration_seconds",
			Help:    "generate image API requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

// RegisterMetrics register metrics
func RegisterMetrics() {
	prometheus.MustRegister(TotalUsers)
	prometheus.MustRegister(TotalRecords)
	prometheus.MustRegister(ConversationDuration)
	prometheus.MustRegister(ImageDuration)
}
