// internal/repository/portfolio_repo.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aquibsayyed9/sentinel/internal/models"
)

var (
	ErrPortfolioNotFound = errors.New("portfolio not found")
	ErrHoldingNotFound   = errors.New("portfolio holding not found")
)

type PortfolioRepository interface {
	// Portfolio methods
	CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error
	GetPortfolioByUserID(ctx context.Context, userID uuid.UUID) (*models.Portfolio, error)
	UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error

	// Holdings methods
	CreateHolding(ctx context.Context, holding *models.PortfolioHolding) error
	GetHolding(ctx context.Context, portfolioID uuid.UUID, symbol string) (*models.PortfolioHolding, error)
	GetAllHoldings(ctx context.Context, portfolioID uuid.UUID) ([]models.PortfolioHolding, error)
	UpdateHolding(ctx context.Context, holding *models.PortfolioHolding) error
	DeleteHolding(ctx context.Context, id uuid.UUID) error
}

type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{db: db}
}

// Portfolio methods
func (r *portfolioRepository) CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	return r.db.WithContext(ctx).Create(portfolio).Error
}

func (r *portfolioRepository) GetPortfolioByUserID(ctx context.Context, userID uuid.UUID) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&portfolio).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}
	return &portfolio, nil
}

func (r *portfolioRepository) UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	result := r.db.WithContext(ctx).Save(portfolio)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPortfolioNotFound
	}
	return nil
}

// Holdings methods
func (r *portfolioRepository) CreateHolding(ctx context.Context, holding *models.PortfolioHolding) error {
	return r.db.WithContext(ctx).Create(holding).Error
}

func (r *portfolioRepository) GetHolding(ctx context.Context, portfolioID uuid.UUID, symbol string) (*models.PortfolioHolding, error) {
	var holding models.PortfolioHolding
	if err := r.db.WithContext(ctx).Where("portfolio_id = ? AND symbol = ?", portfolioID, symbol).First(&holding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHoldingNotFound
		}
		return nil, err
	}
	return &holding, nil
}

func (r *portfolioRepository) GetAllHoldings(ctx context.Context, portfolioID uuid.UUID) ([]models.PortfolioHolding, error) {
	var holdings []models.PortfolioHolding
	if err := r.db.WithContext(ctx).Where("portfolio_id = ?", portfolioID).Find(&holdings).Error; err != nil {
		return nil, err
	}
	return holdings, nil
}

func (r *portfolioRepository) UpdateHolding(ctx context.Context, holding *models.PortfolioHolding) error {
	result := r.db.WithContext(ctx).Save(holding)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHoldingNotFound
	}
	return nil
}

func (r *portfolioRepository) DeleteHolding(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.PortfolioHolding{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHoldingNotFound
	}
	return nil
}
