package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"medods-tz/internal/entity"
	"medods-tz/internal/repository"
	"medods-tz/internal/sender"
	"time"
)

type AuthService interface {
	CreateTokens(ctx context.Context, userID, clientIP string) (*entity.Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken, accessToken string) (*entity.Tokens, error)
}

type ServicesDependencies struct {
	Repository      *repository.Repository
	TokenTTL        time.Duration
	RefreshTokenTTL time.Duration
	SignKey         string
	SecurityLog     *logrus.Logger
	Sender          *sender.Sender
}

type Service struct {
	AuthService
}

func NewService(dependencies ServicesDependencies) *Service {
	return &Service{AuthService: NewAuth(
		dependencies.Repository.UserRepository,
		dependencies.Repository.TokenRepository,
		dependencies.TokenTTL,
		dependencies.RefreshTokenTTL,
		dependencies.SignKey,
		dependencies.SecurityLog,
		dependencies.Sender.Email)}
}
