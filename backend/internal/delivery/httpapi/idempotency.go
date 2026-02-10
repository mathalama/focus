package httpapi

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type idempotencyEntry struct {
	createdAt time.Time
}

// IdempotencyStore protects critical write routes from duplicate submissions.
type IdempotencyStore struct {
	mu      sync.Mutex
	entries map[string]idempotencyEntry
	ttl     time.Duration
}

func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &IdempotencyStore{
		entries: make(map[string]idempotencyEntry),
		ttl:     ttl,
	}
}

func (s *IdempotencyStore) cleanup(now time.Time) {
	for key, entry := range s.entries {
		if now.Sub(entry.createdAt) > s.ttl {
			delete(s.entries, key)
		}
	}
}

func (s *IdempotencyStore) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if rawKey == "" {
			respondError(c, http.StatusBadRequest, "Idempotency-Key header is required")
			c.Abort()
			return
		}

		scope := strings.TrimSpace(getUserID(c))
		if scope == "" {
			scope = strings.TrimSpace(c.ClientIP())
		}

		storeKey := c.Request.Method + ":" + c.FullPath() + ":" + scope + ":" + rawKey
		now := time.Now()

		s.mu.Lock()
		s.cleanup(now)
		if _, exists := s.entries[storeKey]; exists {
			s.mu.Unlock()
			respondError(c, http.StatusConflict, "duplicate request")
			c.Abort()
			return
		}
		s.entries[storeKey] = idempotencyEntry{createdAt: now}
		s.mu.Unlock()

		c.Next()

		// Allow retries on server-side failures.
		if c.Writer.Status() >= http.StatusInternalServerError {
			s.mu.Lock()
			delete(s.entries, storeKey)
			s.mu.Unlock()
		}
	}
}
