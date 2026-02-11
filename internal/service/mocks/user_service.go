package mocks

import (
	"context"

	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"github.com/corradoisidoro/sentinel-rbac/internal/service"
	"github.com/stretchr/testify/mock"
)

type UserServiceMock struct {
	mock.Mock
}

func (m *UserServiceMock) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

// 🔒 Compile-time interface check
var _ service.UserService = (*UserServiceMock)(nil)
