package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
)

func TestCleanerLifecycle(t *testing.T) {
	mgr := room.NewManager(room.ManagerConfig{
		MaxRooms:                  10,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 50,
		RoomIdleExpiration:        20 * time.Millisecond,
	})
	limiter := security.NewIPRateLimiter(60, 10)

	r, err := mgr.CreateRoom(collaboration.LangPython)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	cleaner := NewCleaner(mgr, limiter, 10*time.Millisecond)
	cleaner.Start()

	// Wait for idle expiration + cleanup ticks
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := cleaner.Stop(ctx); err != nil {
		t.Fatalf("cleaner stop error: %v", err)
	}

	// Room should now be cleaned up
	_, err = mgr.GetRoom(r.ID())
	if err == nil {
		t.Errorf("expected room to be cleaned up, but it still exists")
	}
}
