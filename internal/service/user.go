package service

import (
	"context"
	"fmt"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository"
)

var (
	ErrUserNotFound = repository.ErrUserNotFound
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint) (domain.User, error)
	CreateStudent(ctx context.Context, student domain.Student) (domain.Student, error)
	FindStudentByUserID(ctx context.Context, id uint) (domain.Student, error)
	FindParentByUserID(ctx context.Context, id uint) (domain.Parent, error)
	FindStandHolderByUserID(ctx context.Context, id uint) (domain.StandHolder, error)
	UpdateUserTokens(ctx context.Context, userID uint, amount int) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetUser(ctx context.Context, id uint) (domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("s.repo.FindByID -> %w", err)
	}

	return user, nil
}

func (s *UserService) GetStudentByUserID(ctx context.Context, userID uint) (domain.Student, error) {
	student, err := s.repo.FindStudentByUserID(ctx, userID)
	if err != nil {
		return domain.Student{}, fmt.Errorf("s.repo.CreateStudent -> %w", err)
	}

	return student, nil
}

func (s *UserService) GetParentByUserID(ctx context.Context, userID uint) (domain.Parent, error) {
	parent, err := s.repo.FindParentByUserID(ctx, userID)
	if err != nil {
		return domain.Parent{}, fmt.Errorf("s.repo.FindParentByUserID -> %w", err)
	}

	return parent, nil
}

func (s *UserService) GetStandHolderByUserID(ctx context.Context, userID uint) (domain.StandHolder, error) {
	standHolder, err := s.repo.FindStandHolderByUserID(ctx, userID)
	if err != nil {
		return domain.StandHolder{}, fmt.Errorf("s.repo.FindStandHolderByUserID -> %w", err)
	}

	return standHolder, nil
}

func (s *UserService) GetUserTokens(ctx context.Context, userID uint) (int, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("s.repo.FindByID -> %w", err)
	}

	if user.Role == "parent" {
		parent, err := s.repo.FindParentByUserID(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("s.repo.FindParentByUserID -> %w", err)
		}
		return parent.Tokens, nil
	} else if user.Role == "student" {
		student, err := s.repo.FindStudentByUserID(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("s.repo.FindStudentByUserID -> %w", err)
		}
		return student.Tokens, nil
	}

	return 0, nil
}
