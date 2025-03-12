// internal/handlers/marketdata_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/aquibsayyed9/sentinel/internal/services"
	"github.com/gin-gonic/gin"
)

type MarketDataHandler struct {
	marketDataService services.MarketDataService
}

func NewMarketDataHandler(marketDataService services.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{
		marketDataService: marketDataService,
	}
}

func (h *MarketDataHandler) GetLatestPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	price, err := h.marketDataService.GetPrice(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"price": price})
}

func (h *MarketDataHandler) GetHistoricalData(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	timeframe := c.Query("timeframe")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
			return
		}
	} else {
		// Default to last 7 days
		start = time.Now().AddDate(0, 0, -7)
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
			return
		}
	} else {
		// Default to now
		end = time.Now()
	}

	data, err := h.marketDataService.GetHistoricalData(c.Request.Context(), symbol, start, end, timeframe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *MarketDataHandler) GetQuote(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	quote, err := h.marketDataService.GetQuote(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quote": quote})
}
