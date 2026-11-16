package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_http_requests_total", Help: "Completed HTTP requests by matched route and HTTP status.",
	}, []string{"route", "status"})
	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_http_request_duration_seconds", Help: "Time spent serving HTTP requests, including open streams.",
		Buckets: []float64{.005, .025, .1, .5, 1, 5, 15, 30, 60, 180, 900},
	}, []string{"route", "status"})
	HTTPInFlight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "musebot_http_requests_in_flight", Help: "HTTP requests currently being served.",
	}, []string{"route"})
	LLMAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_llm_request_attempts_total", Help: "Chat provider API attempts, including retries and stream setup.",
	}, []string{"provider", "model", "mode", "outcome"})
	LLMAttemptDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_llm_request_attempt_duration_seconds", Help: "Chat API attempt latency, through stream setup for streaming requests.",
		Buckets: []float64{.05, .1, .25, .5, 1, 2, 5, 10, 30, 60, 180},
	}, []string{"provider", "model", "mode", "outcome"})
	LLMStreamDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_llm_stream_duration_seconds", Help: "Time consuming an open chat stream, including time writing chunks.",
		Buckets: []float64{.1, .25, .5, 1, 2, 5, 10, 30, 60, 180, 600},
	}, []string{"provider", "model", "outcome"})
	LLMFirstContent = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_llm_first_content_seconds", Help: "Time from starting a streaming chat round to the first nonempty content chunk.",
		Buckets: []float64{.05, .1, .25, .5, 1, 2, 5, 10, 30, 60},
	}, []string{"provider", "model"})
	ToolCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_tool_calls_total", Help: "Logical tool calls, including built-in tools and MCP calls.",
	}, []string{"service", "tool", "outcome"})
	ToolDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_tool_call_duration_seconds", Help: "Tool call latency including retries.",
		Buckets: []float64{.01, .05, .1, .25, .5, 1, 2, 5, 10, 30, 60, 180},
	}, []string{"service", "tool", "outcome"})
	TaskRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_task_runs_total", Help: "Task runs by terminal state.",
	}, []string{"state"})
	TaskDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "musebot_task_run_duration_seconds", Help: "Time spent executing task runs.",
		Buckets: []float64{.5, 1, 2, 5, 10, 30, 60, 120, 300, 600, 900},
	}, []string{"state"})
	TaskRecovery = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_task_recovery_total", Help: "Tasks by their last step recovery outcome.",
	}, []string{"outcome"})
	FailoverResults = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "musebot_llm_failover_results_total", Help: "Completed fallback attempts by source, destination, and outcome.",
	}, []string{"from", "to", "outcome"})
)

func Outcome(ctx context.Context, err error) string {
	if err == nil {
		return "success"
	}
	var ctxErr error
	if ctx != nil {
		ctxErr = ctx.Err()
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded) || ctxErr == context.DeadlineExceeded:
		return "timeout"
	case errors.Is(err, context.Canceled) || ctxErr == context.Canceled:
		return "canceled"
	default:
		return "error"
	}
}

func ObserveTool(ctx context.Context, service, tool string, start time.Time, err error) {
	outcome := Outcome(ctx, err)
	ToolCalls.WithLabelValues(service, tool, outcome).Inc()
	ToolDuration.WithLabelValues(service, tool, outcome).Observe(time.Since(start).Seconds())
}
