// internal/services/marketdata_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/repository"
)

type MarketDataService interface {
	GetPrice(ctx context.Context, symbol string) (*models.MarketData, error)
	GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, timeframe string) ([]models.MarketData, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
	// Additional methods for external data fetching would be added here
}

type marketDataService struct {
	marketDataRepo repository.MarketDataRepository
	// You might add API clients for external data providers here
}

func NewMarketDataService(marketDataRepo repository.MarketDataRepository) MarketDataService {
	return &marketDataService{
		marketDataRepo: marketDataRepo,
	}
}

func (s *marketDataService) GetPrice(ctx context.Context, symbol string) (*models.MarketData, error) {
	return s.marketDataRepo.GetLatestPrice(ctx, symbol)
}

func (s *marketDataService) GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, timeframe string) ([]models.MarketData, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("end date must be after start date")
	}
	return s.marketDataRepo.GetHistoricalData(ctx, symbol, start, end, timeframe)
}

func (s *marketDataService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return s.marketDataRepo.GetLatestQuote(ctx, symbol)
}
