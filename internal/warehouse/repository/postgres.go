package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Stock struct {
	SKU       string    `db:"sku"`
	Quantity  int64     `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateStockRequest struct {
	SKU      string `json:"sku"`
	Quantity int64  `json:"quantity"`
}

type PostgresWarehouseRepository struct {
	db *pgxpool.Pool
}

func NewPostgresWarehouseRepository(db *pgxpool.Pool) *PostgresWarehouseRepository {
	return &PostgresWarehouseRepository{db: db}
}

func (r *PostgresWarehouseRepository) CreateStock(ctx context.Context, request CreateStockRequest) (*Stock, error) {
	query := `
		INSERT INTO stock (sku, quantity) VALUES ($1, $2) RETURNING sku, quantity, created_at, updated_at
	`

	var res Stock
	if err := r.db.QueryRow(ctx, query, request.SKU, request.Quantity).Scan(
		&res.SKU, &res.Quantity, &res.CreatedAt, &res.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *PostgresWarehouseRepository) GetStock(ctx context.Context, sku string) (*Stock, error) {
	query := `
		SELECT sku, quantity, created_at, updated_at FROM stock WHERE sku = $1
	`

	rows, err := r.db.Query(ctx, query, sku)
	if err != nil {
		return nil, err
	}

	stock, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Stock])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &stock, nil
}
func (r *PostgresWarehouseRepository) UpdateStock(ctx context.Context, sku string, quantity int64) error {
	query := `
		UPDATE stock SET quantity = $2 WHERE sku = $1
	`

	cmd, err := r.db.Exec(ctx, query, sku, quantity)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresWarehouseRepository) Reserve(ctx context.Context, sku string, quantity int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	stock, err := r.GetStock(ctx, sku)
	if err != nil {
		return ErrNotFound
	}

	if stock.Quantity < quantity {
		return ErrNotEnough
	}

	stock.Quantity -= quantity

	err = r.UpdateStock(ctx, sku, stock.Quantity)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

var (
	ErrNotFound  = errors.New("not found")
	ErrNotEnough = errors.New("not enough")
)
