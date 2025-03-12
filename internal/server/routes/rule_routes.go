// internal/server/routes/rule_routes.go
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/aquibsayyed9/sentinel/internal/handlers"
)

// SetupRuleRoutes sets up all rule-related routes
func SetupRuleRoutes(router *gin.RouterGroup, ruleHandler *handlers.RuleHandler) {
	rules := router.Group("/rules")
	{
		rules.POST("", ruleHandler.CreateRule)
		rules.GET("", ruleHandler.GetRules)
		rules.GET("/:id", ruleHandler.GetRule)
		rules.PUT("/:id/activate", ruleHandler.ActivateRule)
		rules.PUT("/:id/deactivate", ruleHandler.DeactivateRule)
		rules.DELETE("/:id", ruleHandler.DeleteRule)
	}
}
