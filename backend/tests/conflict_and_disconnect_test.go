package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/config"
	ws "github.com/syncraft/backend/internal/websocket"
)

// TestConflict_ConcurrentEditsOnSameLine verifies how simultaneous edits to the same line
// are handled by the backend:
// 1. Both edits are accepted atomically without lost updates.
// 2. Monotonic sequence numbers (Seq N, Seq N+1) establish canonical server ordering.
// 3. Both updates are cross-relayed to peers so CRDT engines deterministically converge.
func TestConflict_ConcurrentEditsOnSameLine(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, err := mgr.CreateRoom(collaboration.LangTypeScript)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())

	// Client 1 (Alice) connects
	cAlice, _, err := websocket.Dial(ctx, wsURL+"?nickname=Alice", nil)
	if err != nil {
		t.Fatalf("Alice dial failed: %v", err)
	}
	defer cAlice.Close(websocket.StatusNormalClosure, "")
	_, _, _ = cAlice.Read(ctx) // room_joined

	// Client 2 (Bob) connects
	cBob, _, err := websocket.Dial(ctx, wsURL+"?nickname=Bob", nil)
	if err != nil {
		t.Fatalf("Bob dial failed: %v", err)
	}
	defer cBob.Close(websocket.StatusNormalClosure, "")
	_, _, _ = cBob.Read(ctx)   // room_joined
	_, _, _ = cAlice.Read(ctx) // participant_joined for Bob

	// Establish base document line: "let counter = 0;"
	baseFrame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", []byte("line1: let counter = 0;"))
	if err := cAlice.Write(ctx, websocket.MessageBinary, baseFrame); err != nil {
		t.Fatalf("failed writing base line: %v", err)
	}

	// Bob receives base line
	_, _, _ = cBob.Read(ctx)

	// Now Alice and Bob SIMULTANEOUSLY edit the exact same line:
	// Alice edits line 1: "let counter = 10; // edited by alice"
	// Bob edits line 1:   "let counter = 20; // edited by bob"
	aliceEdit := []byte("line1_diff: client_alice: let counter = 10;")
	bobEdit := []byte("line1_diff: client_bob: let counter = 20;")

	aliceFrame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", aliceEdit)
	bobFrame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", bobEdit)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = cAlice.Write(ctx, websocket.MessageBinary, aliceFrame)
	}()

	go func() {
		defer wg.Done()
		_ = cBob.Write(ctx, websocket.MessageBinary, bobFrame)
	}()

	wg.Wait()

	// Alice should receive Bob's concurrent update
	_, aliceRecvBytes, err := cAlice.Read(ctx)
	if err != nil {
		t.Fatalf("Alice failed to receive Bob's concurrent edit: %v", err)
	}
	_, _, aliceReceivedPayload, _ := ws.DecodeBinaryFrame(aliceRecvBytes)
	if string(aliceReceivedPayload) != string(bobEdit) {
		t.Errorf("Alice expected to receive Bob's edit, got: %s", string(aliceReceivedPayload))
	}

	// Bob should receive Alice's concurrent update
	_, bobRecvBytes, err := cBob.Read(ctx)
	if err != nil {
		t.Fatalf("Bob failed to receive Alice's concurrent edit: %v", err)
	}
	_, _, bobReceivedPayload, _ := ws.DecodeBinaryFrame(bobRecvBytes)
	if string(bobReceivedPayload) != string(aliceEdit) {
		t.Errorf("Bob expected to receive Alice's edit, got: %s", string(bobReceivedPayload))
	}

	// Verify server canonical UpdateLog contains all 3 updates with strict monotonic sequence
	entries := r.UpdateLog().GetAll()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries in UpdateLog, got %d", len(entries))
	}

	if entries[0].Seq != 1 || entries[1].Seq != 2 || entries[2].Seq != 3 {
		t.Errorf("expected strictly monotonic sequence [1, 2, 3], got [%d, %d, %d]",
			entries[0].Seq, entries[1].Seq, entries[2].Seq)
	}

	// Neither user's edit was dropped or destroyed
	hasAlice := false
	hasBob := false
	for _, e := range entries {
		if string(e.Data) == string(aliceEdit) {
			hasAlice = true
		}
		if string(e.Data) == string(bobEdit) {
			hasBob = true
		}
	}
	if !hasAlice || !hasBob {
		t.Fatalf("expected both Alice and Bob edits preserved in update log! hasAlice=%v hasBob=%v", hasAlice, hasBob)
	}
}

