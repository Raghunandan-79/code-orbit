package room

import (
	"errors"
	"sync"
	"time"

	"github.com/syncraft/backend/internal/collaboration"
)

var (
	ErrRoomFull   = errors.New("room participant capacity reached")
	ErrRoomClosed = errors.New("room is closed")
)

// RoomSnapshot is an immutable value object representing non-sensitive room state.
type RoomSnapshot struct {
	ID               string                          `json:"room_id"`
	CreatedAt        time.Time                       `json:"created_at"`
	LastActiveAt     time.Time                       `json:"last_active_at"`
	ParticipantCount int                             `json:"participant_count"`
	MaxParticipants  int                             `json:"max_participants"`
	Language         collaboration.SupportedLanguage `json:"language"`
	DocumentTitle    string                          `json:"document_title"`
	UpdateCount      int                             `json:"update_count"`
	IsClosed         bool                            `json:"is_closed"`
}

// ParticipantSnapshot is an immutable view of a participant.
type ParticipantSnapshot struct {
	ID       string    `json:"participant_id"`
	Nickname string    `json:"nickname,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

// Room represents an active in-memory collaborative session.
type Room struct {
	id              string
	createdAt       time.Time
	maxParticipants int

	mu           sync.RWMutex
	lastActiveAt time.Time
	closed       bool
	participants map[string]*Participant
	document     collaboration.DocumentMetadata
	updateLog    *collaboration.UpdateLog
}

// NewRoom creates a new room with initialized collaboration state.
func NewRoom(id string, lang collaboration.SupportedLanguage, maxParticipants int, maxBufferedUpdates int) *Room {
	now := time.Now().UTC()
	return &Room{
		id:              id,
		createdAt:       now,
		lastActiveAt:    now,
		maxParticipants: maxParticipants,
		participants:    make(map[string]*Participant),
		document:        collaboration.NewDefaultDocument(lang),
		updateLog:       collaboration.NewUpdateLog(maxBufferedUpdates),
	}
}

// ID returns the room's unique identifier.
func (r *Room) ID() string {
	return r.id
}

// CreatedAt returns room creation time.
func (r *Room) CreatedAt() time.Time {
	return r.createdAt
}

// MaxParticipants returns the configured participant limit.
func (r *Room) MaxParticipants() int {
	return r.maxParticipants
}

// AddParticipant attempts to register a participant in the room.
// Returns ErrRoomFull if capacity is reached, or ErrRoomClosed if the room is closing.
func (r *Room) AddParticipant(p *Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRoomClosed
	}

	if len(r.participants) >= r.maxParticipants {
		return ErrRoomFull
	}

	r.participants[p.ID] = p
	r.lastActiveAt = time.Now().UTC()
	return nil
}

// RemoveParticipant unregisters a participant and returns the removed instance.
func (r *Room) RemoveParticipant(participantID string) (*Participant, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.participants[participantID]
	if exists {
		delete(r.participants, participantID)
		r.lastActiveAt = time.Now().UTC()
	}
	return p, exists
}

// GetParticipant looks up a participant by ID safely.
func (r *Room) GetParticipant(participantID string) (*Participant, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.participants[participantID]
	return p, exists
}

// ParticipantCount returns current number of active participants.
func (r *Room) ParticipantCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.participants)
}

// ListParticipants returns a slice copy of current participants.
func (r *Room) ListParticipants() []*Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Participant, 0, len(r.participants))
	for _, p := range r.participants {
		list = append(list, p)
	}
	return list
}

// ParticipantSnapshots returns immutable snapshots of all current participants.
func (r *Room) ParticipantSnapshots() []ParticipantSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshots := make([]ParticipantSnapshot, 0, len(r.participants))
	for _, p := range r.participants {
		snapshots = append(snapshots, ParticipantSnapshot{
			ID:       p.ID,
			Nickname: p.Nickname,
			JoinedAt: p.JoinedAt,
		})
	}
	return snapshots
}

// Touch refreshes the room's last activity timestamp.
func (r *Room) Touch() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastActiveAt = time.Now().UTC()
}

// LastActiveAt returns the last recorded activity timestamp.
func (r *Room) LastActiveAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastActiveAt
}

// IsExpired checks whether an idle room has exceeded the configured expiration duration.
// An active room with connected participants is never considered expired.
func (r *Room) IsExpired(idleTimeout time.Duration) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Never expire while participants are actively connected
	if len(r.participants) > 0 {
		return false
	}

	return time.Since(r.lastActiveAt) > idleTimeout
}

// Snapshot creates a non-sensitive snapshot of the room's current state.
func (r *Room) Snapshot() RoomSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RoomSnapshot{
		ID:               r.id,
		CreatedAt:        r.createdAt,
		LastActiveAt:     r.lastActiveAt,
		ParticipantCount: len(r.participants),
		MaxParticipants:  r.maxParticipants,
		Language:         r.document.Language,
		DocumentTitle:    r.document.Title,
		UpdateCount:      r.updateLog.Count(),
		IsClosed:         r.closed,
	}
}

// UpdateLog returns the room's collaboration update log.
func (r *Room) UpdateLog() *collaboration.UpdateLog {
	return r.updateLog
}

// DocumentMetadata returns the current document metadata.
func (r *Room) DocumentMetadata() collaboration.DocumentMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.document
}

// Close gracefully terminates the room and all connected participants.
func (r *Room) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true

	// Gather participants to close outside the lock
	participants := make([]*Participant, 0, len(r.participants))
	for _, p := range r.participants {
		participants = append(participants, p)
	}
	r.participants = make(map[string]*Participant)
	r.mu.Unlock()

	for _, p := range participants {
		p.Close()
	}
}

// IsClosed returns whether the room is closed.
func (r *Room) IsClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}
