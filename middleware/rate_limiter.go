package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Second), 5) // 1 req/sec, burst 5
		visitors[ip] = limiter
	}
	return limiter
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, "Trop de requêtes, réessayez plus tard", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
