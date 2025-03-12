// internal/repository/rule_repo.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aquibsayyed9/sentinel/internal/models"
)

var (
	ErrRuleNotFound = errors.New("trading rule not found")
)

type RuleRepository interface {
	Create(ctx context.Context, rule *models.TradingRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.TradingRule, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TradingRule, error)
	GetActiveRules(ctx context.Context) ([]models.TradingRule, error)
	Update(ctx context.Context, rule *models.TradingRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ruleRepository struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) RuleRepository {
	return &ruleRepository{db: db}
}

func (r *ruleRepository) Create(ctx context.Context, rule *models.TradingRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *ruleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TradingRule, error) {
	var rule models.TradingRule
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	return &rule, nil
}

func (r *ruleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TradingRule, error) {
	var rules []models.TradingRule
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *ruleRepository) GetActiveRules(ctx context.Context) ([]models.TradingRule, error) {
	var rules []models.TradingRule
	if err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *ruleRepository) Update(ctx context.Context, rule *models.TradingRule) error {
	result := r.db.WithContext(ctx).Save(rule)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}

func (r *ruleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.TradingRule{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}
