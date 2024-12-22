package repository

import (
	"context"
	"github.com/jackc/pgx/v4"
	"medods-tz/internal/entity"
	"medods-tz/internal/repository/postgres"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token entity.RefreshToken) error
	GetRefreshTokenEntitiesByUserID(ctx context.Context, userID string) ([]entity.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, refreshID string) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type Repository struct {
	TokenRepository
	UserRepository
}

func NewRepository(pgConn *pgx.Conn) *Repository {
	return &Repository{
		TokenRepository: postgres.NewTokenPostgres(pgConn),
		UserRepository:  postgres.NewUserPostgres(pgConn),
	}
}
