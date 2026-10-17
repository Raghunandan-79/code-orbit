package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/syncraft/backend/internal/config"
	"github.com/syncraft/backend/internal/http/handlers"
	"github.com/syncraft/backend/internal/http/routes"
	"github.com/syncraft/backend/internal/lifecycle"
	"github.com/syncraft/backend/internal/relay"
	"github.com/syncraft/backend/internal/room"
	"github.com/syncraft/backend/internal/security"
	"github.com/syncraft/backend/internal/websocket"
)

func main() {
	log.Println("Starting Syncraft Anonymous Real-Time Collaboration Backend...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// 2. Initialize Core Components
	roomMgr := room.NewManager(room.ManagerConfig{
		MaxRooms:                  cfg.MaxRooms,
		MaxParticipantsPerRoom:    cfg.MaxParticipantsPerRoom,
		MaxBufferedUpdatesPerRoom: cfg.MaxBufferedUpdatesPerRoom,
		RoomIdleExpiration:        cfg.RoomIdleExpiration,
	})

	localRelay := relay.NewLocalRelay(256)
	ipLimiter := security.NewIPRateLimiter(cfg.HTTPRateLimitRPM, cfg.HTTPRateLimitBurst)

	// 3. Start Lifecycle Background Cleaner
	cleaner := lifecycle.NewCleaner(roomMgr, ipLimiter, cfg.RoomCleanupInterval)
	cleaner.Start()

	// 4. Initialize Handlers and Routing
	healthH := handlers.NewHealthHandler(roomMgr)
	roomH := handlers.NewRoomHandler(cfg, roomMgr)
	wsH := websocket.NewHandler(cfg, roomMgr, localRelay)

	router := routes.SetupRouter(cfg, healthH, roomH, wsH, ipLimiter)

	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 5. Start Server in Goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Syncraft HTTP & WebSocket server listening on %s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 6. Graceful Shutdown on OS Signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("Server startup failed: %v", err)
	case sig := <-quit:
		log.Printf("Received signal %s. Initiating graceful shutdown...", sig)
	}

	// Step A: Mark unready for load balancers
	healthH.SetReady(false)

	// Step B: Graceful HTTP shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Step C: Stop cleaner
	if err := cleaner.Stop(shutdownCtx); err != nil {
		log.Printf("Cleaner stop error: %v", err)
	}

	// Step D: Close rooms and relay
	roomMgr.Close()
	_ = localRelay.Close()

	log.Println("Syncraft backend stopped gracefully. Bye.")
}
