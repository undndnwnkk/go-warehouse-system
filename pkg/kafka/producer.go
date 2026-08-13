package kafka

import (
	"time"
)

type KafkaOrderEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
