package handlers

import (
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/room"
)

// HealthHandler provides health check and readiness endpoints.
type HealthHandler struct {
	startTime time.Time
	roomMgr   *room.Manager
	ready     atomic.Bool
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(roomMgr *room.Manager) *HealthHandler {
	h := &HealthHandler{
		startTime: time.Now().UTC(),
		roomMgr:   roomMgr,
	}
	h.ready.Store(true)
	return h
}

// SetReady updates readiness status.
func (h *HealthHandler) SetReady(ready bool) {
	h.ready.Store(ready)
}

// Healthz responds with process health metrics.
func (h *HealthHandler) Healthz(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime).Seconds()

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"uptime_seconds":  int64(uptime),
		"active_rooms":    h.roomMgr.ActiveRoomCount(),
		"participants":    h.roomMgr.TotalParticipantsCount(),
		"memory_alloc_mb": m.Alloc / 1024 / 1024,
		"goroutines":      runtime.NumGoroutine(),
	})
}

// Readyz responds with server readiness to accept traffic.
func (h *HealthHandler) Readyz(c *gin.Context) {
	if !h.ready.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "shutting_down",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
