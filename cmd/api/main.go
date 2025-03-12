// cmd/api/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/aquibsayyed9/sentinel/internal/auth"
	"github.com/aquibsayyed9/sentinel/internal/config"
	"github.com/aquibsayyed9/sentinel/internal/db"
	"github.com/aquibsayyed9/sentinel/internal/handlers"
	"github.com/aquibsayyed9/sentinel/internal/repository"
	"github.com/aquibsayyed9/sentinel/internal/server"
	"github.com/aquibsayyed9/sentinel/internal/services"
	"github.com/aquibsayyed9/sentinel/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	l := logger.NewLogger(cfg.Environment)
	defer l.Sync()

	l.Info("Starting Sentinel API server", zap.String("environment", cfg.Environment))

	// Connect to database
	database, err := db.Connect(cfg)
	if err != nil {
		l.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	ruleRepo := repository.NewRuleRepository(database)
	// executionRepo := repository.NewExecutionRepository(database)
	portfolioRepo := repository.NewPortfolioRepository(database)

	// Initialize token service
	tokenService := auth.NewTokenService(cfg)

	// Initialize services
	userService := services.NewUserService(userRepo)
	ruleService := services.NewRuleService(ruleRepo)
	portfolioService := services.NewPortfolioService(portfolioRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, tokenService)
	ruleHandler := handlers.NewRuleHandler(ruleService)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)

	// Create HTTP server
	srv := server.NewServer(cfg, authHandler, ruleHandler, portfolioHandler, tokenService)

	// Start server in a goroutine
	go func() {
		l.Info("Starting HTTP server", zap.String("address", cfg.Server.Host+":"+cfg.Server.Port))
		if err := srv.Start(); err != nil {
			l.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		l.Fatal("Server forced to shutdown", zap.Error(err))
	}

	l.Info("Server exited gracefully")
}
