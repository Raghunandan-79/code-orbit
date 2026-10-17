package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/http/handlers"
	"github.com/syncraft/backend/internal/http/routes"
	"github.com/syncraft/backend/internal/relay"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
	ws "github.com/syncraft/backend/internal/websocket"
)

func setupTestServer(t *testing.T, customCfg ...func(*config.Config)) (*httptest.Server, *room.Manager, *config.Config) {
	cfg := &config.Config{
		Host:                      "127.0.0.1",
		Port:                      "0",
		ReadTimeout:               5 * time.Second,
		WriteTimeout:              5 * time.Second,
		GracefulShutdownTimeout:   5 * time.Second,
		AllowedOrigins:            []string{"*"},
		MaxRooms:                  100,
		MaxParticipantsPerRoom:    5,
		MaxWebSocketMessageSize:   64 * 1024,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        time.Hour,
		RoomCleanupInterval:       time.Hour,
		HTTPRateLimitRPM:          1000,
		HTTPRateLimitBurst:        500,
		WSMessageRateLimit:        200,
	}

	for _, c := range customCfg {
		c(cfg)
	}

	roomMgr := room.NewManager(room.ManagerConfig{
		MaxRooms:                  cfg.MaxRooms,
		MaxParticipantsPerRoom:    cfg.MaxParticipantsPerRoom,
		MaxBufferedUpdatesPerRoom: cfg.MaxBufferedUpdatesPerRoom,
		RoomIdleExpiration:        cfg.RoomIdleExpiration,
	})

	localRelay := relay.NewLocalRelay(128)
	ipLimiter := security.NewIPRateLimiter(cfg.HTTPRateLimitRPM, cfg.HTTPRateLimitBurst)

	healthH := handlers.NewHealthHandler(roomMgr)
	roomH := handlers.NewRoomHandler(cfg, roomMgr)
	wsH := ws.NewHandler(cfg, roomMgr, localRelay)

	engine := routes.SetupRouter(cfg, healthH, roomH, wsH, ipLimiter)
	server := httptest.NewServer(engine)

	t.Cleanup(func() {
		server.Close()
		localRelay.Close()
		roomMgr.Close()
	})

	return server, roomMgr, cfg
}

func TestIntegration_MultiClientBroadcast(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	// Create Room
	r, err := mgr.CreateRoom(collaboration.LangPython)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())

	// Client 1 (Alice) connects
	c1, _, err := websocket.Dial(ctx, wsURL+"?nickname=Alice", nil)
	if err != nil {
		t.Fatalf("client 1 dial failed: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	// Read Client 1 initial room_joined envelope
	_, c1Msg, err := c1.Read(ctx)
	if err != nil {
		t.Fatalf("c1 failed to read room_joined: %v", err)
	}
	var env1 ws.Envelope
	if err := json.Unmarshal(c1Msg, &env1); err != nil || env1.Type != ws.TypeRoomJoined {
		t.Fatalf("expected room_joined envelope, got: %s", string(c1Msg))
	}

	// Client 2 (Bob) connects
	c2, _, err := websocket.Dial(ctx, wsURL+"?nickname=Bob", nil)
	if err != nil {
		t.Fatalf("client 2 dial failed: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	// Read Client 2 initial room_joined envelope
	_, _, err = c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed to read room_joined: %v", err)
	}

	// Client 1 should receive participant_joined for Bob
	_, pJoinedMsg, err := c1.Read(ctx)
	if err != nil {
		t.Fatalf("c1 failed to read participant_joined: %v", err)
	}
	var pJoinedEnv ws.Envelope
	if err := json.Unmarshal(pJoinedMsg, &pJoinedEnv); err != nil || pJoinedEnv.Type != ws.TypeParticipantJoined {
		t.Fatalf("expected participant_joined, got: %s", string(pJoinedMsg))
	}

	// Client 1 sends a Binary CRDT update
	crdtPayload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	binaryFrame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", crdtPayload)

	if err := c1.Write(ctx, websocket.MessageBinary, binaryFrame); err != nil {
		t.Fatalf("c1 failed to write binary frame: %v", err)
	}

	// Client 2 should receive the binary frame
	msgType, receivedBytes, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed to read broadcast: %v", err)
	}
	if msgType != websocket.MessageBinary {
		t.Fatalf("expected binary message type, got: %v", msgType)
	}

	subtype, docID, payload, err := ws.DecodeBinaryFrame(receivedBytes)
	if err != nil {
		t.Fatalf("failed to decode received binary frame: %v", err)
	}
	if subtype != ws.BinarySubtypeCRDTUpdate || docID != "default" {
		t.Errorf("mismatched subtype (%d) or docID (%s)", subtype, docID)
	}
	if string(payload) != string(crdtPayload) {
		t.Errorf("payload mismatch: expected %v, got %v", crdtPayload, payload)
	}
}

func TestIntegration_InitialSyncAndReplay(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, err := mgr.CreateRoom(collaboration.LangGo)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())

	// Client 1 connects and sends 2 updates
	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("c1 dial failed: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	// Read c1 handshake
	_, _, _ = c1.Read(ctx)

	// Send update 1
	frame1 := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", []byte("update-one"))
	_ = c1.Write(ctx, websocket.MessageBinary, frame1)

	// Send update 2
	frame2 := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", []byte("update-two"))
	_ = c1.Write(ctx, websocket.MessageBinary, frame2)

	// Brief pause to allow update log recording
	time.Sleep(50 * time.Millisecond)

	// Now Client 2 (late joiner) connects
	c2, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("c2 dial failed: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	// c2 first receives room_joined
	_, rJoined, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed reading handshake: %v", err)
	}
	var joinedEnv ws.Envelope
	_ = json.Unmarshal(rJoined, &joinedEnv)
	if joinedEnv.Type != ws.TypeRoomJoined {
		t.Fatalf("expected room_joined, got %s", string(rJoined))
	}

	// c2 then immediately receives replayed update 1
	_, up1Bytes, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed reading replayed update 1: %v", err)
	}
	_, _, p1, _ := ws.DecodeBinaryFrame(up1Bytes)
	if string(p1) != "update-one" {
		t.Errorf("expected update-one, got %s", string(p1))
	}

	// c2 then receives replayed update 2
	_, up2Bytes, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed reading replayed update 2: %v", err)
	}
	_, _, p2, _ := ws.DecodeBinaryFrame(up2Bytes)
	if string(p2) != "update-two" {
		t.Errorf("expected update-two, got %s", string(p2))
	}

	// Now test sync_request with since_seq = 1
	reqEnv, _ := ws.NewEnvelope(ws.TypeSyncRequest, r.ID(), "client-2", "default", ws.SyncRequestPayload{SinceSeq: 1})
	reqBytes, _ := json.Marshal(reqEnv)
	_ = c2.Write(ctx, websocket.MessageText, reqBytes)

	// c2 should receive update-two (seq 2 > 1)
	_, filteredUpdate, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("c2 failed reading filtered update: %v", err)
	}
	_, _, pFiltered, _ := ws.DecodeBinaryFrame(filteredUpdate)
	if string(pFiltered) != "update-two" {
		t.Errorf("expected update-two on sync_request since_seq=1, got %s", string(pFiltered))
	}
}

