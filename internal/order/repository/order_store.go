package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
)

type PostgresOrderStore struct {
	db *pgxpool.Pool
}

func NewPostgresOrderStore(db *pgxpool.Pool) *PostgresOrderStore {
	return &PostgresOrderStore{db: db}
}

func (s *PostgresOrderStore) CreateOrderWithItems(
	ctx context.Context,
	orderRequest model.CreateOrderRequest,
	items []model.CreateOrderItemRequest,
) (*model.Order, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var order model.Order
	createOrderQuery := `
		INSERT INTO orders (user_id, status)
		VALUES ($1, $2)
		RETURNING id, user_id, status, created_at
	`
	if err := tx.QueryRow(ctx, createOrderQuery, orderRequest.UserID, orderRequest.Status).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.CreatedAt,
	); err != nil {
		return nil, err
	}

	createItemQuery := `
		INSERT INTO order_items (order_id, sku, quantity)
		VALUES ($1, $2, $3)
	`
	for _, item := range items {
		if _, err := tx.Exec(ctx, createItemQuery, order.ID, item.SKU, item.Quantity); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *PostgresOrderStore) ChangeOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	query := `
		UPDATE orders SET status = $1 WHERE id = $2
	`

	cmd, err := s.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
