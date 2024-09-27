package domain

import "time"

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