func TestIntegration_OriginValidation(t *testing.T) {
	allowedOrigin := "https://authorized.example.com"
	server, mgr, _ := setupTestServer(t, func(c *config.Config) {
		c.AllowedOrigins = []string{allowedOrigin}
	})

	r, _ := mgr.CreateRoom(collaboration.LangRust)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Connection with unauthorized origin should fail
	badOpts := &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://malicious.evil.com"}},
	}
	_, resp, err := websocket.Dial(ctx, wsURL, badOpts)
	if err == nil {
		t.Fatal("expected unauthorized dial to fail, but succeeded")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", resp.StatusCode)
	}

	// 2. Connection with authorized origin should succeed
	goodOpts := &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{allowedOrigin}},
	}
	cGood, respGood, err := websocket.Dial(ctx, wsURL, goodOpts)
	if err != nil {
		t.Fatalf("expected authorized dial to succeed, got %v (resp: %v)", err, respGood)
	}
	cGood.Close(websocket.StatusNormalClosure, "")
}

func TestIntegration_RoomCapacityEnforcement(t *testing.T) {
	server, mgr, _ := setupTestServer(t, func(c *config.Config) {
		c.MaxParticipantsPerRoom = 2
	})

	// Create room with capacity 2
	r, _ := mgr.CreateRoom(collaboration.LangPython)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Client 1 connects
	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("c1 connect failed: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	// Client 2 connects
	c2, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("c2 connect failed: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	// Client 3 should be rejected with 403 Forbidden (capacity 2 reached)
	_, resp, err := websocket.Dial(ctx, wsURL, nil)
	if err == nil {
		t.Fatal("expected client 3 to be rejected due to capacity limit")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected HTTP 403 Forbidden on capacity limit, got %d", resp.StatusCode)
	}
}

func TestIntegration_ParticipantDisconnectNotification(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, _ := mgr.CreateRoom(collaboration.LangJavaScript)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL+"?nickname=alice", nil)
	if err != nil {
		t.Fatalf("c1 dial failed: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	// Read c1 room_joined
	_, _, _ = c1.Read(ctx)

	c2, _, err := websocket.Dial(ctx, wsURL+"?nickname=bob", nil)
	if err != nil {
		t.Fatalf("c2 dial failed: %v", err)
	}

	// c1 reads participant_joined
	_, _, _ = c1.Read(ctx)

	// Close c2
	c2.Close(websocket.StatusNormalClosure, "leaving")

	// c1 should receive participant_left
	_, leftMsg, err := c1.Read(ctx)
	if err != nil {
		t.Fatalf("c1 failed to read participant_left: %v", err)
	}

	var leftEnv ws.Envelope
	if err := json.Unmarshal(leftMsg, &leftEnv); err != nil || leftEnv.Type != ws.TypeParticipantLeft {
		t.Fatalf("expected participant_left, got: %s", string(leftMsg))
	}
}

func TestIntegration_MalformedEnvelopeAndUnsupportedVersion(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, _ := mgr.CreateRoom(collaboration.LangPython)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	// Discard room_joined
	_, _, _ = c.Read(ctx)

	// Send unsupported version
	badEnv := map[string]interface{}{
		"version": 999,
		"type":    "sync_request",
	}
	b, _ := json.Marshal(badEnv)
	_ = c.Write(ctx, websocket.MessageText, b)

	// Expect error envelope response
	_, errMsg, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read error response: %v", err)
	}

	var errEnv ws.Envelope
	if err := json.Unmarshal(errMsg, &errEnv); err != nil || errEnv.Type != ws.TypeError {
		t.Fatalf("expected error envelope, got: %s", string(errMsg))
	}
	if !strings.Contains(string(errEnv.Payload), "unsupported protocol version") {
		t.Errorf("unexpected error payload: %s", string(errEnv.Payload))
	}
}

func TestIntegration_ConcurrentUpdatesRaceSafety(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, _ := mgr.CreateRoom(collaboration.LangGo)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	numClients := 3
	updatesPerClient := 15
	var clients []*websocket.Conn

	for i := 0; i < numClients; i++ {
		c, _, err := websocket.Dial(ctx, fmt.Sprintf("%s?nickname=client%d", wsURL, i), nil)
		if err != nil {
			t.Fatalf("client %d failed to connect: %v", i, err)
		}
		defer c.Close(websocket.StatusNormalClosure, "")
		clients = append(clients, c)
	}

	// Drain initial join messages
	for _, c := range clients {
		go func(conn *websocket.Conn) {
			for {
				_, _, err := conn.Read(ctx)
				if err != nil {
					return
				}
			}
		}(c)
	}

	// Concurrent writes from all clients
	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)
		go func(clientID int, conn *websocket.Conn) {
			defer wg.Done()
			for u := 0; u < updatesPerClient; u++ {
				payload := []byte(fmt.Sprintf("client-%d-update-%d", clientID, u))
				frame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", payload)
				if err := conn.Write(ctx, websocket.MessageBinary, frame); err != nil {
					return
				}
				time.Sleep(2 * time.Millisecond)
			}
		}(i, c)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Verify all updates are recorded in UpdateLog without data corruption
	expectedTotal := numClients * updatesPerClient
	if r.UpdateLog().Count() != expectedTotal {
		t.Errorf("expected %d updates recorded in log, got %d", expectedTotal, r.UpdateLog().Count())
	}
}

func TestIntegration_ReconnectAndStateRestoration(t *testing.T) {
	server, mgr, _ := setupTestServer(t)

	r, _ := mgr.CreateRoom(collaboration.LangTypeScript)
	wsURL := fmt.Sprintf("%s/api/v1/rooms/%s/ws", server.URL, r.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Initial connection sends update
	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("c1 connect failed: %v", err)
	}
	_, _, _ = c1.Read(ctx) // handshake

	frame := ws.EncodeBinaryFrame(ws.BinarySubtypeCRDTUpdate, "default", []byte("persisted-crdt-state"))
	_ = c1.Write(ctx, websocket.MessageBinary, frame)

	time.Sleep(50 * time.Millisecond)

	// Disconnect c1
	c1.Close(websocket.StatusNormalClosure, "disconnecting")

	time.Sleep(50 * time.Millisecond)

	// 2. Client reconnects
	cReconnect, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("reconnect failed: %v", err)
	}
	defer cReconnect.Close(websocket.StatusNormalClosure, "")

	// Read handshake
	_, rJoined, err := cReconnect.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read handshake after reconnect: %v", err)
	}
	var joinedEnv ws.Envelope
	_ = json.Unmarshal(rJoined, &joinedEnv)
	if joinedEnv.Type != ws.TypeRoomJoined {
		t.Fatalf("expected room_joined on reconnect, got: %s", string(rJoined))
	}

	// Read replayed state
	_, updateBytes, err := cReconnect.Read(ctx)
	if err != nil {
		t.Fatalf("failed to receive replayed update on reconnect: %v", err)
	}

	_, _, updatePayload, _ := ws.DecodeBinaryFrame(updateBytes)
	if string(updatePayload) != "persisted-crdt-state" {
		t.Errorf("expected replayed state 'persisted-crdt-state', got %s", string(updatePayload))
	}
}
