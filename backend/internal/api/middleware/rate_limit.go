package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type limiterEntry struct {
	Count     int
	ResetTime time.Time
}

type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	items  map[string]*limiterEntry
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]*limiterEntry),
	}
}

func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.items[key]
	now := time.Now()
	if !ok || now.After(entry.ResetTime) {
		l.items[key] = &limiterEntry{Count: 1, ResetTime: now.Add(l.window)}
		return true
	}
	if entry.Count >= l.limit {
		return false
	}
	entry.Count++
	return true
}

func (l *RateLimiter) Middleware(code string, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !l.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    code,
					"message": message,
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
