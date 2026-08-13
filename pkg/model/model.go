package model

type OrderItem struct {
	SKU      string `db:"sku"`
	Quantity int64  `db:"quantity"`
}
