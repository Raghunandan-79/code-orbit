package security

import (
	"sync"
	"time"
)

// TokenBucket implements a thread-safe token bucket rate limiter.
type TokenBucket struct {
	mu           sync.Mutex
	rate         float64   // tokens added per second
	capacity     float64   // maximum token capacity
	tokens       float64   // current token count
	lastRefilled time.Time // last time tokens were added
}

// NewTokenBucket creates a token bucket with given capacity and fill rate (tokens/sec).
func NewTokenBucket(capacity float64, rate float64) *TokenBucket {
	return &TokenBucket{
		rate:         rate,
		capacity:     capacity,
		tokens:       capacity,
		lastRefilled: time.Now().UTC(),
	}
}

// Allow checks if a single token can be consumed.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN checks if n tokens can be consumed.
func (tb *TokenBucket) AllowN(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now().UTC()
	elapsed := now.Sub(tb.lastRefilled).Seconds()
	tb.lastRefilled = now

	// Refill tokens
	tb.tokens = tb.tokens + elapsed*tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}

	return false
}

// IPRateLimiter tracks token buckets per IP address with periodic stale cleanup.
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*ipEntry
	rate     float64
	capacity float64
	ttl      time.Duration
}

type ipEntry struct {
	bucket   *TokenBucket
	lastSeen time.Time
}

// NewIPRateLimiter creates an IP-keyed rate limiter.
func NewIPRateLimiter(ratePerMinute int, burst int) *IPRateLimiter {
	ratePerSec := float64(ratePerMinute) / 60.0
	limiter := &IPRateLimiter{
		limiters: make(map[string]*ipEntry),
		rate:     ratePerSec,
		capacity: float64(burst),
		ttl:      10 * time.Minute,
	}
	return limiter
}

// Allow checks if the given IP address is within rate limits.
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	entry, exists := l.limiters[ip]
	if !exists {
		entry = &ipEntry{
			bucket:   NewTokenBucket(l.capacity, l.rate),
			lastSeen: now,
		}
		l.limiters[ip] = entry
	}

	entry.lastSeen = now
	return entry.bucket.Allow()
}

// CleanupStale removes IP entries that have been inactive for longer than TTL.
func (l *IPRateLimiter) CleanupStale() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	purged := 0
	for ip, entry := range l.limiters {
		if now.Sub(entry.lastSeen) > l.ttl {
			delete(l.limiters, ip)
			purged++
		}
	}
	return purged
}
