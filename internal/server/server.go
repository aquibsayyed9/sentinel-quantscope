// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aquibsayyed9/sentinel/internal/auth"
	"github.com/aquibsayyed9/sentinel/internal/config"
	"github.com/aquibsayyed9/sentinel/internal/handlers"
	"github.com/aquibsayyed9/sentinel/internal/server/routes"
)

// Server represents the HTTP server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	config     *config.Config
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, authHandler *handlers.AuthHandler, ruleHandler *handlers.RuleHandler,
	portfolioHandler *handlers.PortfolioHandler, tokenService auth.TokenService) *Server {

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup routes
	routes.Setup(router, authHandler, ruleHandler, portfolioHandler, tokenService)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		config:     cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
