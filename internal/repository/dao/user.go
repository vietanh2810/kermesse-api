package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserEmailExists = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrStudentNotFound = errors.New("student not found")
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Role      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Student struct {
	UserID   uint `gorm:"primaryKey"`
	User     User `gorm:"foreignKey:UserID"`
	Points   int  `json:"points" default:"0"`
	Tokens   int  `json:"tokens" default:"0"`
	ParentID uint `json:"parent_id" default:"null"`
	IsActive bool `json:"is_active" default:"false"`
}

type Parent struct {
	UserID uint `gorm:"primaryKey"`
	User   User `gorm:"foreignKey:UserID"`
	Tokens int  `gorm:"not null" default:"0"`
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

//func (d *UserDAO) InsertStudent(ctx context.Context, student Student, user User) (Student, error) {
//	result := d.db.WithContext(ctx).Create(&student)
//	if result.Error != nil {
//		return Student{}, result.Error
//	}
//	return student, nil
//}

func (d *UserDAO) InsertStudent(ctx context.Context, user User, student Student) (Student, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Student{}, tx.Error
	}

	// Insert User first
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return Student{}, err
	}

	// Set the UserID for the student
	student.UserID = user.ID

	// Now insert the Student
	if err := tx.Create(&student).Error; err != nil {
		tx.Rollback()
		return Student{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Student{}, err
	}

	// Fetch the complete student record including the associated user
	var completeStudent Student
	if err := d.db.WithContext(ctx).Preload("User").First(&completeStudent, student.UserID).Error; err != nil {
		return Student{}, err
	}

	return completeStudent, nil
}

func (d *UserDAO) UpdateStudent(ctx context.Context, user User, student Student) (Student, error) {
	tx := d.db.WithContext(ctx).Begin()

	fmt.Printf("Student here: ", student.UserID)
	if tx.Error != nil {
		return Student{}, tx.Error
	}

	// Update User first
	if err := tx.Model(&User{}).Where("id = ?", student.UserID).Updates(user).Error; err != nil {
		tx.Rollback()
		return Student{}, err
	}

	// Now update the Student
	if err := tx.Model(&Student{}).Where("user_id = ?", student.UserID).Updates(student).Error; err != nil {
		tx.Rollback()
		return Student{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Student{}, err
	}

	// Fetch the updated student record including the associated user
	var updatedStudent Student
	if err := d.db.WithContext(ctx).Preload("User").First(&updatedStudent, student.UserID).Error; err != nil {
		return Student{}, err
	}

	return updatedStudent, nil
}

func (d *UserDAO) UpdateParent(ctx context.Context, user User, parent Parent) (Parent, error) {
	tx := d.db.WithContext(ctx).Begin()

	if tx.Error != nil {
		return Parent{}, tx.Error
	}

	// Update User first
	if err := tx.Model(&User{}).Where("id = ?", parent.UserID).Updates(user).Error; err != nil {
		tx.Rollback()
		return Parent{}, err
	}

	// Now update the Parent
	if err := tx.Model(&Parent{}).Where("user_id = ?", parent.UserID).Updates(parent).Error; err != nil {
		tx.Rollback()
		return Parent{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Parent{}, err
	}

	// Fetch the updated parent record including the associated user
	var updatedParent Parent
	if err := d.db.WithContext(ctx).Preload("User").First(&updatedParent, parent.UserID).Error; err != nil {
		return Parent{}, err
	}

	return updatedParent, nil
}

func (d *UserDAO) InsertParent(ctx context.Context, user User, parent Parent) (Parent, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Parent{}, tx.Error
	}

	// Insert User first
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return Parent{}, err
	}

	// Set the UserID for the parent
	parent.UserID = user.ID

	// Now insert the Parent
	if err := tx.Create(&parent).Error; err != nil {
		tx.Rollback()
		return Parent{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Parent{}, err
	}

	// Fetch the complete parent record including the associated user
	var completeParent Parent
	if err := d.db.WithContext(ctx).Preload("User").First(&completeParent, parent.UserID).Error; err != nil {
		return Parent{}, err
	}

	fmt.Printf("completedParent UserId: ", completeParent.UserID)

	return completeParent, nil
}

func (d *UserDAO) InsertStandHolder(ctx context.Context, user User, stand Stand, standHolder StandHolder) (StandHolder, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return StandHolder{}, tx.Error
	}

	// Insert User first
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return StandHolder{}, err
	}

	// Set the UserID for the stand holder
	standHolder.UserID = user.ID

	if err := tx.Create(&stand).Error; err != nil {
		tx.Rollback()
		return StandHolder{}, err
	}

	// Set the StandID for the stand holder
	standHolder.StandID = stand.ID

	// Now insert the Stand Holder
	if err := tx.Create(&standHolder).Error; err != nil {
		tx.Rollback()
		return StandHolder{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return StandHolder{}, err
	}

	// Fetch the complete stand holder record including the associated user
	var completeStandHolder StandHolder
	if err := d.db.WithContext(ctx).
		Preload("User").
		Preload("Stand").
		First(&completeStandHolder, standHolder.UserID).Error; err != nil {
		return StandHolder{}, err
	}

	return completeStandHolder, nil
}

