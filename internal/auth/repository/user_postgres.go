package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/model"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash FROM users WHERE email = $1
	`

	rows, err := r.db.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) SaveUser(ctx context.Context, u model.User) (string, error) {
	query := `
		INSERT INTO users VALUES (
			$1, $2
		) RETURNING id
	`
	var id string

	err := r.db.QueryRow(ctx, query, u.Email, u.PasswordHash).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

var (
	ErrNotFound = errors.New("not found")
)
