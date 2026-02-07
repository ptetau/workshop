package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/csrf"
)

// RateLimiter provides a per-IP token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // tokens per interval
	interval time.Duration // refill interval
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter allowing `rate` requests per `interval`.
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
	}
	// Cleanup stale visitors every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// Allow checks if a request from the given IP is allowed.
// PRE: ip is non-empty
// POST: Returns true if within rate limit, false if exceeded
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{tokens: rl.rate - 1, lastSeen: time.Now()}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := time.Since(v.lastSeen)
	refill := int(elapsed/rl.interval) * rl.rate
	v.tokens += refill
	if v.tokens > rl.rate {
		v.tokens = rl.rate
	}
	v.lastSeen = time.Now()

	if v.tokens <= 0 {
		slog.Warn("rate_limit_exceeded", "ip", ip)
		return false
	}
	v.tokens--
	return true
}

// RateLimit returns middleware that limits requests per IP.
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.Allow(ip) {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds OWASP recommended headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy (Basic)
		// Allow scripts/styles from 'self' and inline (for now, scaffold simplicity)
		// In production, use nonces/hashes.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'; img-src 'self' https://img.youtube.com; frame-src https://www.youtube.com; connect-src 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// CSRFMiddleware returns a handler that protects against CSRF attacks.
// It assumes an encryption key is passed (32 bytes).
// JSON API requests (Content-Type: application/json) are exempted from CSRF.
func CSRF(authKey []byte) func(http.Handler) http.Handler {
	csrfProtect := csrf.Protect(
		authKey,
		csrf.Secure(false), // Allow HTTP for local development
		csrf.Path("/"),
		csrf.TrustedOrigins([]string{"localhost:8080", "127.0.0.1:8080"}), // Trust local origins
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Exempt JSON API requests from CSRF protection
			if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				next.ServeHTTP(w, r)
				return
			}
			// Apply CSRF protection for form submissions
			csrfProtect(next).ServeHTTP(w, r)
		})
	}
}

// Chain applies middlewares in order (outer to inner).
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
