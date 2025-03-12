// internal/services/rule_service.go
package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/repository"
)

type RuleCondition struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol"`
	Operator  string      `json:"operator"`
	Value     float64     `json:"value"`
	TimeFrame string      `json:"time_frame,omitempty"`
	Params    interface{} `json:"params,omitempty"`
}

type RuleAction struct {
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
	Quantity  float64 `json:"quantity"`
	OrderType string  `json:"order_type"`
	Limit     float64 `json:"limit,omitempty"`
	Stop      float64 `json:"stop,omitempty"`
}

type RuleService interface {
	CreateRule(ctx context.Context, userID uuid.UUID, name, description, symbol, ruleType string,
		conditions []RuleCondition, actions []RuleAction) (*models.TradingRule, error)
	GetRuleByID(ctx context.Context, id uuid.UUID) (*models.TradingRule, error)
	GetRulesByUserID(ctx context.Context, userID uuid.UUID) ([]models.TradingRule, error)
	UpdateRule(ctx context.Context, rule *models.TradingRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ActivateRule(ctx context.Context, id uuid.UUID) error
	DeactivateRule(ctx context.Context, id uuid.UUID) error
}

type ruleService struct {
	ruleRepo repository.RuleRepository
}

func NewRuleService(ruleRepo repository.RuleRepository) RuleService {
	return &ruleService{
		ruleRepo: ruleRepo,
	}
}

func (s *ruleService) CreateRule(ctx context.Context, userID uuid.UUID, name, description, symbol, ruleType string,
	conditions []RuleCondition, actions []RuleAction) (*models.TradingRule, error) {

	// Convert conditions to JSON
	conditionsBytes, err := json.Marshal(conditions)
	if err != nil {
		return nil, err
	}

	// Convert actions to JSON
	actionsBytes, err := json.Marshal(actions)
	if err != nil {
		return nil, err
	}

	rule := &models.TradingRule{
		UserID:      userID,
		Name:        name,
		Description: description,
		Symbol:      symbol,
		RuleType:    ruleType,
		Conditions:  conditionsBytes,
		Actions:     actionsBytes,
		Status:      "active", // Default status
		IsAIManaged: false,    // Default to not AI-managed
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *ruleService) GetRuleByID(ctx context.Context, id uuid.UUID) (*models.TradingRule, error) {
	return s.ruleRepo.GetByID(ctx, id)
}

func (s *ruleService) GetRulesByUserID(ctx context.Context, userID uuid.UUID) ([]models.TradingRule, error) {
	return s.ruleRepo.GetByUserID(ctx, userID)
}

func (s *ruleService) UpdateRule(ctx context.Context, rule *models.TradingRule) error {
	return s.ruleRepo.Update(ctx, rule)
}

func (s *ruleService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.ruleRepo.Delete(ctx, id)
}

func (s *ruleService) ActivateRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	rule.Status = "active"
	return s.ruleRepo.Update(ctx, rule)
}

func (s *ruleService) DeactivateRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	rule.Status = "inactive"
	return s.ruleRepo.Update(ctx, rule)
}
