package service

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository"
)

var (
	ErrUserEmailExists = repository.ErrUserEmailExists
	ErrWrongPassword   = errors.New("wrong password")
	ErrStudentNotFound = repository.ErrUserNotFound
)

type AuthUserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	CreateStudent(ctx context.Context, student domain.Student) (domain.Student, error)
	FindStudentByEmail(ctx context.Context, email string) (domain.Student, error)
	UpdateStudent(ctx context.Context, student domain.Student) (domain.Student, error)
	CreateParent(ctx context.Context, parent domain.Parent) (domain.Parent, error)
	CreateStandHolder(ctx context.Context, standHolder domain.StandHolder) (domain.StandHolder, error)
	CreateOrganizer(ctx context.Context, organizer domain.Organizer) (domain.Organizer, error)
}

type AuthService struct {
	repo AuthUserRepository
}

func NewAuthService(repo AuthUserRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) Signup(ctx context.Context, user domain.User) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	user.Password = string(hash)

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.Create -> %w", err)
	}

	return created, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("s.repo.FindByEmail -> %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return domain.User{}, ErrWrongPassword
	}

	return user, nil
}

func (s *AuthService) SignupStudent(ctx context.Context, student domain.Student) (domain.User, error) {
	if err := s.checkEmailExists(ctx, student.User.Email); err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := hashPassword(student.User.Password)
	if err != nil {
		return domain.User{}, err
	}
	student.User.Password = hashedPassword

	createdStudent, err := s.repo.CreateStudent(ctx, student)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.CreateStudent -> %w", err)
	}

	return createdStudent.User, nil
}

func (s *AuthService) SignupStandHolder(ctx context.Context, standHolder domain.StandHolder) (domain.User, error) {
	if err := s.checkEmailExists(ctx, standHolder.User.Email); err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := hashPassword(standHolder.User.Password)
	if err != nil {
		return domain.User{}, err
	}
	standHolder.User.Password = hashedPassword

	createdStandHolder, err := s.repo.CreateStandHolder(ctx, standHolder)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.CreateStandHolder -> %w", err)
	}

	return createdStandHolder.User, nil
}

func (s *AuthService) SignupOrganizer(ctx context.Context, organizer domain.Organizer) (domain.User, error) {
	if err := s.checkEmailExists(ctx, organizer.User.Email); err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := hashPassword(organizer.User.Password)
	if err != nil {
		return domain.User{}, err
	}
	organizer.User.Password = hashedPassword

	createdOrganizer, err := s.repo.CreateOrganizer(ctx, organizer)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.CreateOrganizer -> %w", err)
	}

	return createdOrganizer.User, nil
}

func (s *AuthService) SignupParent(ctx context.Context, parent domain.Parent, studentEmails []string) (domain.User, error) {
	if err := s.checkEmailExists(ctx, parent.User.Email); err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := hashPassword(parent.User.Password)
	if err != nil {
		return domain.User{}, err
	}
	parent.User.Password = hashedPassword

	createdParent, err := s.repo.CreateParent(ctx, parent)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.CreateParent -> %w", err)
	}

	for _, studentEmail := range studentEmails {
		student, err := s.repo.FindStudentByEmail(ctx, studentEmail)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return domain.User{}, fmt.Errorf("student with email %s not found", studentEmail)
			}
			return domain.User{}, fmt.Errorf("s.repo.FindStudentByEmail -> %w", err)
		}

		student.IsActive = true
		student.ParentID = createdParent.UserID
		_, err = s.repo.UpdateStudent(ctx, student)
		if err != nil {
			return domain.User{}, fmt.Errorf("s.repo.UpdateStudent -> %w", err)
		}
	}

	return createdParent.User, nil
}

// Helper function for password hashing
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Helper function to check if email exists
func (s *AuthService) checkEmailExists(ctx context.Context, email string) error {
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return ErrUserEmailExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return err
	}
	return nil
}
