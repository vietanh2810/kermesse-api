package domain

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
