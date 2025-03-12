// internal/repository/execution_repo.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aquibsayyed9/sentinel/internal/models"
)

var (
	ErrExecutionNotFound = errors.New("execution not found")
)

type ExecutionRepository interface {
	Create(ctx context.Context, execution *models.Execution) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Execution, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Execution, error)
	GetByRuleID(ctx context.Context, ruleID uuid.UUID) ([]models.Execution, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]models.Execution, error)
	Update(ctx context.Context, execution *models.Execution) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type executionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) ExecutionRepository {
	return &executionRepository{db: db}
}

func (r *executionRepository) Create(ctx context.Context, execution *models.Execution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

func (r *executionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Execution, error) {
	var execution models.Execution
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&execution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExecutionNotFound
		}
		return nil, err
	}
	return &execution, nil
}

func (r *executionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Execution, error) {
	var executions []models.Execution
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("execution_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&executions).Error; err != nil {
		return nil, err
	}

	return executions, nil
}

func (r *executionRepository) GetByRuleID(ctx context.Context, ruleID uuid.UUID) ([]models.Execution, error) {
	var executions []models.Execution
	if err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("execution_time DESC").Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}

func (r *executionRepository) GetRecentExecutions(ctx context.Context, limit int) ([]models.Execution, error) {
	var executions []models.Execution
	if err := r.db.WithContext(ctx).Order("execution_time DESC").Limit(limit).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}

func (r *executionRepository) Update(ctx context.Context, execution *models.Execution) error {
	result := r.db.WithContext(ctx).Save(execution)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExecutionNotFound
	}
	return nil
}

func (r *executionRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Execution{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
