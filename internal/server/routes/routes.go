// internal/server/routes/routes.go
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/aquibsayyed9/sentinel/internal/auth"
	"github.com/aquibsayyed9/sentinel/internal/handlers"
)

// Setup sets up all routes for the API
func Setup(router *gin.Engine, authHandler *handlers.AuthHandler, ruleHandler *handlers.RuleHandler,
	portfolioHandler *handlers.PortfolioHandler, tokenService auth.TokenService) {

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 group
	v1 := router.Group("/api/v1")

	// Auth routes
	SetupAuthRoutes(v1, authHandler)

	// Protected routes
	protected := v1.Group("")
	protected.Use(auth.AuthMiddleware(tokenService))
	{
		// Rule routes
		SetupRuleRoutes(protected, ruleHandler)

		// Portfolio routes
		SetupPortfolioRoutes(protected, portfolioHandler)

		// User profile
		protected.GET("/profile", authHandler.GetProfile)
	}
}
