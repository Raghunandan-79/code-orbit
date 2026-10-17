package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/relay"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
)

// Connection encapsulates an active WebSocket session for a participant in a room.
type Connection struct {
	conn         *websocket.Conn
	participant  *room.Participant
	room         *room.Room
	relay        relay.Relay
	subscription relay.Subscription
	cfg          *config.Config
	msgLimiter   *security.TokenBucket
	sinceSeq     uint64

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// NewConnection creates a new Connection instance.
func NewConnection(
	ctx context.Context,
	conn *websocket.Conn,
	p *room.Participant,
	r *room.Room,
	rel relay.Relay,
	sub relay.Subscription,
	cfg *config.Config,
	sinceSeq uint64,
) *Connection {
	cCtx, cancel := context.WithCancel(ctx)
	return &Connection{
		conn:         conn,
		participant:  p,
		room:         r,
		relay:        rel,
		subscription: sub,
		cfg:          cfg,
		msgLimiter:   security.NewTokenBucket(float64(cfg.WSMessageRateLimit), float64(cfg.WSMessageRateLimit)),
		sinceSeq:     sinceSeq,
		ctx:          cCtx,
		cancel:       cancel,
	}
}

// Serve handles the read, write, and relay pump lifecycles until disconnection.
func (c *Connection) Serve() {
	defer c.cleanup("disconnected")

	// Apply read limit
	c.conn.SetReadLimit(c.cfg.MaxWebSocketMessageSize)

	// Send initial handshake messages to client
	if err := c.sendInitialHandshake(); err != nil {
		log.Printf("[ws] initial handshake failed for %s: %v", c.participant.ID, err)
		return
	}

	// Notify other participants of new participant
	c.broadcastParticipantJoined()

	var wg sync.WaitGroup
	wg.Add(3)

	// Relay pump: receives messages from relay and queues them to participant
	go func() {
		defer wg.Done()
		c.relayPump()
	}()

	// Write pump: writes queued messages to WebSocket
	go func() {
		defer wg.Done()
		c.writePump()
	}()

	// Read pump: reads messages from WebSocket and processes them
	go func() {
		defer wg.Done()
		c.readPump()
	}()

	wg.Wait()
}

// sendInitialHandshake sends room_joined and existing update history to the joining client.
func (c *Connection) sendInitialHandshake() error {
	latestSeq := c.room.UpdateLog().LatestSeq()
	oldestSeq := c.room.UpdateLog().OldestSeq()

	joinedPayload := RoomJoinedPayload{
		RoomID:          c.room.ID(),
		ParticipantID:   c.participant.ID,
		Language:        c.room.DocumentMetadata().Language,
		DocumentTitle:   c.room.DocumentMetadata().Title,
		CreatedAt:       c.room.CreatedAt(),
		MaxParticipants: c.room.MaxParticipants(),
		Participants:    c.room.ParticipantSnapshots(),
		LatestSeq:       latestSeq,
		OldestSeq:       oldestSeq,
	}

	joinedEnv, err := NewEnvelope(TypeRoomJoined, c.room.ID(), c.participant.ID, "default", joinedPayload)
	if err != nil {
		return err
	}

	b, err := json.Marshal(joinedEnv)
	if err != nil {
		return err
	}

	if !c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: b}) {
		return errors.New("failed to enqueue room_joined envelope")
	}

	// Replay updates: only updates since c.sinceSeq (if 0, replays all)
	updates := c.room.UpdateLog().GetSince(c.sinceSeq)
	for _, entry := range updates {
		if entry.IsBinary {
			framed := EncodeBinaryFrame(BinarySubtypeCRDTUpdate, entry.DocID, entry.Data)
			c.participant.Enqueue(room.OutboundMessage{IsBinary: true, Data: framed})
		} else {
			syncPayload := SyncUpdatePayload{
				Seq:   entry.Seq,
				DocID: entry.DocID,
				Data:  entry.Data,
			}
			syncEnv, err := NewEnvelope(TypeSyncUpdate, c.room.ID(), entry.SenderID, entry.DocID, syncPayload)
			if err == nil {
				data, _ := json.Marshal(syncEnv)
				c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: data})
			}
		}
	}

	return nil
}

func (c *Connection) broadcastParticipantJoined() {
	payload := ParticipantJoinedPayload{
		ParticipantID: c.participant.ID,
		Nickname:      c.participant.Nickname,
		JoinedAt:      c.participant.JoinedAt,
	}

	env, err := NewEnvelope(TypeParticipantJoined, c.room.ID(), c.participant.ID, "default", payload)
	if err != nil {
		return
	}

	b, err := json.Marshal(env)
	if err != nil {
		return
	}

	_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
		RoomID:        c.room.ID(),
		SenderID:      c.participant.ID,
		ExcludeSender: true,
		IsBinary:      false,
		Data:          b,
	})
}