// TestDisconnect_InternetDropOfflineEditsAndReconnect verifies:
// 1. Client experiences internet disconnection while other clients continue editing.
// 2. Client makes offline edits during disconnection.
// 3. Client reconnects passing since_seq and receives only missed updates.
// 4. Client streams buffered offline edits; peers receive them seamlessly.
// 5. Full document state convergence is achieved with zero lost updates.
func TestDisconnect_InternetDropOfflineEditsAndReconnect(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, err := mgr.CreateRoom(collaboration.LangPython)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())

	// 1. Alice and Bob connect
	cAlice, _, _ := websocket.Dial(ctx, wsURL+"?nickname=Alice", nil)
	defer cAlice.Close(websocket.StatusNormalClosure, "")
	_, _, _ = cAlice.Read(ctx) // handshake

	cBob, _, _ := websocket.Dial(ctx, wsURL+"?nickname=Bob", nil)
	_, _, _ = cBob.Read(ctx)   // handshake
	_, _, _ = cAlice.Read(ctx) // participant_joined

	// Alice sends update 1 (e.g. def calculate():)
	up1 := []byte("def calculate():")
	_ = cAlice.Write(ctx, websocket.MessageBinary, ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", up1))

	// Bob receives update 1
	_, b1Bytes, _ := cBob.Read(ctx)
	_, _, p1, _ := ws.DecodeBinaryFrame(b1Bytes)
	if string(p1) != string(up1) {
		t.Fatalf("Bob failed to receive update 1")
	}

	// 2. INTERNET DISCONNECTION: Bob's connection abruptly drops (simulating network failure)
	cBob.Close(websocket.StatusGoingAway, "network connection lost")

	// Alice receives participant_left for Bob
	_, pLeftMsg, err := cAlice.Read(ctx)
	if err != nil {
		t.Fatalf("Alice failed to read participant_left: %v", err)
	}
	var leftEnv ws.Envelope
	_ = json.Unmarshal(pLeftMsg, &leftEnv)
	if leftEnv.Type != ws.TypeParticipantLeft {
		t.Fatalf("expected participant_left, got %s", string(pLeftMsg))
	}

	// 3. WHILE BOB IS DISCONNECTED:
	// Alice makes 2 more edits (Seq 2 and Seq 3)
	up2 := []byte("    x = 10")
	up3 := []byte("    y = 20")
	_ = cAlice.Write(ctx, websocket.MessageBinary, ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", up2))
	_ = cAlice.Write(ctx, websocket.MessageBinary, ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", up3))

	time.Sleep(50 * time.Millisecond)

	// Bob, while offline, edited his local document:
	offlineBobEdit := []byte("    # offline edit: return x + y")

	// 4. INTERNET RECOVERS: Bob reconnects passing since_seq=1 (his last acknowledged Seq)
	reconnectURL := fmt.Sprintf("%s?nickname=Bob&since_seq=1", wsURL)
	cBobReconnected, _, err := websocket.Dial(ctx, reconnectURL, nil)
	if err != nil {
		t.Fatalf("Bob failed to reconnect: %v", err)
	}
	defer cBobReconnected.Close(websocket.StatusNormalClosure, "")

	// Handshake on reconnect
	_, bobHandshake, err := cBobReconnected.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read handshake on reconnect: %v", err)
	}
	var joinedEnv ws.Envelope
	_ = json.Unmarshal(bobHandshake, &joinedEnv)
	var joinedPayload ws.RoomJoinedPayload
	_ = json.Unmarshal(joinedEnv.Payload, &joinedPayload)

	// Server should report latest_seq = 3
	if joinedPayload.LatestSeq != 3 {
		t.Errorf("expected LatestSeq 3, got %d", joinedPayload.LatestSeq)
	}

	// Alice receives participant_joined for Bob's reconnected session
	_, _, _ = cAlice.Read(ctx)

	// Bob should immediately receive replayed updates 2 and 3 (and NOT update 1, since since_seq=1)
	_, recUp2, err := cBobReconnected.Read(ctx)
	if err != nil {
		t.Fatalf("Bob failed reading missed update 2: %v", err)
	}
	_, _, p2, _ := ws.DecodeBinaryFrame(recUp2)
	if string(p2) != string(up2) {
		t.Errorf("expected missed update 2 'x = 10', got: %s", string(p2))
	}

	_, recUp3, err := cBobReconnected.Read(ctx)
	if err != nil {
		t.Fatalf("Bob failed reading missed update 3: %v", err)
	}
	_, _, p3, _ := ws.DecodeBinaryFrame(recUp3)
	if string(p3) != string(up3) {
		t.Errorf("expected missed update 3 'y = 20', got: %s", string(p3))
	}

	// 5. BOB STREAMS HIS OFFLINE EDITS TO SERVER
	offlineFrame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", offlineBobEdit)
	if err := cBobReconnected.Write(ctx, websocket.MessageBinary, offlineFrame); err != nil {
		t.Fatalf("failed to send offline edits: %v", err)
	}

	// Alice receives Bob's offline edit (assigned Seq 4)
	_, aliceRecvOffline, err := cAlice.Read(ctx)
	if err != nil {
		t.Fatalf("Alice failed to receive Bob's offline edit: %v", err)
	}
	_, _, pOffline, _ := ws.DecodeBinaryFrame(aliceRecvOffline)
	if string(pOffline) != string(offlineBobEdit) {
		t.Errorf("Alice expected offline edit payload, got: %s", string(pOffline))
	}

	// Verify all 4 updates exist in server log in sequence
	allEntries := r.UpdateLog().GetAll()
	if len(allEntries) != 4 {
		t.Fatalf("expected 4 total updates in room log, got %d", len(allEntries))
	}

	if allEntries[3].Seq != 4 || string(allEntries[3].Data) != string(offlineBobEdit) {
		t.Errorf("expected Seq 4 to be Bob's offline edit, got: %+v", allEntries[3])
	}
}

// TestDisconnect_DeadPeerHeartbeatTimeout verifies that when a client disappears
// (e.g. WiFi dropped without TCP FIN), heartbeat pings time out and clean up the session.
func TestDisconnect_DeadPeerHeartbeatTimeout(t *testing.T) {
	// Configure aggressive heartbeat: ping every 50ms, timeout 50ms
	server, mgr, _ := setupTestServer(t, func(c *config.Config) {
		c.WSPingInterval = 50 * time.Millisecond
		c.WSPingTimeout = 50 * time.Millisecond
	})

	r, _ := mgr.CreateRoom(collaboration.LangPython)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Alice connects and stays active
	cAlice, _, _ := websocket.Dial(ctx, wsURL+"?nickname=Alice", nil)
	defer cAlice.Close(websocket.StatusNormalClosure, "")
	_, _, _ = cAlice.Read(ctx) // handshake

	// Bob connects
	cBob, _, _ := websocket.Dial(ctx, wsURL+"?nickname=Bob", nil)
	_, _, _ = cBob.Read(ctx)   // handshake
	_, _, _ = cAlice.Read(ctx) // participant_joined for Bob

	if r.ParticipantCount() != 2 {
		t.Fatalf("expected 2 participants, got %d", r.ParticipantCount())
	}

	// Abruptly close Bob's socket at the network level
	cBob.CloseNow()

	// Alice should receive participant_left within heartbeat detection window
	receivedLeft := make(chan bool, 1)
	go func() {
		for {
			_, msg, err := cAlice.Read(ctx)
			if err != nil {
				return
			}
			var env ws.Envelope
			if err := json.Unmarshal(msg, &env); err == nil && env.Type == ws.TypeParticipantLeft {
				receivedLeft <- true
				return
			}
		}
	}()

	select {
	case <-receivedLeft:
		// Success: participant_left received promptly
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for participant_left after dead peer disconnect")
	}

	// Verify room participant count dropped to 1
	time.Sleep(50 * time.Millisecond)
	if r.ParticipantCount() != 1 {
		t.Errorf("expected 1 participant remaining in room, got %d", r.ParticipantCount())
	}
}
