package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadDefaults(t *testing.T) {
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading default config, got: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default Port 8080, got %s", cfg.Port)
	}
	if cfg.MaxRooms != 1000 {
		t.Errorf("expected default MaxRooms 1000, got %d", cfg.MaxRooms)
	}
	if cfg.MaxParticipantsPerRoom != 25 {
		t.Errorf("expected default MaxParticipantsPerRoom 25, got %d", cfg.MaxParticipantsPerRoom)
	}
	if cfg.MaxWebSocketMessageSize != 512*1024 {
		t.Errorf("expected default MaxWebSocketMessageSize 512KB, got %d", cfg.MaxWebSocketMessageSize)
	}
	if cfg.RoomIdleExpiration != 24*time.Hour {
		t.Errorf("expected default RoomIdleExpiration 24h, got %v", cfg.RoomIdleExpiration)
	}
}

func TestConfigCustomEnv(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("MAX_ROOMS", "50")
	os.Setenv("ALLOWED_ORIGINS", "https://syncraft.dev, https://app.syncraft.dev")
	os.Setenv("ROOM_IDLE_EXPIRATION", "2h")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.MaxRooms != 50 {
		t.Errorf("expected MaxRooms 50, got %d", cfg.MaxRooms)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "https://syncraft.dev" {
		t.Errorf("expected parsed origins, got %+v", cfg.AllowedOrigins)
	}
	if cfg.RoomIdleExpiration != 2*time.Hour {
		t.Errorf("expected 2h idle expiration, got %v", cfg.RoomIdleExpiration)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		Port:                      "",
		MaxRooms:                  10,
		MaxParticipantsPerRoom:    5,
		MaxWebSocketMessageSize:   1024,
		MaxBufferedUpdatesPerRoom: 100,
		RoomIdleExpiration:        time.Hour,
	}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error when port is empty")
	}

	cfg.Port = "8080"
	cfg.MaxRooms = 0
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error when max rooms <= 0")
	}
}
