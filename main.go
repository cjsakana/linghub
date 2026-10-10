package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"linghub/internal/config"
	"linghub/internal/database"
	"linghub/internal/jwtutil"
	"linghub/internal/router"
	"linghub/internal/service"
)

func main() {
	// Load environment variables from .env file (if it exists)
	_ = godotenv.Load()

	// Load configuration from environment
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Create token manager
	tm := jwtutil.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)

	// Create services
	authService := service.NewAuthService(db, tm)
	chatService, err := service.NewChatService(db, cfg)
	if err != nil {
		log.Fatalf("failed to create chat service: %v", err)
	}

	// Setup router
	r := router.SetupRouter(authService, chatService, tm)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
