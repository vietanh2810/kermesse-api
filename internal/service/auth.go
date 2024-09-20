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
)

type AuthUserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
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
