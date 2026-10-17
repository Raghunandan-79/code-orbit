package relay

import (
	"context"
	"sync"
)

// localSubscription implements Subscription for LocalRelay.
type localSubscription struct {
	roomID        string
	participantID string
	relay         *LocalRelay
	ch            chan *BroadcastMessage

	mu     sync.RWMutex
	closed bool
}

func (s *localSubscription) Channel() <-chan *BroadcastMessage {
	return s.ch
}

func (s *localSubscription) Unsubscribe() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.ch)
	s.mu.Unlock()

	s.relay.removeSubscription(s.roomID, s.participantID)
}

func (s *localSubscription) send(msg *BroadcastMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	select {
	case s.ch <- msg:
	default:
		// Non-blocking drop on subscriber saturation to protect relay performance
	}
}

// LocalRelay implements Relay using in-memory channels and mutexes.
type LocalRelay struct {
	mu         sync.RWMutex
	subs       map[string]map[string]*localSubscription // roomID -> participantID -> sub
	bufferSize int
	closed     bool
}

// NewLocalRelay initializes a new in-process local relay.
func NewLocalRelay(subscriberBufferSize int) *LocalRelay {
	if subscriberBufferSize <= 0 {
		subscriberBufferSize = 256
	}
	return &LocalRelay{
		subs:       make(map[string]map[string]*localSubscription),
		bufferSize: subscriberBufferSize,
	}
}

// Subscribe attaches a participant to room broadcasts.
func (r *LocalRelay) Subscribe(ctx context.Context, roomID string, participantID string) (Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, ErrRelayClosed
	}

	roomSubs, exists := r.subs[roomID]
	if !exists {
		roomSubs = make(map[string]*localSubscription)
		r.subs[roomID] = roomSubs
	}

	// If existing subscription for this participant exists, unsubscribe it first
	if existing, found := roomSubs[participantID]; found {
		existing.Unsubscribe()
	}

	sub := &localSubscription{
		roomID:        roomID,
		participantID: participantID,
		relay:         r,
		ch:            make(chan *BroadcastMessage, r.bufferSize),
	}

	roomSubs[participantID] = sub
	return sub, nil
}

// Publish distributes a message to all subscribed participants in the specified room.
// Does NOT hold locks while sending or blocking on channel operations.
func (r *LocalRelay) Publish(ctx context.Context, msg *BroadcastMessage) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRelayClosed
	}

	roomSubs, exists := r.subs[msg.RoomID]
	if !exists || len(roomSubs) == 0 {
		r.mu.RUnlock()
		return nil
	}

	// Collect target subscriptions snapshot inside read lock
	targets := make([]*localSubscription, 0, len(roomSubs))
	for pid, sub := range roomSubs {
		if msg.ExcludeSender && pid == msg.SenderID {
			continue
		}
		targets = append(targets, sub)
	}
	r.mu.RUnlock()

	// Deliver messages outside of the relay lock
	for _, sub := range targets {
		sub.send(msg)
	}

	return nil
}

func (r *LocalRelay) removeSubscription(roomID string, participantID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if roomSubs, exists := r.subs[roomID]; exists {
		delete(roomSubs, participantID)
		if len(roomSubs) == 0 {
			delete(r.subs, roomID)
		}
	}
}

// ActiveSubscriberCount returns total number of subscriptions across all rooms.
func (r *LocalRelay) ActiveSubscriberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := 0
	for _, roomSubs := range r.subs {
		total += len(roomSubs)
	}
	return total
}

// Close terminates all active subscriptions and cleans up memory.
func (r *LocalRelay) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true

	// Gather all subscriptions
	var allSubs []*localSubscription
	for _, roomSubs := range r.subs {
		for _, sub := range roomSubs {
			allSubs = append(allSubs, sub)
		}
	}
	r.subs = make(map[string]map[string]*localSubscription)
	r.mu.Unlock()

	for _, sub := range allSubs {
		sub.Unsubscribe()
	}

	return nil
}
