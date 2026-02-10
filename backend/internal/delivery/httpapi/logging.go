package httpapi

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type accessLog struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id,omitempty"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Route     string `json:"route"`
	Status    int    `json:"status"`
	Duration  string `json:"duration"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Error     string `json:"error,omitempty"`
}

// StructuredLogger emits one JSON log line per HTTP request.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		entry := accessLog{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Level:     "info",
			Msg:       "http_request",
			RequestID: getRequestID(c),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Route:     c.FullPath(),
			Status:    c.Writer.Status(),
			Duration:  time.Since(startedAt).String(),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		payload, err := json.Marshal(entry)
		if err != nil {
			log.Printf(`{"ts":"%s","level":"error","msg":"marshal_access_log_failed","error":"%v"}`, time.Now().UTC().Format(time.RFC3339Nano), err)
			return
		}
		log.Print(string(payload))
	}
}
