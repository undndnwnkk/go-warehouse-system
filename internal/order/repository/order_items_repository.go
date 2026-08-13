package repository

import (
	"context"
	// "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
)

type PostgresOrderItemsRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderItemsRepository(db *pgxpool.Pool) *PostgresOrderItemsRepository {
	return &PostgresOrderItemsRepository{db: db}
}

func (r *PostgresOrderItemsRepository) CreateOrderItem(ctx context.Context, request model.SaveOrderItemRequest) (*model.OrderItem, error) {
	query := `
		INSERT INTO order_items (order_id, sku, quantity) VALUES ($1, $2, $3) RETURNING id, order_id, sku, quantity
	`

	var item model.OrderItem

	if err := r.db.QueryRow(ctx, query, request.OrderID, request.SKU, request.Quantity).Scan(
		&item.ID, &item.OrderID, &item.SKU, &item.Quantity,
	); err != nil {
		return nil, err
	}

	return &item, nil
}
