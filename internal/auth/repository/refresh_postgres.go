package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/model"
)

type PostgresRefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(db *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) SaveRefreshToken(ctx context.Context, token model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) 
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *PostgresRefreshTokenRepository) GetActiveRefreshTokensByUserID(ctx context.Context, userID string) ([]model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, is_revoked, expires_at, created_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1 AND is_revoked = FALSE
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	tokens, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.RefreshToken])
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *PostgresRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID int64) error {
	query := `
		UPDATE refresh_tokens SET is_revoked = TRUE, updated_at = NOW() WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}