// relayPump forwards room broadcast messages from the relay subscription to the participant's outbound queue.
func (c *Connection) relayPump() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.subscription.Channel():
			if !ok {
				return
			}
			enqueued := c.participant.Enqueue(room.OutboundMessage{
				IsBinary: msg.IsBinary,
				Data:     msg.Data,
			})
			if !enqueued {
				// Slow consumer detected
				if c.participant.QueueDroppedCount() > 20 {
					log.Printf("[ws] terminating slow consumer %s (room %s): dropped %d messages",
						c.participant.ID, c.room.ID(), c.participant.QueueDroppedCount())
					c.cancel()
					return
				}
			}
		}
	}
}

// writePump handles all writes to the WebSocket connection.
func (c *Connection) writePump() {
	pingInterval := c.cfg.WSPingInterval
	if pingInterval <= 0 {
		pingInterval = 15 * time.Second
	}
	pingTimeout := c.cfg.WSPingTimeout
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}

	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-pingTicker.C:
			writeCtx, cancel := context.WithTimeout(c.ctx, pingTimeout)
			err := c.conn.Ping(writeCtx)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		case msg, ok := <-c.participant.OutboundChannel():
			if !ok {
				return
			}

			writeCtx, cancel := context.WithTimeout(c.ctx, c.cfg.WriteTimeout)
			var msgType websocket.MessageType
			if msg.IsBinary {
				msgType = websocket.MessageBinary
			} else {
				msgType = websocket.MessageText
			}

			err := c.conn.Write(writeCtx, msgType, msg.Data)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		}
	}
}

// readPump reads messages from the WebSocket connection and dispatches them.
func (c *Connection) readPump() {
	defer c.cancel()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		msgType, data, err := c.conn.Read(c.ctx)
		if err != nil {
			return
		}

		c.participant.Touch()
		c.room.Touch()

		// Enforce per-connection rate limit
		if !c.msgLimiter.Allow() {
			c.sendErrorEnvelope("RATE_LIMITED", "too many messages sent, please slow down")
			continue
		}

		switch msgType {
		case websocket.MessageBinary:
			c.handleBinaryMessage(data)
		case websocket.MessageText:
			c.handleTextMessage(data)
		}
	}
}

func (c *Connection) handleBinaryMessage(raw []byte) {
	subtype, docID, payload, err := DecodeBinaryFrame(raw)
	if err != nil {
		c.sendErrorEnvelope("INVALID_BINARY_FRAME", "failed to decode binary frame")
		return
	}

	switch subtype {
	case BinarySubtypeCRDTUpdate:
		// Store update in room update log
		entry, err := c.room.UpdateLog().Append(docID, c.participant.ID, payload, true)
		if err != nil {
			if errors.Is(err, collaboration.ErrUpdateLogCapacityExceeded) {
				c.sendErrorEnvelope("UPDATE_LIMIT_EXCEEDED", "room update log capacity exceeded")
			} else {
				c.sendErrorEnvelope("UPDATE_ERROR", "failed to process CRDT update")
			}
			return
		}

		// Broadcast to other participants as framed binary
		framed := EncodeBinaryFrame(BinarySubtypeCRDTUpdate, entry.DocID, entry.Data)
		_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
			RoomID:        c.room.ID(),
			SenderID:      c.participant.ID,
			ExcludeSender: true,
			IsBinary:      true,
			Data:          framed,
		})

	case BinarySubtypeAwareness:
		// Broadcast transient awareness (cursor, selection) to room peers
		framed := EncodeBinaryFrame(BinarySubtypeAwareness, docID, payload)
		_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
			RoomID:        c.room.ID(),
			SenderID:      c.participant.ID,
			ExcludeSender: true,
			IsBinary:      true,
			Data:          framed,
		})

	default:
		// Forward opaque binary frame to room peers
		_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
			RoomID:        c.room.ID(),
			SenderID:      c.participant.ID,
			ExcludeSender: true,
			IsBinary:      true,
			Data:          raw,
		})
	}
}

