package router

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	tokens   int
}

func NewRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, exists := rl.visitors[ip]
	now := time.Now()
	if !exists || time.Since(v.lastSeen) > rl.window {
		rl.visitors[ip] = &visitor{lastSeen: now, tokens: rl.rate - 1}
		return true
	}
	if v.tokens > 0 {
		v.tokens--
		v.lastSeen = now
		return true
	}
	return false
}

// Middleware para usar en los handlers
func RateLimitMiddleware(rl *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !rl.Allow(ip) {
			log.Printf("Rate limit excedido para IP %s, User-Agent: %s, Path: %s", ip, r.UserAgent(), r.URL.Path)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
