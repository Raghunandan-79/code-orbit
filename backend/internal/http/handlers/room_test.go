package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/collaboration"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/http/middleware"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
)

func setupTestRouter() (*gin.Engine, *room.Manager, *config.Config) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Host:                      "127.0.0.1",
		Port:                      "8080",
		AllowedOrigins:            []string{"*"},
		MaxRooms:                  10,
		MaxParticipantsPerRoom:    5,
		MaxBufferedUpdatesPerRoom: 50,
		RoomIdleExpiration:        time.Hour,
	}

	mgr := room.NewManager(room.ManagerConfig{
		MaxRooms:                  cfg.MaxRooms,
		MaxParticipantsPerRoom:    cfg.MaxParticipantsPerRoom,
		MaxBufferedUpdatesPerRoom: cfg.MaxBufferedUpdatesPerRoom,
		RoomIdleExpiration:        cfg.RoomIdleExpiration,
	})

	ipLimiter := security.NewIPRateLimiter(60, 20)

	healthH := NewHealthHandler(mgr)
	roomH := NewRoomHandler(cfg, mgr)

	r.GET("/healthz", healthH.Healthz)
	r.GET("/readyz", healthH.Readyz)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimiter(ipLimiter))
	{
		v1.POST("/rooms", roomH.CreateRoom)
		v1.GET("/rooms/:roomID", roomH.GetRoom)
	}

	return r, mgr, cfg
}

func TestHealthEndpoints(t *testing.T) {
	router, _, _ := setupTestRouter()

	// GET /healthz
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /healthz, got %d", w.Code)
	}

	var healthResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to parse /healthz response: %v", err)
	}
	if healthResp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", healthResp["status"])
	}

	// GET /readyz
	wReady := httptest.NewRecorder()
	reqReady, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(wReady, reqReady)

	if wReady.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /readyz, got %d", wReady.Code)
	}
}

func TestCreateAndGetRoom(t *testing.T) {
	router, _, _ := setupTestRouter()

	// 1. Create Room with Python
	body, _ := json.Marshal(map[string]string{"language": "python"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d (body: %s)", w.Code, w.Body.String())
	}

	var createResp CreateRoomResponse
	if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createResp.RoomID == "" {
		t.Fatalf("expected non-empty room ID")
	}
	if createResp.Language != collaboration.LangPython {
		t.Errorf("expected python language, got %s", createResp.Language)
	}
	if createResp.DocumentTitle != "main.py" {
		t.Errorf("expected main.py document title, got %s", createResp.DocumentTitle)
	}

	// 2. Get Room Metadata
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/rooms/"+createResp.RoomID, nil)
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wGet.Code)
	}

	var metaResp RoomMetadataResponse
	if err := json.Unmarshal(wGet.Body.Bytes(), &metaResp); err != nil {
		t.Fatalf("failed to decode meta response: %v", err)
	}

	if metaResp.RoomID != createResp.RoomID {
		t.Errorf("expected room ID %s, got %s", createResp.RoomID, metaResp.RoomID)
	}
	if metaResp.ParticipantCount != 0 {
		t.Errorf("expected 0 participants, got %d", metaResp.ParticipantCount)
	}
}

func TestCreateRoomInvalidLanguage(t *testing.T) {
	router, _, _ := setupTestRouter()

	body, _ := json.Marshal(map[string]string{"language": "cobol"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unsupported language, got %d", w.Code)
	}
}

func TestGetNonexistentRoom(t *testing.T) {
	router, _, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/rooms/random-not-found-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", w.Code)
	}
}
