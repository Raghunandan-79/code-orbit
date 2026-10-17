package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/http/handlers"
	"github.com/syncraft/backend/internal/http/middleware"
	"github.com/syncraft/backend/internal/security"
	"github.com/syncraft/backend/internal/websocket"
)

// SetupRouter initializes the Gin engine with all routes and middleware.
func SetupRouter(
	cfg *config.Config,
	healthH *handlers.HealthHandler,
	roomH *handlers.RoomHandler,
	wsH *websocket.Handler,
	ipLimiter *security.IPRateLimiter,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Global Middlewares
	r.Use(middleware.SafeRecovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health and Readiness probes (unthrottled)
	r.GET("/healthz", healthH.Healthz)
	r.GET("/readyz", healthH.Readyz)

	// API v1 group with IP rate limiting
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimiter(ipLimiter))
	{
		rooms := v1.Group("/rooms")
		{
			rooms.POST("", roomH.CreateRoom)
			rooms.GET("/:roomID", roomH.GetRoom)
			rooms.GET("/:roomID/ws", wsH.Upgrade)
		}
	}

	return r
}
