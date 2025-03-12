package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RuleExecution tracks when a trading rule is executed
type RuleExecution struct {
	gorm.Model
	RuleID       uuid.UUID `json:"rule_id" gorm:"index"`
	UserID       uuid.UUID `json:"user_id" gorm:"index"`
	Symbol       string    `json:"symbol"`
	ExecutedAt   time.Time `json:"executed_at"`
	TriggerPrice float64   `json:"trigger_price"`
	Quantity     float64   `json:"quantity"`
	OrderType    string    `json:"order_type"` // e.g., "market", "limit"
	Direction    string    `json:"direction"`  // e.g., "buy", "sell"
	Status       string    `json:"status"`     // e.g., "pending", "executed", "failed"
	Notes        string    `json:"notes"`
}

// RuleCondition represents a specific condition for a trading rule
type RuleCondition struct {
	gorm.Model
	RuleID          string  `json:"rule_id" gorm:"index"`
	ConditionType   string  `json:"condition_type"` // e.g., "price_above", "price_below", "moving_average"
	Symbol          string  `json:"symbol"`
	Parameter1      string  `json:"parameter1"`       // Flexible parameter (e.g., "price", "ma_period")
	Value1          float64 `json:"value1"`           // Threshold value
	Parameter2      string  `json:"parameter2"`       // Optional second parameter
	Value2          float64 `json:"value2"`           // Optional second value
	LogicalOperator string  `json:"logical_operator"` // For combining with other conditions ("AND", "OR")
}
