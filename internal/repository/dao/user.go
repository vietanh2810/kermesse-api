package dao

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserEmailExists = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
)

type User struct {
	ID uint `gorm:"primaryKey"`

	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`

	Type string `gorm:"not null"` // "Eleve", "Parent", "TeneurDeStand", or "Organisateur"
	Name string `gorm:"not null"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Student struct {
	User
	Points   int `gorm:"not null"`
	Tokens   int `gorm:"not null"`
	ParentID uint
}

type Parent struct {
	User
	Tokens int `gorm:"not null"`
}

type StandHolder struct {
	User
	StandID uint `gorm:"not null"`
}

type Organizer struct {
	User
}

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (d *UserDAO) Insert(ctx context.Context, user User) (User, error) {
	result := d.db.WithContext(ctx).Create(&user)
	if result.Error != nil {
		var err *pgconn.PgError
		if errors.As(result.Error, &err) &&
			err.Code == pgerrcode.UniqueViolation &&
			strings.Contains(err.Message, `unique constraint "uni_users_email"`) {
			return User{}, ErrUserEmailExists
		}

		return User{}, result.Error
	}

	return user, nil
}

func (d *UserDAO) InsertStudent(ctx context.Context, student Student) (Student, error) {
	result := d.db.WithContext(ctx).Create(&student)
	if result.Error != nil {
		return Student{}, result.Error
	}
	return student, nil
}

func (d *UserDAO) InsertParent(ctx context.Context, parent Parent) (Parent, error) {
	result := d.db.WithContext(ctx).Create(&parent)
	if result.Error != nil {
		return Parent{}, result.Error
	}
	return parent, nil
}

func (d *UserDAO) InsertStandHolder(ctx context.Context, standHolder StandHolder) (StandHolder, error) {
	result := d.db.WithContext(ctx).Create(&standHolder)
	if result.Error != nil {
		return StandHolder{}, result.Error
	}
	return standHolder, nil
}

func (d *UserDAO) InsertOrganizer(ctx context.Context, organizer Organizer) (Organizer, error) {
	result := d.db.WithContext(ctx).Create(&organizer)
	if result.Error != nil {
		return Organizer{}, result.Error
	}
	return organizer, nil
}

func (d *UserDAO) FindByID(ctx context.Context, id uint) (User, error) {
	var user User

	result := d.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}

		return User{}, result.Error
	}

	return user, nil
}

func (d *UserDAO) FindStudentByID(ctx context.Context, id uint) (Student, error) {
	var student Student
	result := d.db.WithContext(ctx).First(&student, id)
	if result.Error != nil {
		return Student{}, result.Error
	}
	return student, nil
}

func (d *UserDAO) FindParentByID(ctx context.Context, id uint) (Parent, error) {
	var parent Parent
	result := d.db.WithContext(ctx).First(&parent, id)
	if result.Error != nil {
		return Parent{}, result.Error
	}
	return parent, nil
}

func (d *UserDAO) FindStandHolderByID(ctx context.Context, id uint) (StandHolder, error) {
	var standHolder StandHolder
	result := d.db.WithContext(ctx).First(&standHolder, id)
	if result.Error != nil {
		return StandHolder{}, result.Error
	}
	return standHolder, nil
}

func (d *UserDAO) FindOrganizerByID(ctx context.Context, id uint) (Organizer, error) {
	var organizer Organizer
	result := d.db.WithContext(ctx).First(&organizer, id)
	if result.Error != nil {
		return Organizer{}, result.Error
	}
	return organizer, nil
}

func (d *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User

	result := d.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}

		return User{}, result.Error
	}

	return user, nil
}

func (d *UserDAO) FindByType(ctx context.Context, userType string) ([]User, error) {
	var users []User

	result := d.db.WithContext(ctx).Where("type = ?", userType).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
