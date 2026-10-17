package lifecycle

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
)

// Cleaner runs periodic background maintenance tasks (idle room cleanup, rate limiter pruning).
type Cleaner struct {
	manager  *room.Manager
	limiter  *security.IPRateLimiter
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewCleaner creates a lifecycle maintenance worker.
func NewCleaner(mgr *room.Manager, limiter *security.IPRateLimiter, interval time.Duration) *Cleaner {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	return &Cleaner{
		manager:  mgr,
		limiter:  limiter,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start spawns the periodic cleanup loop.
func (c *Cleaner) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.runOnce()
			case <-c.stopCh:
				return
			}
		}
	}()
}

func (c *Cleaner) runOnce() {
	if c.manager != nil {
		purged := c.manager.CleanupExpiredRooms()
		if purged > 0 {
			log.Printf("[lifecycle] cleaned up %d expired idle rooms", purged)
		}
	}
	if c.limiter != nil {
		c.limiter.CleanupStale()
	}
}

// Stop halts the cleaner and waits for in-flight cleanup to finish.
func (c *Cleaner) Stop(ctx context.Context) error {
	close(c.stopCh)
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
