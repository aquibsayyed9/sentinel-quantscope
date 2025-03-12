// internal/handlers/portfolio_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aquibsayyed9/sentinel/internal/services"
)

type PortfolioHandler struct {
	portfolioService services.PortfolioService
}

func NewPortfolioHandler(portfolioService services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
	}
}

type createPortfolioRequest struct {
	InitialBalance float64 `json:"initial_balance" binding:"required,gt=0"`
}

type addHoldingRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

type portfolioResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TotalValue  float64   `json:"total_value"`
	CashBalance float64   `json:"cash_balance"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type holdingResponse struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	Quantity      float64   `json:"quantity"`
	AverageCost   float64   `json:"average_cost"`
	CurrentPrice  float64   `json:"current_price"`
	MarketValue   float64   `json:"market_value"`
	ProfitLoss    float64   `json:"profit_loss"`
	ProfitLossPct float64   `json:"profit_loss_pct"`
	LastUpdated   time.Time `json:"last_updated"`
}

func (h *PortfolioHandler) CreatePortfolio(c *gin.Context) {
	var req createPortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	portfolio, err := h.portfolioService.CreatePortfolio(c.Request.Context(), userID.(uuid.UUID), req.InitialBalance)
	if err != nil {
		if err == services.ErrPortfolioExists {
			c.JSON(http.StatusConflict, gin.H{"error": "portfolio already exists for this user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"portfolio": portfolioResponse{
			ID:          portfolio.ID.String(),
			UserID:      portfolio.UserID.String(),
			TotalValue:  portfolio.TotalValue,
			CashBalance: portfolio.CashBalance,
			CreatedAt:   portfolio.CreatedAt,
			UpdatedAt:   portfolio.UpdatedAt,
		},
	})
}

func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	portfolio, err := h.portfolioService.GetPortfolioByUserID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portfolio": portfolioResponse{
			ID:          portfolio.ID.String(),
			UserID:      portfolio.UserID.String(),
			TotalValue:  portfolio.TotalValue,
			CashBalance: portfolio.CashBalance,
			CreatedAt:   portfolio.CreatedAt,
			UpdatedAt:   portfolio.UpdatedAt,
		},
	})
}

func (h *PortfolioHandler) GetHoldings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	holdings, err := h.portfolioService.GetHoldings(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]holdingResponse, len(holdings))
	for i, holding := range holdings {
		marketValue := holding.Quantity * holding.CurrentPrice
		profitLoss := marketValue - (holding.Quantity * holding.AverageCost)
		profitLossPct := 0.0
		if holding.AverageCost > 0 {
			profitLossPct = (profitLoss / (holding.Quantity * holding.AverageCost)) * 100
		}

		response[i] = holdingResponse{
			ID:            holding.ID.String(),
			Symbol:        holding.Symbol,
			Quantity:      holding.Quantity,
			AverageCost:   holding.AverageCost,
			CurrentPrice:  holding.CurrentPrice,
			MarketValue:   marketValue,
			ProfitLoss:    profitLoss,
			ProfitLossPct: profitLossPct,
			LastUpdated:   holding.LastUpdated,
		}
	}

	c.JSON(http.StatusOK, gin.H{"holdings": response})
}

func (h *PortfolioHandler) AddHolding(c *gin.Context) {
	var req addHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.portfolioService.AddOrUpdateHolding(c.Request.Context(), userID.(uuid.UUID), req.Symbol, req.Quantity, req.Price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "holding added successfully"})
}

func (h *PortfolioHandler) RemoveHolding(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.portfolioService.RemoveHolding(c.Request.Context(), userID.(uuid.UUID), symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "holding removed successfully"})
}
