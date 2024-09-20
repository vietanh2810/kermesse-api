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
