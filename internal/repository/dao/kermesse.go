package dao

import "time"

type Stand struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Type        string `gorm:"not null"` // "food", "drink", or "activity"
	Description string
	KermesseID  *uint    `gorm:"index"`
	Kermesse    Kermesse `gorm:"foreignKey:KermesseID"`
	Stock       []Stock  `gorm:"foreignKey:StandID"`
	TokensSpent int      `gorm:"default:0"`
	PointsGiven int      `gorm:"default:0"` // Only for activity stands
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Stock struct {
	ID        uint   `gorm:"primaryKey"`
	StandID   uint   `gorm:"not null"`
	ItemName  string `gorm:"not null"`
	Quantity  int    `gorm:"not null"`
	TokenCost int    `gorm:"not null"` // Cost in tokens
}

type Kermesse struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"not null"`
	Date         time.Time `gorm:"not null"`
	Location     string    `gorm:"not null"`
	Description  string
	Organizers   []Organizer `gorm:"many2many:organizer_kermesses;"`
	Participants []User      `gorm:"many2many:kermesse_participants;"`
	Stands       []Stand     `gorm:"foreignKey:KermesseID"`
	TokensSold   int         `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Tombola struct {
	ID         uint     `gorm:"primaryKey"`
	KermesseID uint     `gorm:"not null"`
	Kermesse   Kermesse `gorm:"foreignKey:KermessID"`
	Prizes     []Prize  `gorm:"foreignKey:TombolaID"`
	Tickets    []Ticket `gorm:"foreignKey:TombolaID"`
}

type Prize struct {
	ID        uint   `gorm:"primaryKey"`
	TombolaID uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
	Quantity  int    `gorm:"not null"`
}

type Ticket struct {
	ID        uint   `gorm:"primaryKey"`
	TombolaID uint   `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	Number    string `gorm:"not null"`
}
