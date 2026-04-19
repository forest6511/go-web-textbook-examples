package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipLimiter
	r       rate.Limit
	b       int
}

// NewIPRateLimiter は IP 単位のリミッタを作る
// rps=秒あたりの補充速度、burst=瞬間的な許容量
func NewIPRateLimiter(rps rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		buckets: make(map[string]*ipLimiter),
		r:       rps,
		b:       burst,
	}
}

func (l *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[ip]
	if !ok {
		b = &ipLimiter{limiter: rate.NewLimiter(l.r, l.b)}
		l.buckets[ip] = b
	}
	b.lastSeen = time.Now()
	return b.limiter
}

// StartGC は古いエントリを定期的に削除する。goroutine リーク防止のため
// ctx キャンセルで終了する設計にしている
func (l *IPRateLimiter) StartGC(
	ctx context.Context, interval, ttl time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			l.mu.Lock()
			for ip, b := range l.buckets {
				if now.Sub(b.lastSeen) > ttl {
					delete(l.buckets, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

// Middleware は Gin ミドルウェアを返す
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.getLimiter(ip).Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{"error": "rate limit exceeded"},
			)
			return
		}
		c.Next()
	}
}
