package service

import (
	"context"
	"time"

	"github.com/corradoisidoro/sentinel-rbac/internal/config"
	appErr "github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password, role string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type userService struct {
	repo   repository.UserRepository
	config config.Config
}

func NewUserService(repo repository.UserRepository, config config.Config) UserService {
	return &userService{repo: repo, config: config}
}

func (s *userService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	if email == "" || password == "" || len(password) < 8 || role == "" {
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
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, appErr.ErrInternal
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" || len(password) < 6 {
		return "", appErr.ErrInvalidInput
	}

	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", appErr.ErrInternal
	}
	if existing == nil {
		return "", appErr.ErrUserNotFound
	}

	// Look up the user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", appErr.ErrInternal
	}
	if user == nil {
		return "", appErr.ErrUserNotFound
	}

	// Compare sent in pass with saves hashed pass
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", appErr.ErrInvalidPassword
	}

	// If valid, generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
		"role": user.Role,
	})

	// Sign and get the complete encoded token as a string using the secretkey
	secretKey := []byte(s.config.JWTSecret)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", appErr.ErrFailedToGenerateToken
	}

	// Sent back token and user info
	return tokenString, nil
}
