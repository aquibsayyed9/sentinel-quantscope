// internal/repository/marketdata_repo.go
package repository

import (
	"context"
	"time"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"gorm.io/gorm"
)

type MarketDataRepository interface {
	SaveMarketData(ctx context.Context, data *models.MarketData) error
	GetLatestPrice(ctx context.Context, symbol string) (*models.MarketData, error)
	GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, timeframe string) ([]models.MarketData, error)
	SaveQuote(ctx context.Context, quote *models.Quote) error
	GetLatestQuote(ctx context.Context, symbol string) (*models.Quote, error)
}

type marketDataRepository struct {
	db *gorm.DB
}

func NewMarketDataRepository(db *gorm.DB) MarketDataRepository {
	return &marketDataRepository{db: db}
}

func (r *marketDataRepository) SaveMarketData(ctx context.Context, data *models.MarketData) error {
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *marketDataRepository) GetLatestPrice(ctx context.Context, symbol string) (*models.MarketData, error) {
	var data models.MarketData
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("timestamp desc").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *marketDataRepository) GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, timeframe string) ([]models.MarketData, error) {
	var data []models.MarketData
	query := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Where("timestamp BETWEEN ? AND ?", start, end)

	if timeframe != "" {
		query = query.Where("time_frame = ?", timeframe)
	}

	err := query.Order("timestamp asc").Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *marketDataRepository) SaveQuote(ctx context.Context, quote *models.Quote) error {
	return r.db.WithContext(ctx).Create(quote).Error
}

func (r *marketDataRepository) GetLatestQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	var quote models.Quote
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("timestamp desc").
		First(&quote).Error
	if err != nil {
		return nil, err
	}
	return &quote, nil
}
