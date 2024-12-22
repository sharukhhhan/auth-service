package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"medods-tz/internal/entity"
	"medods-tz/internal/repository/repoerrors"
)

type TokenPostgres struct {
	*pgx.Conn
}

func NewTokenPostgres(conn *pgx.Conn) *TokenPostgres {
	return &TokenPostgres{Conn: conn}
}

func (p *TokenPostgres) CreateRefreshToken(ctx context.Context, token entity.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, refresh_hash, issued_at, expires_at, client_ip)
				VALUES($1, $2, $3, $4, $5)`

	_, err := p.Exec(ctx, query,
		token.UserID,
		token.RefreshHash,
		token.IssuedAt,
		token.ExpiresAt,
		token.ClientIP,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repoerrors.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (p *TokenPostgres) GetRefreshTokenEntitiesByUserID(ctx context.Context, userID string) ([]entity.RefreshToken, error) {
	query := `SELECT * FROM refresh_tokens WHERE user_id = $1`
	rows, err := p.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var refreshTokens []entity.RefreshToken
	for rows.Next() {
		var token entity.RefreshToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.RefreshHash,
			&token.IssuedAt,
			&token.ExpiresAt,
			&token.ClientIP,
			&token.Used)
		if err != nil {
			return nil, err
		}

		refreshTokens = append(refreshTokens, token)
	}

	return refreshTokens, nil
}

func (p *TokenPostgres) MarkRefreshTokenUsed(ctx context.Context, refreshID string) error {
	query := `UPDATE refresh_tokens SET used = true WHERE id = $1`
	res, err := p.Exec(ctx, query, refreshID)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected < 1 {
		return repoerrors.ErrNotFound
	}

	return nil
}
