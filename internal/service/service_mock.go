package service

import (
	"context"
	"github.com/stretchr/testify/mock"
	"medods-tz/internal/entity"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entity.User), args.Error(1)
}

type mockTokenRepo struct {
	mock.Mock
}

func (m *mockTokenRepo) CreateRefreshToken(ctx context.Context, token entity.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockTokenRepo) GetRefreshTokenEntitiesByUserID(ctx context.Context, userID string) ([]entity.RefreshToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.RefreshToken), args.Error(1)
}

func (m *mockTokenRepo) MarkRefreshTokenUsed(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

type mockEmail struct {
	mock.Mock
}

func (m *mockEmail) SendWarningEmail(toEmail, subject, body string) error {
	args := m.Called(toEmail, subject, body)
	return args.Error(0)
}

func (m *mockEmail) EnsureSMTPConnection() error {
	args := m.Called()
	return args.Error(0)
}
