package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/room"
)

const ProtocolVersion = 1

type MessageType string

const (
	TypeJoinRoom          MessageType = "join_room"
	TypeRoomJoined        MessageType = "room_joined"
	TypeSyncRequest       MessageType = "sync_request"
	TypeSyncUpdate        MessageType = "sync_update"
	TypeAwarenessUpdate   MessageType = "awareness_update"
	TypeParticipantJoined MessageType = "participant_joined"
	TypeParticipantLeft   MessageType = "participant_left"
	TypeRoomState         MessageType = "room_state"
	TypeError             MessageType = "error"
	TypePing              MessageType = "ping"
	TypePong              MessageType = "pong"
)

// Binary Message Subtypes
const (
	BinarySubtypeCRDTUpdate byte = 0x01
	BinarySubtypeAwareness  byte = 0x02
	BinarySubtypeSyncStep1  byte = 0x03
	BinarySubtypeSyncStep2  byte = 0x04
)

var (
	ErrMalformedEnvelope   = errors.New("malformed envelope: missing required fields")
	ErrUnsupportedVersion  = errors.New("unsupported protocol version")
	ErrBinaryFrameTooShort = errors.New("binary frame too short")
	ErrBinaryDocIDTooLong  = errors.New("binary document ID exceeds maximum allowed length")
)

// Envelope is the standard JSON control envelope.
type Envelope struct {
	Version       int             `json:"version"`
	Type          MessageType     `json:"type"`
	RoomID        string          `json:"room_id,omitempty"`
	ParticipantID string          `json:"participant_id,omitempty"`
	DocID         string          `json:"doc_id,omitempty"`
	Timestamp     int64           `json:"timestamp"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

// NewEnvelope creates a validated JSON envelope.
func NewEnvelope(msgType MessageType, roomID, participantID, docID string, payload interface{}) (*Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		raw = b
	}

	if docID == "" {
		docID = "default"
	}

	return &Envelope{
		Version:       ProtocolVersion,
		Type:          msgType,
		RoomID:        roomID,
		ParticipantID: participantID,
		DocID:         docID,
		Timestamp:     time.Now().UTC().UnixMilli(),
		Payload:       raw,
	}, nil
}

// Validate ensures envelope satisfies structural constraints.
func (e *Envelope) Validate() error {
	if e.Version != ProtocolVersion {
		return fmt.Errorf("%w: received %d, expected %d", ErrUnsupportedVersion, e.Version, ProtocolVersion)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: missing message type", ErrMalformedEnvelope)
	}
	return nil
}

// Payload structs

type JoinRoomPayload struct {
	Nickname string `json:"nickname,omitempty"`
}

type RoomJoinedPayload struct {
	RoomID          string                          `json:"room_id"`
	ParticipantID   string                          `json:"participant_id"`
	Language        collaboration.SupportedLanguage `json:"language"`
	DocumentTitle   string                          `json:"document_title"`
	CreatedAt       time.Time                       `json:"created_at"`
	MaxParticipants int                             `json:"max_participants"`
	Participants    []room.ParticipantSnapshot      `json:"participants"`
	LatestSeq       uint64                          `json:"latest_seq"`
	OldestSeq       uint64                          `json:"oldest_seq"`
}

type SyncRequestPayload struct {
	SinceSeq uint64 `json:"since_seq"`
}

type SyncUpdatePayload struct {
	Seq   uint64 `json:"seq"`
	DocID string `json:"doc_id,omitempty"`
	Data  []byte `json:"data"`
}

type ParticipantJoinedPayload struct {
	ParticipantID string    `json:"participant_id"`
	Nickname      string    `json:"nickname"`
	JoinedAt      time.Time `json:"joined_at"`
}

type ParticipantLeftPayload struct {
	ParticipantID string `json:"participant_id"`
	Reason        string `json:"reason"`
}

type RoomStatePayload struct {
	RoomID           string                          `json:"room_id"`
	Language         collaboration.SupportedLanguage `json:"language"`
	ParticipantCount int                             `json:"participant_count"`
	MaxParticipants  int                             `json:"max_participants"`
	Participants     []room.ParticipantSnapshot      `json:"participants"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Binary Framing Helpers
// Layout: [0: Version(1B)][1: Subtype(1B)][2: DocIDLength(1B)][3: Flags(1B)][4..4+DocIDLen: DocID][remaining: OpaquePayload]
func EncodeBinaryFrame(subtype byte, docID string, payload []byte) []byte {
	if docID == "" {
		docID = "default"
	}
	docBytes := []byte(docID)
	docLen := len(docBytes)
	if docLen > 255 {
		docLen = 255
		docBytes = docBytes[:255]
	}

	frame := make([]byte, 4+docLen+len(payload))
	frame[0] = ProtocolVersion
	frame[1] = subtype
	frame[2] = byte(docLen)
	frame[3] = 0x00 // reserved flags

	copy(frame[4:4+docLen], docBytes)
	copy(frame[4+docLen:], payload)
	return frame
}

// DecodeBinaryFrame decodes framed binary frames, or falls back to raw Yjs payload.
func DecodeBinaryFrame(data []byte) (byte, string, []byte, error) {
	if len(data) == 0 {
		return 0, "", nil, ErrBinaryFrameTooShort
	}

	// If header matches protocol version 1 and has at least 4 bytes
	if len(data) >= 4 && data[0] == ProtocolVersion {
		subtype := data[1]
		docLen := int(data[2])
		if len(data) >= 4+docLen {
			docID := "default"
			if docLen > 0 {
				docID = string(data[4 : 4+docLen])
			}
			payload := data[4+docLen:]
			return subtype, docID, payload, nil
		}
	}

	// Fallback to raw binary frame: subtype CRDT update on default doc
	return BinarySubtypeCRDTUpdate, "default", data, nil
}
