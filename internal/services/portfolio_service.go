// internal/services/portfolio_service.go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/repository"
)

var (
	ErrPortfolioExists = errors.New("portfolio already exists for this user")
)

type PortfolioService interface {
	CreatePortfolio(ctx context.Context, userID uuid.UUID, initialBalance float64) (*models.Portfolio, error)
	GetPortfolioByUserID(ctx context.Context, userID uuid.UUID) (*models.Portfolio, error)
	UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error
	GetHoldings(ctx context.Context, userID uuid.UUID) ([]models.PortfolioHolding, error)
	AddOrUpdateHolding(ctx context.Context, userID uuid.UUID, symbol string, quantity, price float64) error
	RemoveHolding(ctx context.Context, userID uuid.UUID, symbol string) error
}

type portfolioService struct {
	portfolioRepo repository.PortfolioRepository
}

func NewPortfolioService(portfolioRepo repository.PortfolioRepository) PortfolioService {
	return &portfolioService{
		portfolioRepo: portfolioRepo,
	}
}

func (s *portfolioService) CreatePortfolio(ctx context.Context, userID uuid.UUID, initialBalance float64) (*models.Portfolio, error) {
	// Check if user already has a portfolio
	existing, err := s.portfolioRepo.GetPortfolioByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, ErrPortfolioExists
	} else if err != nil && !errors.Is(err, repository.ErrPortfolioNotFound) {
		return nil, err
	}

	portfolio := &models.Portfolio{
		UserID:      userID,
		TotalValue:  initialBalance,
		CashBalance: initialBalance,
	}

	if err := s.portfolioRepo.CreatePortfolio(ctx, portfolio); err != nil {
		return nil, err
	}

	return portfolio, nil
}

func (s *portfolioService) GetPortfolioByUserID(ctx context.Context, userID uuid.UUID) (*models.Portfolio, error) {
	return s.portfolioRepo.GetPortfolioByUserID(ctx, userID)
}

func (s *portfolioService) UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	return s.portfolioRepo.UpdatePortfolio(ctx, portfolio)
}

func (s *portfolioService) GetHoldings(ctx context.Context, userID uuid.UUID) ([]models.PortfolioHolding, error) {
	portfolio, err := s.portfolioRepo.GetPortfolioByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.portfolioRepo.GetAllHoldings(ctx, portfolio.ID)
}

func (s *portfolioService) AddOrUpdateHolding(ctx context.Context, userID uuid.UUID, symbol string, quantity, price float64) error {
	portfolio, err := s.portfolioRepo.GetPortfolioByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if holding already exists
	holding, err := s.portfolioRepo.GetHolding(ctx, portfolio.ID, symbol)
	if err != nil {
		if errors.Is(err, repository.ErrHoldingNotFound) {
			// Create new holding
			newHolding := &models.PortfolioHolding{
				PortfolioID:  portfolio.ID,
				Symbol:       symbol,
				Quantity:     quantity,
				AverageCost:  price,
				CurrentPrice: price,
				LastUpdated:  time.Now(),
			}
			return s.portfolioRepo.CreateHolding(ctx, newHolding)
		}
		return err
	}

	// Update existing holding (simple implementation - in real app, would calculate weighted average cost)
	totalValue := holding.AverageCost*holding.Quantity + price*quantity
	totalQuantity := holding.Quantity + quantity

	if totalQuantity <= 0 {
		// Remove holding if quantity is zero or negative
		return s.portfolioRepo.DeleteHolding(ctx, holding.ID)
	}

	holding.Quantity = totalQuantity
	holding.AverageCost = totalValue / totalQuantity
	holding.CurrentPrice = price
	holding.LastUpdated = time.Now()

	return s.portfolioRepo.UpdateHolding(ctx, holding)
}

func (s *portfolioService) RemoveHolding(ctx context.Context, userID uuid.UUID, symbol string) error {
	portfolio, err := s.portfolioRepo.GetPortfolioByUserID(ctx, userID)
	if err != nil {
		return err
	}

	holding, err := s.portfolioRepo.GetHolding(ctx, portfolio.ID, symbol)
	if err != nil {
		return err
	}

	return s.portfolioRepo.DeleteHolding(ctx, holding.ID)
}
