package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"medods-tz/internal/entity"
	"medods-tz/internal/repository/repoerrors"
)

type UserPostgres struct {
	*pgx.Conn
}

func NewUserPostgres(conn *pgx.Conn) *UserPostgres {
	return &UserPostgres{Conn: conn}
}

func (p *UserPostgres) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err := p.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrors.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
