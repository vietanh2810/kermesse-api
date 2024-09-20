package repository

import (
	"context"
	"fmt"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
)

var (
	ErrUserEmailExists = dao.ErrUserEmailExists
	ErrUserNotFound    = dao.ErrUserNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user dao.User) (dao.User, error)
	FindByID(ctx context.Context, id uint) (dao.User, error)
	FindByEmail(ctx context.Context, email string) (dao.User, error)
}

type UserRepository struct {
	dao UserDAO
}

func NewUserRepository(dao UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	created, err := r.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.Insert -> %w", err)
	}

	return r.daoToDomain(created), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (domain.User, error) {
	found, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.FindByID -> %w", err)
	}

	return r.daoToDomain(found), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	found, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.FindByEmail -> %w", err)
	}

	return r.daoToDomain(found), nil
}

func (r *UserRepository) daoToDomain(u dao.User) domain.User {
	return domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
