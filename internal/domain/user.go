package domain

import "time"

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Student struct {
	User
	Points   int  `json:"points"`
	Tokens   int  `json:"tokens"`
	ParentID uint `json:"parent_id"`
}

type Parent struct {
	User
	Tokens int `json:"tokens"`
}

type StandHolder struct {
	User
	StandID uint `json:"stand_id"`
}

type Organizer struct {
	User
}
