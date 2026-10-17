package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/room"
)

// RoomHandler provides HTTP endpoints for room creation and metadata lookup.
type RoomHandler struct {
	cfg     *config.Config
	roomMgr *room.Manager
}

// NewRoomHandler creates a RoomHandler.
func NewRoomHandler(cfg *config.Config, roomMgr *room.Manager) *RoomHandler {
	return &RoomHandler{
		cfg:     cfg,
		roomMgr: roomMgr,
	}
}

type CreateRoomRequest struct {
	Language string `json:"language"`
}

type CreateRoomResponse struct {
	RoomID          string                          `json:"room_id"`
	WSURL           string                          `json:"ws_url"`
	Language        collaboration.SupportedLanguage `json:"language"`
	DocumentTitle   string                          `json:"document_title"`
	CreatedAt       time.Time                       `json:"created_at"`
	ExpiresAt       time.Time                       `json:"expires_at"`
	MaxParticipants int                             `json:"max_participants"`
}

type RoomMetadataResponse struct {
	RoomID           string                          `json:"room_id"`
	Language         collaboration.SupportedLanguage `json:"language"`
	DocumentTitle    string                          `json:"document_title"`
	ParticipantCount int                             `json:"participant_count"`
	MaxParticipants  int                             `json:"max_participants"`
	CreatedAt        time.Time                       `json:"created_at"`
	LastActiveAt     time.Time                       `json:"last_active_at"`
	UpdateCount      int                             `json:"update_count"`
}

// CreateRoom handles POST /api/v1/rooms.
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "malformed request payload",
			})
			return
		}
	}

	// Validate language
	lang, err := collaboration.NormalizeLanguage(req.Language)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create room in manager
	r, err := h.roomMgr.CreateRoom(lang)
	if err != nil {
		if errors.Is(err, room.ErrMaxRoomsReached) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "maximum room capacity reached, please try again later",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create room",
		})
		return
	}

	resp := CreateRoomResponse{
		RoomID:          r.ID(),
		WSURL:           "/api/v1/rooms/" + r.ID() + "/ws",
		Language:        lang,
		DocumentTitle:   r.DocumentMetadata().Title,
		CreatedAt:       r.CreatedAt(),
		ExpiresAt:       r.CreatedAt().Add(h.cfg.RoomIdleExpiration),
		MaxParticipants: r.MaxParticipants(),
	}

	c.JSON(http.StatusCreated, resp)
}

// GetRoom handles GET /api/v1/rooms/:roomID.
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("roomID")

	if err := room.ValidateRoomID(roomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid room id format",
		})
		return
	}

	snap, err := h.roomMgr.GetRoomSnapshot(roomID)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "room not found or expired",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error retrieving room",
		})
		return
	}

	resp := RoomMetadataResponse{
		RoomID:           snap.ID,
		Language:         snap.Language,
		DocumentTitle:    snap.DocumentTitle,
		ParticipantCount: snap.ParticipantCount,
		MaxParticipants:  snap.MaxParticipants,
		CreatedAt:        snap.CreatedAt,
		LastActiveAt:     snap.LastActiveAt,
		UpdateCount:      snap.UpdateCount,
	}

	c.JSON(http.StatusOK, resp)
}
