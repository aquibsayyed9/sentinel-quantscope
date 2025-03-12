// test/mocks/rule_repository_mock.go
package mocks

import (
	"context"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *models.TradingRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TradingRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TradingRule), args.Error(1)
}

func (m *MockRuleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TradingRule, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.TradingRule), args.Error(1)
}

func (m *MockRuleRepository) GetActiveRules(ctx context.Context) ([]models.TradingRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.TradingRule), args.Error(1)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *models.TradingRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
