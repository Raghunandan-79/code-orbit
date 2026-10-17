package security

import (
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	// Capacity 2, refill rate 1 token per 100ms (10 tokens/sec)
	tb := NewTokenBucket(2, 10)

	if !tb.Allow() {
		t.Fatalf("first token should be allowed")
	}
	if !tb.Allow() {
		t.Fatalf("second token should be allowed")
	}
	// Bucket exhausted
	if tb.Allow() {
		t.Fatalf("third token should be rejected")
	}

	// Wait for refill (150ms -> ~1.5 tokens)
	time.Sleep(150 * time.Millisecond)

	if !tb.Allow() {
		t.Fatalf("token should be allowed after refill")
	}
}

func TestIPRateLimiter(t *testing.T) {
	// 60 requests per minute = 1 req/sec, burst of 2
	limiter := NewIPRateLimiter(60, 2)

	ip := "192.168.1.10"
	if !limiter.Allow(ip) {
		t.Errorf("request 1 should be allowed")
	}
	if !limiter.Allow(ip) {
		t.Errorf("request 2 should be allowed")
	}
	if limiter.Allow(ip) {
		t.Errorf("request 3 should be rejected due to burst limit")
	}

	// Different IP should still have its own quota
	otherIP := "10.0.0.1"
	if !limiter.Allow(otherIP) {
		t.Errorf("different IP should be allowed")
	}
}
