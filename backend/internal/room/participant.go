package room

import (
	"sync"
	"sync/atomic"
	"time"
)

// OutboundMessage represents a queued packet to be sent over the participant's WebSocket.
type OutboundMessage struct {
	IsBinary bool
	Data     []byte
}

// Participant represents a connected client session within a room.
type Participant struct {
	ID       string
	RoomID   string
	Nickname string
	JoinedAt time.Time
	LastSeen time.Time

	mu          sync.RWMutex
	sendQueue   chan OutboundMessage
	closed      bool
	overflowCnt atomic.Uint64
}

// NewParticipant creates an initialized participant instance with a bounded outbound queue.
func NewParticipant(id string, roomID string, nickname string, queueCapacity int) *Participant {
	if queueCapacity <= 0 {
		queueCapacity = 256
	}
	now := time.Now().UTC()
	return &Participant{
		ID:        id,
		RoomID:    roomID,
		Nickname:  nickname,
		JoinedAt:  now,
		LastSeen:  now,
		sendQueue: make(chan OutboundMessage, queueCapacity),
	}
}

// Enqueue attempts to place a message in the participant's outbound queue without blocking.
// Returns false if closed or queue is full (slow consumer detected).
func (p *Participant) Enqueue(msg OutboundMessage) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return false
	}

	select {
	case p.sendQueue <- msg:
		return true
	default:
		p.overflowCnt.Add(1)
		return false
	}
}

// OutboundChannel returns the read-only channel for the write pump.
func (p *Participant) OutboundChannel() <-chan OutboundMessage {
	return p.sendQueue
}

// Touch updates the participant's last seen timestamp.
func (p *Participant) Touch() {
	p.LastSeen = time.Now().UTC()
}

// Close gracefully closes the outbound channel once without racing against Enqueue.
func (p *Participant) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true
	close(p.sendQueue)
}

// IsClosed returns whether the participant has been closed.
func (p *Participant) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// QueueDroppedCount returns total dropped messages for this participant.
func (p *Participant) QueueDroppedCount() uint64 {
	return p.overflowCnt.Load()
}
