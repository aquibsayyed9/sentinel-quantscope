package models

import (
	"time"

	"gorm.io/gorm"
)

// MarketData represents price information for a security
type MarketData struct {
	gorm.Model
	Symbol    string    `json:"symbol" gorm:"index:idx_symbol_timestamp"`
	Timestamp time.Time `json:"timestamp" gorm:"index:idx_symbol_timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Source    string    `json:"source"`
	TimeFrame string    `json:"time_frame"` // e.g., "1m", "5m", "1h", "1d"
}

// Quote represents real-time bid/ask data
type Quote struct {
	gorm.Model
	Symbol    string    `json:"symbol" gorm:"index"`
	Timestamp time.Time `json:"timestamp"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	BidSize   int       `json:"bid_size"`
	AskSize   int       `json:"ask_size"`
	Source    string    `json:"source"`
}
