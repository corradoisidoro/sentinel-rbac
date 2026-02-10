package service

import (
	"context"

	appErr "github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password string) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" || password == "" || len(password) < 6 {
		return nil, appErr.ErrInvalidInput
	}

	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, appErr.ErrInternal
	}
	if existing != nil {
		return nil, appErr.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErr.ErrInternal
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, appErr.ErrInternal
	}

	return user, nil
}
