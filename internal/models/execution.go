// internal/models/execution.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Execution represents a trade execution
type Execution struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RuleID          *uuid.UUID `gorm:"type:uuid"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null"`
	Symbol          string     `gorm:"not null"`
	ExecutionType   string     `gorm:"not null"` // buy, sell
	Quantity        float64    `gorm:"not null"`
	Price           float64    `gorm:"not null"`
	TotalAmount     float64    `gorm:"not null"`
	Status          string     `gorm:"not null"`
	ExecutionTime   time.Time  `gorm:"not null"`
	Exchange        string
	ExternalOrderID string
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	// Relationships
	User User         `gorm:"foreignKey:UserID"`
	Rule *TradingRule `gorm:"foreignKey:RuleID"`
}

// TableName specifies the table name for Execution model
func (Execution) TableName() string {
	return "executions"
}

// BeforeCreate will set ID if not provided
func (e *Execution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
