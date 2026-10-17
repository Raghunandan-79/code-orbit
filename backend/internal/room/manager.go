package room

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/syncraft/backend/internal/collaboration"
)

var (
	ErrMaxRoomsReached = errors.New("maximum active room capacity reached")
	ErrRoomNotFound    = errors.New("room not found or expired")
)

// ManagerConfig holds operational limits for the room manager.
type ManagerConfig struct {
	MaxRooms                  int
	MaxParticipantsPerRoom    int
	MaxBufferedUpdatesPerRoom int
	RoomIdleExpiration        time.Duration
}

// Manager maintains the concurrency-safe registry of active rooms.
type Manager struct {
	mu     sync.RWMutex
	rooms  map[string]*Room
	config ManagerConfig
	closed bool
}

// NewManager creates a new RoomManager with defined limits.
func NewManager(cfg ManagerConfig) *Manager {
	return &Manager{
		rooms:  make(map[string]*Room),
		config: cfg,
	}
}

// CreateRoom generates a new room with a secure random ID and validated language.
func (m *Manager) CreateRoom(lang collaboration.SupportedLanguage) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrRoomClosed
	}

	if len(m.rooms) >= m.config.MaxRooms {
		return nil, ErrMaxRoomsReached
	}

	// Generate a unique room ID
	var roomID string
	var err error
	for attempts := 0; attempts < 5; attempts++ {
		roomID, err = GenerateRoomID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate room ID: %w", err)
		}
		if _, exists := m.rooms[roomID]; !exists {
			break
		}
	}

	if _, exists := m.rooms[roomID]; exists {
		return nil, errors.New("unable to generate unique room ID")
	}

	r := NewRoom(roomID, lang, m.config.MaxParticipantsPerRoom, m.config.MaxBufferedUpdatesPerRoom)
	m.rooms[roomID] = r
	return r, nil
}

// GetRoom retrieves an active room by ID.
func (m *Manager) GetRoom(roomID string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrRoomClosed
	}

	r, exists := m.rooms[roomID]
	if !exists || r.IsClosed() {
		return nil, ErrRoomNotFound
	}

	return r, nil
}

// GetRoomSnapshot retrieves a non-sensitive immutable snapshot of a room.
func (m *Manager) GetRoomSnapshot(roomID string) (RoomSnapshot, error) {
	r, err := m.GetRoom(roomID)
	if err != nil {
		return RoomSnapshot{}, err
	}
	return r.Snapshot(), nil
}

// RemoveRoom deletes a room from the registry and closes it.
func (m *Manager) RemoveRoom(roomID string) {
	m.mu.Lock()
	r, exists := m.rooms[roomID]
	if exists {
		delete(m.rooms, roomID)
	}
	m.mu.Unlock()

	if exists && r != nil {
		r.Close()
	}
}

// CleanupExpiredRooms removes rooms that have had no active participants and have exceeded idle expiration.
// Returns the number of purged rooms.
func (m *Manager) CleanupExpiredRooms() int {
	m.mu.Lock()
	var toRemove []*Room
	for id, r := range m.rooms {
		if r.IsExpired(m.config.RoomIdleExpiration) {
			toRemove = append(toRemove, r)
			delete(m.rooms, id)
		}
	}
	m.mu.Unlock()

	for _, r := range toRemove {
		r.Close()
	}

	return len(toRemove)
}

// ActiveRoomCount returns the number of currently tracked rooms.
func (m *Manager) ActiveRoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// TotalParticipantsCount calculates total connected participants across all rooms.
func (m *Manager) TotalParticipantsCount() int {
	m.mu.RLock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	m.mu.RUnlock()

	total := 0
	for _, r := range rooms {
		total += r.ParticipantCount()
	}
	return total
}

// Close terminates all rooms and clears the manager.
func (m *Manager) Close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true

	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	m.rooms = make(map[string]*Room)
	m.mu.Unlock()

	for _, r := range rooms {
		r.Close()
	}
}
