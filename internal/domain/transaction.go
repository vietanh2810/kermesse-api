package domain

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
	ID         uint
	KermesseID uint
	FromID     uint
	FromType   string
	ToID       uint
	ToType     string
	Amount     int
	Type       TokenTransactionType
	StandID    *uint
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (tt *TokenTransaction) Approve() {
	if tt.Type == TokenPurchase && tt.Status == "Pending" {
		tt.Status = "Approved"
	}
}

func (tt *TokenTransaction) Reject() {
	if tt.Type == TokenPurchase && tt.Status == "Pending" {
		tt.Status = "Rejected"
	}
}

func (tt *TokenTransaction) IsValid() bool {
	// Implement validation logic here
	if tt.FromID == tt.ToID {
		return false
	}
	if tt.Amount <= 0 {
		return false
	}
	// Add more validation as needed
	return true
}
