package collaboration

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrUpdateLogCapacityExceeded is returned when the room's update log limit has been reached.
	ErrUpdateLogCapacityExceeded = errors.New("update log capacity exceeded: maximum buffered updates reached")
	// ErrInvalidUpdateData is returned when empty or malformed payload is provided.
	ErrInvalidUpdateData = errors.New("update payload cannot be empty")
)

// UpdateEntry represents a single CRDT update recorded in a room.
type UpdateEntry struct {
	Seq       uint64    `json:"seq"`
	DocID     string    `json:"doc_id"`
	SenderID  string    `json:"sender_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
	IsBinary  bool      `json:"is_binary"`
}

// UpdateLog maintains an in-memory sequential history of CRDT updates for a room.
type UpdateLog struct {
	mu         sync.RWMutex
	maxEntries int
	currentSeq uint64
	entries    []UpdateEntry
}

// NewUpdateLog creates a bounded update log.
func NewUpdateLog(maxEntries int) *UpdateLog {
	if maxEntries <= 0 {
		maxEntries = 2000
	}
	return &UpdateLog{
		maxEntries: maxEntries,
		entries:    make([]UpdateEntry, 0, 128),
	}
}

// Append adds a new update to the log. If the capacity is exceeded, an error is returned.
// It assigns a monotonically increasing sequence number and server timestamp.
func (l *UpdateLog) Append(docID string, senderID string, data []byte, isBinary bool) (UpdateEntry, error) {
	if len(data) == 0 {
		return UpdateEntry{}, ErrInvalidUpdateData
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) >= l.maxEntries {
		return UpdateEntry{}, fmt.Errorf("%w (limit=%d)", ErrUpdateLogCapacityExceeded, l.maxEntries)
	}

	l.currentSeq++
	// Make a defensive copy of the update data so callers cannot mutate stored bytes
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	entry := UpdateEntry{
		Seq:       l.currentSeq,
		DocID:     docID,
		SenderID:  senderID,
		Timestamp: time.Now().UTC(),
		Data:      dataCopy,
		IsBinary:  isBinary,
	}

	l.entries = append(l.entries, entry)
	return entry, nil
}

// GetSince returns all updates with sequence number strictly greater than sinceSeq.
// Returns a defensive copy of entries slice.
func (l *UpdateLog) GetSince(sinceSeq uint64) []UpdateEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return nil
	}

	// Fast path: if sinceSeq >= currentSeq, no updates to return
	if sinceSeq >= l.currentSeq {
		return nil
	}

	// Collect matching updates
	var result []UpdateEntry
	for _, e := range l.entries {
		if e.Seq > sinceSeq {
			entryCopy := e
			entryCopy.Data = make([]byte, len(e.Data))
			copy(entryCopy.Data, e.Data)
			result = append(result, entryCopy)
		}
	}

	return result
}

// GetAll returns a copy of all recorded updates.
func (l *UpdateLog) GetAll() []UpdateEntry {
	return l.GetSince(0)
}

// Count returns the number of currently buffered updates.
func (l *UpdateLog) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// LatestSeq returns the most recent sequence number.
func (l *UpdateLog) LatestSeq() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentSeq
}

// OldestSeq returns the sequence number of the earliest buffered update, or 0 if empty.
func (l *UpdateLog) OldestSeq() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.entries) == 0 {
		return 0
	}
	return l.entries[0].Seq
}
