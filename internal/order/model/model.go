package model

import (
	"time"
)

type OrderStatus string

const (
	StatusPending OrderStatus = "pending"
	StatusPlaced  OrderStatus = "placed"
	StatusFailed  OrderStatus = "failed"
)

type Order struct {
	ID        string      `db:"id"`
	UserID    string      `db:"user_id"`
	Status    OrderStatus `db:"status"`
	CreatedAt time.Time   `db:"created_at"`
}

type CreateOrderRequest struct {
	UserID string
	Status OrderStatus
}

type OrderItem struct {
	ID       string `db:"id"`
	OrderID  string `db:"order_id"`
	SKU      string `db:"sku"`
	Quantity int64  `db:"quantity"`
}

type SaveOrderItemRequest struct {
	OrderID  string
	SKU      string
	Quantity int64
}

type CreateOrderItemRequest struct {
	SKU      string `json:"sku"`
	Quantity int64  `json:"quantity"`
}

type CreateOrderHttpRequest struct {
	Items []CreateOrderItemRequest `json:"items"`
}
