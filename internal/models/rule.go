// internal/models/rule.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TradingRule represents a user-defined trading rule
type TradingRule struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"not null"`
	Description string
	Symbol      string         `gorm:"not null"`
	RuleType    string         `gorm:"not null"` // stop_loss, take_profit, etc.
	Conditions  []byte         `gorm:"type:jsonb"`
	Actions     []byte         `gorm:"type:jsonb"`
	Status      string         `gorm:"default:active"`
	IsAIManaged bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for TradingRule model
func (TradingRule) TableName() string {
	return "trading_rules"
}

// BeforeCreate will set ID if not provided
func (r *TradingRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
