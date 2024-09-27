package domain

import "time"

type Stand struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Type        string `gorm:"not null"` // "food", "drink", or "activity"
	Description string
	KermesseID  uint     `gorm:"not null"`
	Kermesse    Kermesse `gorm:"foreignKey:KermessID"`
	Stock       []Stock  `gorm:"foreignKey:StandID"`
	TokensSpent int      `gorm:"default:0"`
	PointsGiven int      `gorm:"default:0"` // Only for activity stands
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
