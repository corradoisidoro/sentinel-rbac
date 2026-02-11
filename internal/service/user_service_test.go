package service_test

import (
	"context"
	"testing"

	"github.com/corradoisidoro/sentinel-rbac/internal/config"
	appErr "github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository/mocks"
	"github.com/corradoisidoro/sentinel-rbac/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_SignUp_Success(t *testing.T) {
	repo := new(mocks.UserRepositoryMock)
	config := config.Config{}
	svc := service.NewUserService(repo, config)

	repo.On("FindByEmail", mock.Anything, "test@example.com").
		Return(nil, nil)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(nil)

	user, err := svc.Register(context.Background(), "test@example.com", "password123", "user")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
	assert.NotEmpty(t, user.Password)

	repo.AssertExpectations(t)
}

func TestUserService_SignUp_UserAlreadyExists(t *testing.T) {
	repo := new(mocks.UserRepositoryMock)
	config := config.Config{}
	svc := service.NewUserService(repo, config)

	repo.On("FindByEmail", mock.Anything, "test@example.com").
		Return(&models.User{Email: "test@example.com"}, nil)

	user, err := svc.Register(context.Background(), "test@example.com", "password123", "user")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, appErr.ErrUserAlreadyExists, err)

	repo.AssertExpectations(t)
}

func TestUserService_SignUp_InvalidInput(t *testing.T) {
	repo := new(mocks.UserRepositoryMock)
	config := config.Config{}
	svc := service.NewUserService(repo, config)

	user, err := svc.Register(context.Background(), "", "123", "user")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, appErr.ErrInvalidInput, err)
}
