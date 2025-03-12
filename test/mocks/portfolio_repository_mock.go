// test/mocks/portfolio_repository_mock.go
package mocks

import (
	"context"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockPortfolioRepository struct {
	mock.Mock
}

func (m *MockPortfolioRepository) CreatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	args := m.Called(ctx, portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) GetPortfolioByUserID(ctx context.Context, userID uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) UpdatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	args := m.Called(ctx, portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) CreateHolding(ctx context.Context, holding *models.PortfolioHolding) error {
	args := m.Called(ctx, holding)
	return args.Error(0)
}

func (m *MockPortfolioRepository) GetHolding(ctx context.Context, portfolioID uuid.UUID, symbol string) (*models.PortfolioHolding, error) {
	args := m.Called(ctx, portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PortfolioHolding), args.Error(1)
}

func (m *MockPortfolioRepository) GetAllHoldings(ctx context.Context, portfolioID uuid.UUID) ([]models.PortfolioHolding, error) {
	args := m.Called(ctx, portfolioID)
	return args.Get(0).([]models.PortfolioHolding), args.Error(1)
}

func (m *MockPortfolioRepository) UpdateHolding(ctx context.Context, holding *models.PortfolioHolding) error {
	args := m.Called(ctx, holding)
	return args.Error(0)
}

func (m *MockPortfolioRepository) DeleteHolding(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
