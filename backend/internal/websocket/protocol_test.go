package websocket

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEnvelopeValidation(t *testing.T) {
	env, err := NewEnvelope(TypeJoinRoom, "room-1", "p-1", "default", JoinRoomPayload{Nickname: "alice"})
	if err != nil {
		t.Fatalf("failed to create envelope: %v", err)
	}

	if err := env.Validate(); err != nil {
		t.Fatalf("envelope validation failed: %v", err)
	}

	// Invalid version
	env.Version = 2
	if err := env.Validate(); !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("expected ErrUnsupportedVersion, got: %v", err)
	}

	// Empty type
	env.Version = ProtocolVersion
	env.Type = ""
	if err := env.Validate(); !errors.Is(err, ErrMalformedEnvelope) {
		t.Errorf("expected ErrMalformedEnvelope, got: %v", err)
	}
}

func TestBinaryFramingRoundtrip(t *testing.T) {
	docID := "document-42"
	payload := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}

	framed := EncodeBinaryFrame(BinarySubtypeCRDTUpdate, docID, payload)
	subtype, decodedDoc, decodedPayload, err := DecodeBinaryFrame(framed)
	if err != nil {
		t.Fatalf("failed to decode framed binary: %v", err)
	}

	if subtype != BinarySubtypeCRDTUpdate {
		t.Errorf("expected subtype %d, got %d", BinarySubtypeCRDTUpdate, subtype)
	}
	if decodedDoc != docID {
		t.Errorf("expected docID %s, got %s", docID, decodedDoc)
	}
	if string(decodedPayload) != string(payload) {
		t.Errorf("payload mismatch")
	}
}

func TestBinaryFramingRawFallback(t *testing.T) {
	// Raw payload without framing header (e.g., standard Yjs binary update starting with non-version-1 or arbitrary bytes)
	rawPayload := []byte{0x05, 0x12, 0x34}
	subtype, docID, payload, err := DecodeBinaryFrame(rawPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if subtype != BinarySubtypeCRDTUpdate {
		t.Errorf("expected default subtype CRDTUpdate, got %d", subtype)
	}
	if docID != "default" {
		t.Errorf("expected default docID, got %s", docID)
	}
	if len(payload) != len(rawPayload) {
		t.Errorf("expected raw payload preserved")
	}
}

func TestJSONEnvelopeSerialization(t *testing.T) {
	env, err := NewEnvelope(TypeRoomState, "r-1", "p-1", "default", RoomStatePayload{
		RoomID:           "r-1",
		ParticipantCount: 3,
		MaxParticipants:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Envelope
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Type != TypeRoomState || parsed.RoomID != "r-1" {
		t.Errorf("mismatched parsed envelope: %+v", parsed)
	}
}
