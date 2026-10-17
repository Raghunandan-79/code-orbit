package relay

import (
	"context"
	"errors"
)

var (
	ErrRelayClosed        = errors.New("relay is closed")
	ErrSubscriptionClosed = errors.New("subscription is closed")
)

// BroadcastMessage encapsulates a room-scoped message to be relayed to participants.
type BroadcastMessage struct {
	RoomID        string
	SenderID      string
	ExcludeSender bool
	IsBinary      bool
	Data          []byte
}

// Subscription represents an active listener stream for room broadcasts.
type Subscription interface {
	Channel() <-chan *BroadcastMessage
	Unsubscribe()
}

// Relay defines the transport-agnostic interface for room-scoped messaging.
// Current implementation is in-process (LocalRelay). Future distributed topologies
// (e.g., Redis Pub/Sub, NATS) can implement this interface without modifying
// WebSocket handling logic.
type Relay interface {
	Publish(ctx context.Context, msg *BroadcastMessage) error
	Subscribe(ctx context.Context, roomID string, participantID string) (Subscription, error)
	Close() error
}
