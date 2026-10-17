package relay

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLocalRelayPublishSubscribe(t *testing.T) {
	rel := NewLocalRelay(10)
	defer rel.Close()

	ctx := context.Background()
	roomID := "room-123"

	sub1, err := rel.Subscribe(ctx, roomID, "p1")
	if err != nil {
		t.Fatalf("failed to subscribe p1: %v", err)
	}
	defer sub1.Unsubscribe()

	sub2, err := rel.Subscribe(ctx, roomID, "p2")
	if err != nil {
		t.Fatalf("failed to subscribe p2: %v", err)
	}
	defer sub2.Unsubscribe()

	// Publish message from p1, excluding sender
	msg := &BroadcastMessage{
		RoomID:        roomID,
		SenderID:      "p1",
		ExcludeSender: true,
		IsBinary:      false,
		Data:          []byte("hello room"),
	}

	if err := rel.Publish(ctx, msg); err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// sub2 should receive
	select {
	case received := <-sub2.Channel():
		if string(received.Data) != "hello room" {
			t.Errorf("unexpected message data: %s", string(received.Data))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for message on sub2")
	}

	// sub1 should NOT receive (excluded)
	select {
	case unexpected := <-sub1.Channel():
		t.Fatalf("sub1 unexpectedly received message: %v", unexpected)
	default:
		// success
	}
}

func TestLocalRelayRoomIsolation(t *testing.T) {
	rel := NewLocalRelay(10)
	defer rel.Close()

	ctx := context.Background()

	subA, _ := rel.Subscribe(ctx, "room-A", "pA")
	defer subA.Unsubscribe()

	subB, _ := rel.Subscribe(ctx, "room-B", "pB")
	defer subB.Unsubscribe()

	// Publish to room-A
	_ = rel.Publish(ctx, &BroadcastMessage{
		RoomID:   "room-A",
		SenderID: "sender",
		Data:     []byte("for room A only"),
	})

	// subA receives
	select {
	case rec := <-subA.Channel():
		if string(rec.Data) != "for room A only" {
			t.Errorf("unexpected data: %s", string(rec.Data))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for message on subA")
	}

	// subB receives nothing
	select {
	case rec := <-subB.Channel():
		t.Fatalf("subB received message intended for room A: %v", rec)
	default:
		// success
	}
}

func TestLocalRelayConcurrentPublishAndSubscribe(t *testing.T) {
	rel := NewLocalRelay(50)
	defer rel.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	numRooms := 5
	participantsPerRoom := 5

	// Create subscribers
	for r := 0; r < numRooms; r++ {
		roomID := fmt.Sprintf("room-%d", r)
		for p := 0; p < participantsPerRoom; p++ {
			pid := fmt.Sprintf("p-%d", p)
			sub, err := rel.Subscribe(ctx, roomID, pid)
			if err != nil {
				t.Fatalf("subscribe error: %v", err)
			}

			// Drain routine
			wg.Add(1)
			go func(s Subscription) {
				defer wg.Done()
				for range s.Channel() {
					// drain
				}
			}(sub)
		}
	}

	// Concurrent publishers
	for r := 0; r < numRooms; r++ {
		roomID := fmt.Sprintf("room-%d", r)
		wg.Add(1)
		go func(rID string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				_ = rel.Publish(ctx, &BroadcastMessage{
					RoomID:   rID,
					SenderID: "pub",
					Data:     []byte("test data"),
				})
			}
		}(roomID)
	}

	// Allow publishers to finish
	time.Sleep(50 * time.Millisecond)

	// Close relay to unblock drain routines
	_ = rel.Close()
	wg.Wait()
}
