package domain

import "time"

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Student struct {
	UserID   uint `gorm:"primaryKey"`
	User     User `gorm:"foreignKey:UserID"`
	Points   int  `json:"points"`
	Tokens   int  `json:"tokens" default:"0"`
	ParentID uint `json:"parent_id" default:"null"`
	IsActive bool `json:"is_active" default:"false"`
}

type Parent struct {
	UserID uint `gorm:"primaryKey"`
	User   User `gorm:"foreignKey:UserID"`
	Tokens int  `json:"tokens"  default:"0"`
}

type StandHolder struct {
	UserID  uint `gorm:"primaryKey"`
	User    User `gorm:"foreignKey:UserID"`
	StandID uint
	Stand   Stand `gorm:"foreignKey:StandID"`
}

type Organizer struct {
	UserID             uint       `gorm:"primaryKey"`
	User               User       `gorm:"foreignKey:UserID"`
	OrganizedKermesses []Kermesse `gorm:"many2many:organizer_kermesses;"`
}