func (d *UserDAO) InsertOrganizer(ctx context.Context, user User) (Organizer, error) {

	organizer := Organizer{}
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Organizer{}, tx.Error
	}

	// Insert User first
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return Organizer{}, err
	}

	// Set the UserID for the stand holder
	organizer.UserID = user.ID

	if err := tx.Create(&organizer).Error; err != nil {
		tx.Rollback()
		return Organizer{}, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return Organizer{}, err
	}

	// Fetch the complete stand holder record including the associated user
	var completeOrganizer Organizer
	if err := d.db.WithContext(ctx).
		Preload("User").
		First(&completeOrganizer, organizer.UserID).Error; err != nil {
		return Organizer{}, err
	}

	return completeOrganizer, nil
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

//func (d *UserDAO) FindStudentByID(ctx context.Context, id uint) (Student, error) {
//	var student Student
//	result := d.db.WithContext(ctx).First(&student, id)
//	if result.Error != nil {
//		return Student{}, result.Error
//	}
//	return student, nil
//}
//
//func (d *UserDAO) FindParentByID(ctx context.Context, id uint) (Parent, error) {
//	var parent Parent
//	result := d.db.WithContext(ctx).First(&parent, id)
//	if result.Error != nil {
//		return Parent{}, result.Error
//	}
//	return parent, nil
//}
//
//func (d *UserDAO) FindStandHolderByID(ctx context.Context, id uint) (StandHolder, error) {
//	var standHolder StandHolder
//	result := d.db.WithContext(ctx).First(&standHolder, id)
//	if result.Error != nil {
//		return StandHolder{}, result.Error
//	}
//	return standHolder, nil
//}
//
//func (d *UserDAO) FindOrganizerByID(ctx context.Context, id uint) (Organizer, error) {
//	var organizer Organizer
//	result := d.db.WithContext(ctx).First(&organizer, id)
//	if result.Error != nil {
//		return Organizer{}, result.Error
//	}
//	return organizer, nil
//}

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

func (d *UserDAO) FindPendingStudentsByParentEmail(ctx context.Context, email string) ([]Student, error) {
	var students []Student
	result := d.db.WithContext(ctx).Where("parent_email = ? AND parent_id IS NULL", email).Find(&students)
	if result.Error != nil {
		return nil, result.Error
	}
	return students, nil
}

func (d *UserDAO) FindStudentByEmail(ctx context.Context, email string) (Student, error) {
	var student Student

	result := d.db.WithContext(ctx).
		Joins("JOIN users ON students.user_id = users.id").
		Where("students.user_id = (SELECT id FROM users WHERE email = ?)", email).
		Preload("User").
		First(&student)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Student{}, ErrStudentNotFound
		}
		return Student{}, result.Error
	}

	return student, nil
}

func (d *UserDAO) FindStudentByUserID(ctx context.Context, id uint) (Student, error) {
	var student Student
	result := d.db.WithContext(ctx).
		Where("user_id = ?", id).
		Preload("User").
		First(&student)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Student{}, ErrStudentNotFound
		}
		return Student{}, fmt.Errorf("failed to find student: %w", result.Error)
	}

	return student, nil
}

func (d *UserDAO) FindParentByUserID(ctx context.Context, id uint) (Parent, error) {
	var parent Parent
	result := d.db.WithContext(ctx).
		Where("user_id = ?", id).
		Preload("User").
		First(&parent)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Parent{}, ErrUserNotFound
		}
		return Parent{}, fmt.Errorf("failed to find parent: %w", result.Error)
	}

	return parent, nil
}

func (d *UserDAO) FindStandHolderByUserID(ctx context.Context, id uint) (StandHolder, error) {
	var standHolder StandHolder
	result := d.db.WithContext(ctx).
		Where("user_id = ?", id).
		Preload("User").
		Preload("Stand").
		First(&standHolder)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return StandHolder{}, ErrUserNotFound
		}
		return StandHolder{}, fmt.Errorf("failed to find stand holder: %w", result.Error)
	}

	return standHolder, nil
}
