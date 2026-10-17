package websocket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/relay"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
)

// Handler manages WebSocket upgrades and lifecycle orchestration.
type Handler struct {
	cfg     *config.Config
	roomMgr *room.Manager
	relay   relay.Relay
}

// NewHandler creates a WebSocket endpoint handler.
func NewHandler(cfg *config.Config, roomMgr *room.Manager, rel relay.Relay) *Handler {
	return &Handler{
		cfg:     cfg,
		roomMgr: roomMgr,
		relay:   rel,
	}
}

// Upgrade handles GET /api/v1/rooms/:roomID/ws requests.
func (h *Handler) Upgrade(c *gin.Context) {
	roomID := c.Param("roomID")

	// Validate room ID format
	if err := room.ValidateRoomID(roomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id format"})
		return
	}

	// Validate origin header
	origin := c.GetHeader("Origin")
	if !security.ValidateOrigin(origin, h.cfg.AllowedOrigins) {
		c.JSON(http.StatusForbidden, gin.H{"error": "cross-origin request forbidden"})
		return
	}

	// Verify room exists in memory
	targetRoom, err := h.roomMgr.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found or expired"})
		return
	}

	// Check room participant capacity before upgrading connection
	if targetRoom.ParticipantCount() >= targetRoom.MaxParticipants() {
		c.JSON(http.StatusForbidden, gin.H{"error": "room participant capacity reached"})
		return
	}

	// Configure WebSocket accept options
	acceptOpts := &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	}
	isWildcard := false
	for _, o := range h.cfg.AllowedOrigins {
		if o == "*" {
			isWildcard = true
			break
		}
	}
	if isWildcard {
		acceptOpts.InsecureSkipVerify = true
	} else {
		acceptOpts.OriginPatterns = h.cfg.AllowedOrigins
	}

	conn, err := websocket.Accept(c.Writer, c.Request, acceptOpts)
	if err != nil {
		log.Printf("[ws] handshake accept error: %v", err)
		return
	}

	// Generate ephemeral server-assigned participant ID
	participantID, err := room.GenerateParticipantID()
	if err != nil {
		_ = conn.Close(websocket.StatusInternalError, "failed to assign participant ID")
		return
	}

	nickname := c.Query("nickname")
	sinceSeqStr := c.Query("since_seq")
	var sinceSeq uint64
	if sinceSeqStr != "" {
		if val, err := strconv.ParseUint(sinceSeqStr, 10, 64); err == nil {
			sinceSeq = val
		}
	}

	participant := room.NewParticipant(participantID, roomID, nickname, 256)

	if err := targetRoom.AddParticipant(participant); err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, err.Error())
		return
	}

	sub, err := h.relay.Subscribe(c.Request.Context(), roomID, participantID)
	if err != nil {
		targetRoom.RemoveParticipant(participantID)
		_ = conn.Close(websocket.StatusInternalError, "failed to subscribe to relay")
		return
	}

	wsConn := NewConnection(c.Request.Context(), conn, participant, targetRoom, h.relay, sub, h.cfg, sinceSeq)
	wsConn.Serve()
}
