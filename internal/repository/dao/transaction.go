package dao

import (
	"time"
)

type TokenTransactionType string

const (
	TokenPurchase     TokenTransactionType = "Purchase"
	TokenDistribution TokenTransactionType = "Distribution"
	TokenSpend        TokenTransactionType = "Spend"
)

type TokenTransaction struct {
	ID         uint                 `gorm:"primaryKey"`
	KermesseID uint                 `gorm:"not null"`
	FromID     uint                 `gorm:"not null"`
	FromType   string               `gorm:"not null"`
	ToID       uint                 `gorm:"not null"`
	ToType     string               `gorm:"not null"`
	Amount     int                  `gorm:"not null"`
	Type       TokenTransactionType `gorm:"not null"`
	StandID    *uint
	Status     string `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (TokenTransaction) TableName() string {
	return "token_transactions"
}
