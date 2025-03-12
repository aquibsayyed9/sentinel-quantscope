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
	ErrInvalidExecution = errors.New("invalid execution data")
	ErrRuleNotFound     = errors.New("rule not found")
)

// ExecutionService defines the interface for execution-related business operations
type ExecutionService interface {
	// Create a new execution record
	CreateExecution(ctx context.Context, execution *models.Execution) error

	// Get an execution by its ID
	GetExecutionByID(ctx context.Context, id uuid.UUID) (*models.Execution, error)

	// Get executions for a specific user with pagination
	GetUserExecutions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Execution, error)

	// Get executions for a specific trading rule
	GetRuleExecutions(ctx context.Context, ruleID uuid.UUID) ([]models.Execution, error)

	// Get recent executions across all users (for admin/monitoring purposes)
	GetRecentExecutions(ctx context.Context, limit int) ([]models.Execution, error)

	// Update an existing execution
	UpdateExecution(ctx context.Context, execution *models.Execution) error

	// Count total executions for a user
	CountUserExecutions(ctx context.Context, userID uuid.UUID) (int64, error)

	// Get user's executions with statistics
	GetUserExecutionStats(ctx context.Context, userID uuid.UUID, timeRange time.Duration) (*ExecutionStats, error)

	// Records an execution and updates related rule if needed
	ProcessExecution(ctx context.Context, execution *models.Execution) error
}

// ExecutionStats holds aggregated statistics about a user's trading executions
type ExecutionStats struct {
	TotalExecutions  int64              `json:"totalExecutions"`
	BuyCount         int64              `json:"buyCount"`
	SellCount        int64              `json:"sellCount"`
	TotalVolume      float64            `json:"totalVolume"`      // Total trade volume in currency
	AverageTradeSize float64            `json:"averageTradeSize"` // Average size per trade
	SymbolBreakdown  map[string]int     `json:"symbolBreakdown"`  // Count by symbol
	ExecutionsByDay  []DailyActivity    `json:"executionsByDay"`  // For activity charts
	TopSymbols       []SymbolStat       `json:"topSymbols"`       // Most traded symbols
	RecentExecutions []models.Execution `json:"recentExecutions"` // Recent executions
}

// DailyActivity represents trading activity for a specific day
type DailyActivity struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// SymbolStat represents statistics for a specific trading symbol
type SymbolStat struct {
	Symbol    string  `json:"symbol"`
	Count     int     `json:"count"`
	Volume    float64 `json:"volume"`
	BuyCount  int     `json:"buyCount"`
	SellCount int     `json:"sellCount"`
}

type executionService struct {
	executionRepo repository.ExecutionRepository
	ruleRepo      repository.RuleRepository // You'll need to implement this
}

// NewExecutionService creates a new instance of execution service
func NewExecutionService(executionRepo repository.ExecutionRepository, ruleRepo repository.RuleRepository) ExecutionService {
	return &executionService{
		executionRepo: executionRepo,
		ruleRepo:      ruleRepo,
	}
}

func (s *executionService) CreateExecution(ctx context.Context, execution *models.Execution) error {
	// Validate required fields
	if execution.UserID == uuid.Nil || execution.Symbol == "" ||
		execution.Quantity <= 0 || execution.Price <= 0 {
		return ErrInvalidExecution
	}

	// Set ID if not provided
	if execution.ID == uuid.Nil {
		execution.ID = uuid.New()
	}

	// Calculate total amount if not already set
	if execution.TotalAmount == 0 {
		execution.TotalAmount = execution.Price * execution.Quantity
	}

	// Set execution time to now if not provided
	if execution.ExecutionTime.IsZero() {
		execution.ExecutionTime = time.Now()
	}

	return s.executionRepo.Create(ctx, execution)
}

func (s *executionService) GetExecutionByID(ctx context.Context, id uuid.UUID) (*models.Execution, error) {
	return s.executionRepo.GetByID(ctx, id)
}

func (s *executionService) GetUserExecutions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Execution, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	offset := (page - 1) * pageSize
	return s.executionRepo.GetByUserID(ctx, userID, pageSize, offset)
}

func (s *executionService) GetRuleExecutions(ctx context.Context, ruleID uuid.UUID) ([]models.Execution, error) {
	return s.executionRepo.GetByRuleID(ctx, ruleID)
}

func (s *executionService) GetRecentExecutions(ctx context.Context, limit int) ([]models.Execution, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	return s.executionRepo.GetRecentExecutions(ctx, limit)
}

func (s *executionService) UpdateExecution(ctx context.Context, execution *models.Execution) error {
	// Validate execution before updating
	if execution.ID == uuid.Nil {
		return ErrInvalidExecution
	}

	return s.executionRepo.Update(ctx, execution)
}

func (s *executionService) CountUserExecutions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.executionRepo.CountByUserID(ctx, userID)
}

