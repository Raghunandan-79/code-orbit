package room

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/syncraft/backend/internal/collaboration"
)

func TestRoomIDGenerationAndValidation(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 500; i++ {
		id, err := GenerateRoomID()
		if err != nil {
			t.Fatalf("failed to generate room ID: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate room ID generated: %s", id)
		}
		seen[id] = true

		if err := ValidateRoomID(id); err != nil {
			t.Fatalf("generated room ID %s failed validation: %v", id, err)
		}
	}

	// Test invalid room IDs
	if err := ValidateRoomID("short"); err == nil {
		t.Errorf("expected validation failure for short ID")
	}
	if err := ValidateRoomID("invalid@characters#in$id"); err == nil {
		t.Errorf("expected validation failure for special characters")
	}
}

func TestParticipantIDGeneration(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 500; i++ {
		id, err := GenerateParticipantID()
		if err != nil {
			t.Fatalf("failed to generate participant ID: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate participant ID generated: %s", id)
		}
		seen[id] = true
		if len(id) < 10 || id[:2] != "p_" {
			t.Errorf("participant ID does not have expected prefix or length: %s", id)
		}
	}
}

func TestManagerCreateAndLookup(t *testing.T) {
	mgr := NewManager(ManagerConfig{
		MaxRooms:                  10,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        time.Hour,
	})

	r, err := mgr.CreateRoom(collaboration.LangPython)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	if mgr.ActiveRoomCount() != 1 {
		t.Errorf("expected 1 active room, got %d", mgr.ActiveRoomCount())
	}

	found, err := mgr.GetRoom(r.ID())
	if err != nil {
		t.Fatalf("failed to get room: %v", err)
	}
	if found.ID() != r.ID() {
		t.Errorf("expected room ID %s, got %s", r.ID(), found.ID())
	}

	snap, err := mgr.GetRoomSnapshot(r.ID())
	if err != nil {
		t.Fatalf("failed to get room snapshot: %v", err)
	}
	if snap.Language != collaboration.LangPython {
		t.Errorf("expected language python, got %s", snap.Language)
	}
	if snap.ParticipantCount != 0 {
		t.Errorf("expected 0 participants, got %d", snap.ParticipantCount)
	}

	// Non-existent room
	_, err = mgr.GetRoom("nonexistent-room-id")
	if !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestManagerMaxRoomsCapacity(t *testing.T) {
	mgr := NewManager(ManagerConfig{
		MaxRooms:                  3,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        time.Hour,
	})

	for i := 0; i < 3; i++ {
		_, err := mgr.CreateRoom(collaboration.LangGo)
		if err != nil {
			t.Fatalf("failed to create room %d: %v", i, err)
		}
	}

	// 4th room should exceed capacity
	_, err := mgr.CreateRoom(collaboration.LangGo)
	if !errors.Is(err, ErrMaxRoomsReached) {
		t.Errorf("expected ErrMaxRoomsReached, got %v", err)
	}
}

func TestConcurrentRoomCreation(t *testing.T) {
	mgr := NewManager(ManagerConfig{
		MaxRooms:                  200,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        time.Hour,
	})

	var wg sync.WaitGroup
	count := 100

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mgr.CreateRoom(collaboration.LangTypeScript)
			if err != nil {
				t.Errorf("concurrent room creation failed: %v", err)
			}
		}()
	}

	wg.Wait()

	if mgr.ActiveRoomCount() != count {
		t.Errorf("expected %d rooms, got %d", count, mgr.ActiveRoomCount())
	}
}

func TestRoomParticipantLimitsAndLifecycle(t *testing.T) {
	r := NewRoom("test-room", collaboration.LangRust, 2, 50)

	p1 := NewParticipant("p1", "test-room", "alice", 10)
	p2 := NewParticipant("p2", "test-room", "bob", 10)
	p3 := NewParticipant("p3", "test-room", "charlie", 10)

	if err := r.AddParticipant(p1); err != nil {
		t.Fatalf("unexpected error adding p1: %v", err)
	}
	if err := r.AddParticipant(p2); err != nil {
		t.Fatalf("unexpected error adding p2: %v", err)
	}

	// 3rd participant should hit capacity
	err := r.AddParticipant(p3)
	if !errors.Is(err, ErrRoomFull) {
		t.Fatalf("expected ErrRoomFull, got %v", err)
	}

	if r.ParticipantCount() != 2 {
		t.Errorf("expected 2 participants, got %d", r.ParticipantCount())
	}

	// Remove p1
	removed, ok := r.RemoveParticipant("p1")
	if !ok || removed.ID != "p1" {
		t.Errorf("expected p1 removed, got %v, %v", removed, ok)
	}

	if r.ParticipantCount() != 1 {
		t.Errorf("expected 1 participant remaining, got %d", r.ParticipantCount())
	}

	// Now p3 can join
	if err := r.AddParticipant(p3); err != nil {
		t.Fatalf("failed to add p3 after vacancy: %v", err)
	}
}

func TestRoomExpirationAndCleanup(t *testing.T) {
	mgr := NewManager(ManagerConfig{
		MaxRooms:                  10,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        50 * time.Millisecond,
	})

	r1, _ := mgr.CreateRoom(collaboration.LangPython)
	r2, _ := mgr.CreateRoom(collaboration.LangPython)

	// Add a participant to r2 so it's not idle
	p := NewParticipant("p1", r2.ID(), "active-user", 10)
	_ = r2.AddParticipant(p)

	// Sleep past idle timeout
	time.Sleep(70 * time.Millisecond)

	cleaned := mgr.CleanupExpiredRooms()
	if cleaned != 1 {
		t.Errorf("expected 1 room cleaned up, got %d", cleaned)
	}

	// r1 should be removed
	_, err := mgr.GetRoom(r1.ID())
	if !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("expected r1 to be expired and not found, got %v", err)
	}

	// r2 should still exist because it has a participant
	_, err = mgr.GetRoom(r2.ID())
	if err != nil {
		t.Errorf("expected r2 to remain active, got: %v", err)
	}
}

func TestParticipantQueueOverflow(t *testing.T) {
	p := NewParticipant("p1", "room-1", "fast-producer", 2)

	// Fill queue
	ok1 := p.Enqueue(OutboundMessage{IsBinary: false, Data: []byte("1")})
	ok2 := p.Enqueue(OutboundMessage{IsBinary: false, Data: []byte("2")})
	if !ok1 || !ok2 {
		t.Errorf("failed to enqueue first two items")
	}

	// 3rd item should overflow
	ok3 := p.Enqueue(OutboundMessage{IsBinary: false, Data: []byte("3")})
	if ok3 {
		t.Errorf("expected enqueue to fail on full queue")
	}

	if p.QueueDroppedCount() != 1 {
		t.Errorf("expected 1 dropped message, got %d", p.QueueDroppedCount())
	}
}
