package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all server configuration values.
type Config struct {
	// Server
	Host                    string        `json:"host"`
	Port                    string        `json:"port"`
	ReadTimeout             time.Duration `json:"read_timeout"`
	WriteTimeout            time.Duration `json:"write_timeout"`
	GracefulShutdownTimeout time.Duration `json:"graceful_shutdown_timeout"`

	// CORS and WebSocket Origins
	AllowedOrigins []string `json:"allowed_origins"`

	// Capacity & Lifecycle Limits
	MaxRooms                  int           `json:"max_rooms"`
	MaxParticipantsPerRoom    int           `json:"max_participants_per_room"`
	MaxWebSocketMessageSize   int64         `json:"max_websocket_message_size"`
	MaxBufferedUpdatesPerRoom int           `json:"max_buffered_updates_per_room"`
	RoomIdleExpiration        time.Duration `json:"room_idle_expiration"`
	RoomCleanupInterval       time.Duration `json:"room_cleanup_interval"`

	// Heartbeat & Dead Peer Detection
	WSPingInterval time.Duration `json:"ws_ping_interval"`
	WSPingTimeout  time.Duration `json:"ws_ping_timeout"`

	// Rate Limiting
	HTTPRateLimitRPM   int `json:"http_rate_limit_rpm"` // requests per minute per IP
	HTTPRateLimitBurst int `json:"http_rate_limit_burst"`
	WSMessageRateLimit int `json:"ws_message_rate_limit"` // msgs per second per connection
}

// Load reads configuration from environment variables with safe defaults.
func Load() (*Config, error) {
	port := getEnv("PORT", "")
	if port == "" {
		port = getEnv("HTTP_PORT", "8080")
	}

	cfg := &Config{
		Host:                    getEnv("HTTP_HOST", "0.0.0.0"),
		Port:                    port,
		ReadTimeout:             getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:            getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		GracefulShutdownTimeout: getDurationEnv("GRACEFUL_SHUTDOWN_TIMEOUT", 10*time.Second),

		AllowedOrigins: parseCommaSeparated(getEnv("ALLOWED_ORIGINS", "*")),

		MaxRooms:                  getIntEnv("MAX_ROOMS", 1000),
		MaxParticipantsPerRoom:    getIntEnv("MAX_PARTICIPANTS_PER_ROOM", 25),
		MaxWebSocketMessageSize:   getInt64Env("MAX_WEBSOCKET_MESSAGE_SIZE", 512*1024), // 512 KB
		MaxBufferedUpdatesPerRoom: getIntEnv("MAX_BUFFERED_UPDATES_PER_ROOM", 2000),
		RoomIdleExpiration:        getDurationEnv("ROOM_IDLE_EXPIRATION", 24*time.Hour),
		RoomCleanupInterval:       getDurationEnv("ROOM_CLEANUP_INTERVAL", 10*time.Minute),

		WSPingInterval: getDurationEnv("WS_PING_INTERVAL", 15*time.Second),
		WSPingTimeout:  getDurationEnv("WS_PING_TIMEOUT", 5*time.Second),

		HTTPRateLimitRPM:   getIntEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", 120),
		HTTPRateLimitBurst: getIntEnv("RATE_LIMIT_BURST", 30),
		WSMessageRateLimit: getIntEnv("WS_MESSAGE_RATE_LIMIT", 100),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks for critical misconfigurations.
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if c.MaxRooms <= 0 {
		return fmt.Errorf("max_rooms must be greater than 0")
	}
	if c.MaxParticipantsPerRoom <= 0 {
		return fmt.Errorf("max_participants_per_room must be greater than 0")
	}
	if c.MaxWebSocketMessageSize <= 0 {
		return fmt.Errorf("max_websocket_message_size must be greater than 0")
	}
	if c.MaxBufferedUpdatesPerRoom <= 0 {
		return fmt.Errorf("max_buffered_updates_per_room must be greater than 0")
	}
	if c.RoomIdleExpiration <= 0 {
		return fmt.Errorf("room_idle_expiration must be greater than 0")
	}
	if c.WSPingInterval <= 0 {
		c.WSPingInterval = 15 * time.Second
	}
	if c.WSPingTimeout <= 0 {
		c.WSPingTimeout = 5 * time.Second
	}
	return nil
}

// Address returns host:port formatted string.
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getInt64Env(key string, defaultVal int64) int64 {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	dur, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return dur
}

func parseCommaSeparated(val string) []string {
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}