func (s *executionService) GetUserExecutionStats(ctx context.Context, userID uuid.UUID, timeRange time.Duration) (*ExecutionStats, error) {
	// Get all executions for the user within the time range
	startTime := time.Now().Add(-timeRange)

	// Get a large batch of recent executions to analyze
	// In a production application, you might want to implement specific repository methods
	// to calculate these statistics directly in the database for better performance
	executions, err := s.executionRepo.GetByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Filter executions by time range and initialize stats
	filtered := make([]models.Execution, 0)
	stats := &ExecutionStats{
		SymbolBreakdown: make(map[string]int),
		ExecutionsByDay: make([]DailyActivity, 0),
		TopSymbols:      make([]SymbolStat, 0),
	}

	// Maps to track statistics
	symbolStats := make(map[string]*SymbolStat)
	dailyActivity := make(map[string]int)

	for _, exec := range executions {
		if exec.ExecutionTime.After(startTime) {
			filtered = append(filtered, exec)

			// Count by symbol
			stats.SymbolBreakdown[exec.Symbol]++

			// Track daily activity
			dateKey := exec.ExecutionTime.Format("2006-01-02")
			dailyActivity[dateKey]++

			// Track symbol statistics
			if _, exists := symbolStats[exec.Symbol]; !exists {
				symbolStats[exec.Symbol] = &SymbolStat{
					Symbol: exec.Symbol,
				}
			}

			symbolStat := symbolStats[exec.Symbol]
			symbolStat.Count++
			symbolStat.Volume += exec.TotalAmount

			if exec.ExecutionType == "buy" {
				stats.BuyCount++
				symbolStat.BuyCount++
			} else if exec.ExecutionType == "sell" {
				stats.SellCount++
				symbolStat.SellCount++
			}

			stats.TotalVolume += exec.TotalAmount
		}
	}

	// Set summary statistics
	stats.TotalExecutions = int64(len(filtered))
	if stats.TotalExecutions > 0 {
		stats.AverageTradeSize = stats.TotalVolume / float64(stats.TotalExecutions)
	}

	// Convert daily activity map to slice
	for dateStr, count := range dailyActivity {
		date, _ := time.Parse("2006-01-02", dateStr)
		stats.ExecutionsByDay = append(stats.ExecutionsByDay, DailyActivity{
			Date:  date,
			Count: count,
		})
	}

	// Convert symbol stats map to slice and find top symbols
	for _, stat := range symbolStats {
		stats.TopSymbols = append(stats.TopSymbols, *stat)
	}

	// Sort top symbols by count (simplified - in production use sort.Slice)
	// This is a simple bubble sort - replace with a more efficient algorithm for production
	for i := 0; i < len(stats.TopSymbols)-1; i++ {
		for j := 0; j < len(stats.TopSymbols)-i-1; j++ {
			if stats.TopSymbols[j].Count < stats.TopSymbols[j+1].Count {
				stats.TopSymbols[j], stats.TopSymbols[j+1] = stats.TopSymbols[j+1], stats.TopSymbols[j]
			}
		}
	}

	// Limit to top 5 symbols
	if len(stats.TopSymbols) > 5 {
		stats.TopSymbols = stats.TopSymbols[:5]
	}

	// Get most recent executions
	if len(filtered) > 5 {
		stats.RecentExecutions = filtered[:5] // Assuming they're already sorted by time
	} else {
		stats.RecentExecutions = filtered
	}

	return stats, nil
}

func (s *executionService) ProcessExecution(ctx context.Context, execution *models.Execution) error {
	// Validate the execution
	if execution.UserID == uuid.Nil || execution.Symbol == "" ||
		execution.Quantity <= 0 || execution.Price <= 0 {
		return ErrInvalidExecution
	}

	// Calculate total amount if not set
	if execution.TotalAmount == 0 {
		execution.TotalAmount = execution.Price * execution.Quantity
	}

	// Set execution time to now if not provided
	if execution.ExecutionTime.IsZero() {
		execution.ExecutionTime = time.Now()
	}

	// Create the execution record
	err := s.executionRepo.Create(ctx, execution)
	if err != nil {
		return err
	}

	// If this execution is associated with a rule, update the rule's last execution time
	if execution.RuleID != nil {
		rule, err := s.ruleRepo.GetByID(ctx, *execution.RuleID)
		if err != nil {
			return err
		}

		// Update the rule's last execution time
		// Note: This assumes you have a method to update last execution time in your rule repository
		// You'll need to implement or adjust this based on your actual rule model
		err = s.updateRuleExecutionTime(ctx, rule, execution.ExecutionTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper method to update a rule's last execution time
// Note: Implement this method based on your actual rule model and repository interface
func (s *executionService) updateRuleExecutionTime(ctx context.Context, rule interface{}, executionTime time.Time) error {
	// This is a placeholder - implement based on your actual rule model and repository
	// For example:
	// rule.LastExecutionTime = executionTime
	// return s.ruleRepo.Update(ctx, rule)
	return nil
}
