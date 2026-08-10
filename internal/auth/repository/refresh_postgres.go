package repository

import (
	"context"
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
		INSERT INTO refresh_tokens VALUES (
			$1, $2, $3
		)
	`

	_, err := r.db.Exec(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) GetRefreshTokenByID(ctx context.Context, id int64) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, is_revoked, expires_at, cretaed_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1
	`

	var token model.RefreshToken
	if err := r.db.QueryRow(ctx, query, id).Scan(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *PostgresRefreshTokenRepository) GetActiveRefreshTokensByUserID(ctx context.Context, userID string) ([]model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, is_revoked, expires_at, cretaed_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1
		AND
		is_revoked = FALSE
	`

	var tokens []model.RefreshToken
	if err := r.db.QueryRow(ctx, query, userID).Scan(&tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *PostgresRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID int64) error {
	query := `
		UPDATE refresh_tokens SET is_revoked = TRUE WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return err
	}

	return nil
}
