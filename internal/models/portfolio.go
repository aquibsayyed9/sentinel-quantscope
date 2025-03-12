// internal/models/portfolio.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Portfolio represents a user's portfolio
type Portfolio struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null"`
	TotalValue  float64        `gorm:"not null"`
	CashBalance float64        `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for Portfolio model
func (Portfolio) TableName() string {
	return "portfolios"
}

// BeforeCreate will set ID if not provided
func (p *Portfolio) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PortfolioHolding represents a specific holding in a user's portfolio
type PortfolioHolding struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PortfolioID  uuid.UUID `gorm:"type:uuid;not null"`
	Symbol       string    `gorm:"not null"`
	Quantity     float64   `gorm:"not null"`
	AverageCost  float64   `gorm:"not null"`
	CurrentPrice float64
	LastUpdated  time.Time
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID"`
}

// TableName specifies the table name for PortfolioHolding model
func (PortfolioHolding) TableName() string {
	return "portfolio_holdings"
}

// BeforeCreate will set ID if not provided
func (h *PortfolioHolding) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
