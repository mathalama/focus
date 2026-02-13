package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter applies per-IP token-bucket limits.
type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	r        rate.Limit
	burst    int
	ttl      time.Duration
}

func NewIPRateLimiter(r rate.Limit, burst int, ttl time.Duration) *IPRateLimiter {
	if burst <= 0 {
		burst = 1
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	return &IPRateLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		burst:    burst,
		ttl:      ttl,
	}
}

func (rl *IPRateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.ttl {
			delete(rl.visitors, ip)
		}
	}

	v, ok := rl.visitors[key]
	if !ok {
		limiter := rate.NewLimiter(rl.r, rl.burst)
		rl.visitors[key] = &visitor{limiter: limiter, lastSeen: now}
		return limiter
	}

	v.lastSeen = now
	return v.limiter
}

func (rl *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow OPTIONS requests to pass through without rate limiting
		// to avoid breaking CORS preflight.
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		key := c.ClientIP() + ":" + c.FullPath()
		if !rl.getLimiter(key).Allow() {
			markRateLimitBlocked(c.FullPath())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
