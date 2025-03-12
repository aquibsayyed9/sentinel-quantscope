// internal/handlers/rule_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aquibsayyed9/sentinel/internal/services"
)

type RuleHandler struct {
	ruleService services.RuleService
}

func NewRuleHandler(ruleService services.RuleService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

type createRuleRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Symbol      string                   `json:"symbol" binding:"required"`
	RuleType    string                   `json:"rule_type" binding:"required"`
	Conditions  []services.RuleCondition `json:"conditions" binding:"required"`
	Actions     []services.RuleAction    `json:"actions" binding:"required"`
}

type ruleResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Symbol      string                   `json:"symbol"`
	RuleType    string                   `json:"rule_type"`
	Conditions  []services.RuleCondition `json:"conditions"`
	Actions     []services.RuleAction    `json:"actions"`
	Status      string                   `json:"status"`
	IsAIManaged bool                     `json:"is_ai_managed"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rule, err := h.ruleService.CreateRule(
		c.Request.Context(),
		userID.(uuid.UUID),
		req.Name,
		req.Description,
		req.Symbol,
		req.RuleType,
		req.Conditions,
		req.Actions,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse the conditions and actions from JSON
	var conditions []services.RuleCondition
	var actions []services.RuleAction

	if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule conditions"})
		return
	}

	if err := json.Unmarshal(rule.Actions, &actions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule actions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"rule": ruleResponse{
			ID:          rule.ID.String(),
			Name:        rule.Name,
			Description: rule.Description,
			Symbol:      rule.Symbol,
			RuleType:    rule.RuleType,
			Conditions:  conditions,
			Actions:     actions,
			Status:      rule.Status,
			IsAIManaged: rule.IsAIManaged,
			CreatedAt:   rule.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   rule.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *RuleHandler) GetRules(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rules, err := h.ruleService.GetRulesByUserID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]ruleResponse, len(rules))
	for i, rule := range rules {
		var conditions []services.RuleCondition
		var actions []services.RuleAction

		if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule conditions"})
			return
		}

		if err := json.Unmarshal(rule.Actions, &actions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule actions"})
			return
		}

		response[i] = ruleResponse{
			ID:          rule.ID.String(),
			Name:        rule.Name,
			Description: rule.Description,
			Symbol:      rule.Symbol,
			RuleType:    rule.RuleType,
			Conditions:  conditions,
			Actions:     actions,
			Status:      rule.Status,
			IsAIManaged: rule.IsAIManaged,
			CreatedAt:   rule.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   rule.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, gin.H{"rules": response})
}

func (h *RuleHandler) GetRule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rule, err := h.ruleService.GetRuleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if rule belongs to user
	if rule.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not have access to this rule"})
		return
	}

	var conditions []services.RuleCondition
	var actions []services.RuleAction

	if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule conditions"})
		return
	}

	if err := json.Unmarshal(rule.Actions, &actions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse rule actions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rule": ruleResponse{
			ID:          rule.ID.String(),
			Name:        rule.Name,
			Description: rule.Description,
			Symbol:      rule.Symbol,
			RuleType:    rule.RuleType,
			Conditions:  conditions,
			Actions:     actions,
			Status:      rule.Status,
			IsAIManaged: rule.IsAIManaged,
			CreatedAt:   rule.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   rule.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *RuleHandler) ActivateRule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := h.ruleService.ActivateRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule activated successfully"})
}

func (h *RuleHandler) DeactivateRule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := h.ruleService.DeactivateRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule deactivated successfully"})
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule deleted successfully"})
}
