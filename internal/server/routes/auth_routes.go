// internal/server/routes/auth_routes.go
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/aquibsayyed9/sentinel/internal/handlers"
)

// SetupAuthRoutes sets up all auth-related routes
func SetupAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}
