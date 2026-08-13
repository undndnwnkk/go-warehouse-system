package repository

import (
	"context"
	// "github.com/jackc/pgx/v5"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, request model.CreateOrderRequest) (*model.Order, error) {
	query := `
		INSERT INTO orders (user_id, status) VALUES ($1, $2) RETURNING id, user_id, status, created_at
	`

	var order model.Order
	if err := r.db.QueryRow(ctx, query, request.UserID, request.Status).Scan(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, request model.Order) (*model.Order, error) {
	query := `
		INSERT INTO orders (id, user_id, status) VALUES ($1, $2, $3) RETURNING id, user_id, status, created_at
	`

	var order model.Order
	if err := r.db.QueryRow(ctx, query, request.UserID, request.Status).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepository) ChangeOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	query := `
		UPDATE orders SET status = $1 WHERE id = $2
	`

	cmd, err := r.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

var (
	ErrNotFound = errors.New("not found")
)
