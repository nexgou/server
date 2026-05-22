package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nexgou/server/src/common"
)

// ── Global rate limit middleware ───────────────────────────────────────────────

// RateLimit returns a global middleware that limits each client IP to max requests
// within the given window. Excess requests receive 429 Too Many Requests.
//
// Response headers set on every request:
//
//	X-RateLimit-Limit:     maximum requests allowed in the window
//	X-RateLimit-Remaining: requests remaining in the current window
//
// Response header set on 429:
//
//	Retry-After: seconds until the current window resets
//
// Usage:
//
//	app.Use(middleware.RateLimit(100, time.Minute))
func RateLimit(max int, window time.Duration) common.MiddlewareFunc {
	store := newBucketStore(window)
	return func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			ip := clientIP(ctx.Request)
			remaining, retryAfter, allowed := store.take(ip, max)

			ctx.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(max))
			ctx.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				ctx.Writer.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
				return common.NewHttpException(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(ctx)
		}
	}
}

// ── Per-route rate limit guard ─────────────────────────────────────────────────

// RateLimitGuard is a Guard that enforces a per-route rate limit per client IP.
// Attach it to individual routes using .Guard(...).
//
// Usage:
//
//	nexgou.Post("/login", c.Login).
//	    Guard(&middleware.RateLimitGuard{Max: 5, Window: time.Minute})
type RateLimitGuard struct {
	// Max is the maximum number of requests allowed within Window.
	Max int
	// Window is the sliding time window for the rate limit.
	Window time.Duration

	once  sync.Once
	store *bucketStore
}

func (g *RateLimitGuard) CanActivate(ctx *common.Context) (bool, error) {
	g.once.Do(func() {
		g.store = newBucketStore(g.Window)
	})

	ip := clientIP(ctx.Request)
	remaining, retryAfter, allowed := g.store.take(ip, g.Max)

	ctx.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(g.Max))
	ctx.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

	if !allowed {
		ctx.Writer.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
		return false, common.NewHttpException(http.StatusTooManyRequests, "rate limit exceeded")
	}
	return true, nil
}

// ── Bucket store (fixed window counter per IP) ─────────────────────────────────

type bucket struct {
	count     int
	resetAt   time.Time
}

type bucketStore struct {
	mu     sync.Mutex
	window time.Duration
	data   map[string]*bucket
}

func newBucketStore(window time.Duration) *bucketStore {
	s := &bucketStore{
		window: window,
		data:   make(map[string]*bucket),
	}
	go s.cleanup()
	return s
}

// take increments the counter for ip and returns (remaining, retryAfter, allowed).
func (s *bucketStore) take(ip string, max int) (int, time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	b, ok := s.data[ip]
	if !ok || now.After(b.resetAt) {
		b = &bucket{count: 0, resetAt: now.Add(s.window)}
		s.data[ip] = b
	}

	b.count++
	remaining := max - b.count
	if remaining < 0 {
		remaining = 0
	}

	if b.count > max {
		return remaining, time.Until(b.resetAt), false
	}
	return remaining, 0, true
}

// cleanup periodically removes expired buckets to prevent unbounded memory growth.
func (s *bucketStore) cleanup() {
	ticker := time.NewTicker(s.window * 2)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for ip, b := range s.data {
			if now.After(b.resetAt) {
				delete(s.data, ip)
			}
		}
		s.mu.Unlock()
	}
}

// ── IP extraction ──────────────────────────────────────────────────────────────

// clientIP extracts the real client IP from the request.
// It checks X-Forwarded-For first (leftmost, most client-side IP),
// then X-Real-IP, and finally falls back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be "client, proxy1, proxy2"
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
