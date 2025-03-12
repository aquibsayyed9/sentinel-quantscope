// internal/server/routes/portfolio_routes.go
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/aquibsayyed9/sentinel/internal/handlers"
)

// SetupPortfolioRoutes sets up all portfolio-related routes
func SetupPortfolioRoutes(router *gin.RouterGroup, portfolioHandler *handlers.PortfolioHandler) {
	portfolio := router.Group("/portfolio")
	{
		portfolio.POST("", portfolioHandler.CreatePortfolio)
		portfolio.GET("", portfolioHandler.GetPortfolio)
		portfolio.GET("/holdings", portfolioHandler.GetHoldings)
		portfolio.POST("/holdings", portfolioHandler.AddHolding)
		portfolio.DELETE("/holdings/:symbol", portfolioHandler.RemoveHolding)
	}
}