func (c *Connection) handleTextMessage(raw []byte) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		c.sendErrorEnvelope("INVALID_JSON", "malformed JSON envelope")
		return
	}

	if err := env.Validate(); err != nil {
		c.sendErrorEnvelope("INVALID_ENVELOPE", err.Error())
		return
	}

	switch env.Type {
	case TypePing:
		pongEnv, _ := NewEnvelope(TypePong, c.room.ID(), c.participant.ID, "default", nil)
		if b, err := json.Marshal(pongEnv); err == nil {
			c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: b})
		}

	case TypePong:
		c.participant.Touch()

	case TypeSyncRequest:
		var req SyncRequestPayload
		if len(env.Payload) > 0 {
			_ = json.Unmarshal(env.Payload, &req)
		}

		updates := c.room.UpdateLog().GetSince(req.SinceSeq)
		for _, entry := range updates {
			if entry.IsBinary {
				framed := EncodeBinaryFrame(BinarySubtypeCRDTUpdate, entry.DocID, entry.Data)
				c.participant.Enqueue(room.OutboundMessage{IsBinary: true, Data: framed})
			} else {
				syncPayload := SyncUpdatePayload{
					Seq:   entry.Seq,
					DocID: entry.DocID,
					Data:  entry.Data,
				}
				syncEnv, err := NewEnvelope(TypeSyncUpdate, c.room.ID(), entry.SenderID, entry.DocID, syncPayload)
				if err == nil {
					data, _ := json.Marshal(syncEnv)
					c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: data})
				}
			}
		}

	case TypeSyncUpdate:
		var syncPayload SyncUpdatePayload
		if err := json.Unmarshal(env.Payload, &syncPayload); err != nil {
			c.sendErrorEnvelope("INVALID_PAYLOAD", "invalid sync_update payload")
			return
		}

		docID := env.DocID
		if docID == "" {
			docID = "default"
		}

		entry, err := c.room.UpdateLog().Append(docID, c.participant.ID, syncPayload.Data, false)
		if err != nil {
			if errors.Is(err, collaboration.ErrUpdateLogCapacityExceeded) {
				c.sendErrorEnvelope("UPDATE_LIMIT_EXCEEDED", "room update log capacity exceeded")
			} else {
				c.sendErrorEnvelope("UPDATE_ERROR", "failed to record sync update")
			}
			return
		}

		// Broadcast sync_update to other participants
		broadcastPayload := SyncUpdatePayload{
			Seq:   entry.Seq,
			DocID: entry.DocID,
			Data:  entry.Data,
		}
		broadcastEnv, err := NewEnvelope(TypeSyncUpdate, c.room.ID(), c.participant.ID, entry.DocID, broadcastPayload)
		if err == nil {
			data, _ := json.Marshal(broadcastEnv)
			_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
				RoomID:        c.room.ID(),
				SenderID:      c.participant.ID,
				ExcludeSender: true,
				IsBinary:      false,
				Data:          data,
			})
		}

	case TypeAwarenessUpdate:
		// Forward awareness update to peers, overriding participant and room IDs for security
		env.ParticipantID = c.participant.ID
		env.RoomID = c.room.ID()
		env.Timestamp = time.Now().UTC().UnixMilli()

		data, err := json.Marshal(env)
		if err == nil {
			_ = c.relay.Publish(c.ctx, &relay.BroadcastMessage{
				RoomID:        c.room.ID(),
				SenderID:      c.participant.ID,
				ExcludeSender: true,
				IsBinary:      false,
				Data:          data,
			})
		}

	case TypeRoomState:
		statePayload := RoomStatePayload{
			RoomID:           c.room.ID(),
			Language:         c.room.DocumentMetadata().Language,
			ParticipantCount: c.room.ParticipantCount(),
			MaxParticipants:  c.room.MaxParticipants(),
			Participants:     c.room.ParticipantSnapshots(),
		}
		stateEnv, err := NewEnvelope(TypeRoomState, c.room.ID(), c.participant.ID, "default", statePayload)
		if err == nil {
			data, _ := json.Marshal(stateEnv)
			c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: data})
		}

	default:
		c.sendErrorEnvelope("UNSUPPORTED_MESSAGE_TYPE", "unsupported message type")
	}
}

func (c *Connection) sendErrorEnvelope(code, message string) {
	errEnv, err := NewEnvelope(TypeError, c.room.ID(), c.participant.ID, "default", ErrorPayload{
		Code:    code,
		Message: message,
	})
	if err != nil {
		return
	}

	b, err := json.Marshal(errEnv)
	if err == nil {
		c.participant.Enqueue(room.OutboundMessage{IsBinary: false, Data: b})
	}
}

// cleanup ensures graceful resource deallocation and broadcast of participant_left.
func (c *Connection) cleanup(reason string) {
	c.closeOnce.Do(func() {
		c.cancel()

		// Remove from room
		c.room.RemoveParticipant(c.participant.ID)

		// Unsubscribe from relay
		if c.subscription != nil {
			c.subscription.Unsubscribe()
		}

		// Close participant queue
		c.participant.Close()

		// Broadcast participant_left to remaining room members
		leftPayload := ParticipantLeftPayload{
			ParticipantID: c.participant.ID,
			Reason:        reason,
		}
		if env, err := NewEnvelope(TypeParticipantLeft, c.room.ID(), c.participant.ID, "default", leftPayload); err == nil {
			if b, err := json.Marshal(env); err == nil {
				_ = c.relay.Publish(context.Background(), &relay.BroadcastMessage{
					RoomID:        c.room.ID(),
					SenderID:      c.participant.ID,
					ExcludeSender: true,
					IsBinary:      false,
					Data:          b,
				})
			}
		}

		// Close WebSocket connection cleanly
		_ = c.conn.Close(websocket.StatusNormalClosure, "session closed")
	})
}
