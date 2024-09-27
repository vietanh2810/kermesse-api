package domain

type Stock struct {
	ID        uint   `gorm:"primaryKey"`
	StandID   uint   `gorm:"not null"`
	ItemName  string `gorm:"not null"`
	Quantity  int    `gorm:"not null"`
	TokenCost int    `gorm:"not null"` // Cost in tokens
}
