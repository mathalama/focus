package httpapi

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsOnce sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	rateLimitBlockedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_rate_limit_blocked_total",
			Help: "Total number of requests blocked by rate limiting.",
		},
		[]string{"route"},
	)
)

func registerMetrics() {
	metricsOnce.Do(func() {
		prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, rateLimitBlockedTotal)
	})
}

// MetricsRoute returns a Gin handler for /metrics.
func MetricsRoute() gin.HandlerFunc {
	registerMetrics()
	return gin.WrapH(promhttp.Handler())
}

// MetricsMiddleware records request totals and latency.
func MetricsMiddleware() gin.HandlerFunc {
	registerMetrics()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

func markRateLimitBlocked(route string) {
	registerMetrics()
	if route == "" {
		route = "unknown"
	}
	rateLimitBlockedTotal.WithLabelValues(route).Inc()
}
