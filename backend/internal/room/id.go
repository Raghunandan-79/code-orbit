package room

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

var (
	// RoomIDRegex restricts room IDs to URL-safe alphanumeric characters and hyphens/underscores.
	roomIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,64}$`)

	ErrInvalidRoomID = errors.New("invalid room ID format: must be 8-64 alphanumeric, dash or underscore characters")
)

// GenerateRoomID generates a cryptographically secure, unpredictable room identifier.
// 16 bytes of entropy encoded with URL-safe base64 without padding yields a 22-character string.
func GenerateRoomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random room ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateParticipantID generates an ephemeral, server-assigned participant ID.
// Prefix 'p_' followed by 16 hex characters (8 bytes entropy).
func GenerateParticipantID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random participant ID: %w", err)
	}
	return "p_" + hex.EncodeToString(bytes), nil
}

// ValidateRoomID verifies that a room ID is well-formed.
func ValidateRoomID(roomID string) error {
	if !roomIDRegex.MatchString(roomID) {
		return ErrInvalidRoomID
	}
	return nil
}
