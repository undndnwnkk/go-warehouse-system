package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	kafka_events "github.com/undndnwnkk/go-warehouse-system/pkg/kafka"
)

type OrderStore interface {
	CreateOrderWithItems(ctx context.Context, orderRequest model.CreateOrderRequest, items []model.CreateOrderItemRequest) (*model.Order, error)
	ChangeOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error
}

type EventWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type OrderService struct {
	store           OrderStore
	warehouseClient pb.WarehouseClient
	eventWriter     EventWriter
}

func NewOrderService(
	store OrderStore,
	wc pb.WarehouseClient,
	eventWriter EventWriter,
) *OrderService {
	return &OrderService{store: store, warehouseClient: wc, eventWriter: eventWriter}
}

func (s *OrderService) CreateOrder(ctx context.Context, items []model.CreateOrderItemRequest) (*model.Order, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized: missing user id")
	}

	orderToSave := model.CreateOrderRequest{UserID: userID, Status: model.StatusPending}
	order, err := s.store.CreateOrderWithItems(ctx, orderToSave, items)
	if err != nil {
		return nil, fmt.Errorf("create order with items: %w", err)
	}

	grpcItems := make([]*pb.OrderItem, len(items))
	for i, item := range items {
		grpcItems[i] = &pb.OrderItem{
			Sku:      item.SKU,
			Quantity: item.Quantity,
		}
	}

	grpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	grpcResponse, err := s.warehouseClient.ReserveStock(
		grpcCtx,
		&pb.ReserveStockRequest{OrderId: order.ID, Items: grpcItems},
	)

	if err != nil || !grpcResponse.GetSuccess() {
		order.Status = model.StatusFailed
		if statusErr := s.store.ChangeOrderStatus(ctx, order.ID, model.StatusFailed); statusErr != nil {
			return order, fmt.Errorf("reserve stock failed and update order status failed: %w", statusErr)
		}
		return order, errors.New("failed to reserve stock in warehouse")
	}

	order.Status = model.StatusPlaced
	if err := s.store.ChangeOrderStatus(ctx, order.ID, model.StatusPlaced); err != nil {
		return nil, fmt.Errorf("update order status to placed: %w", err)
	}

	event := kafka_events.KafkaOrderEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Status:    string(model.StatusPlaced),
		Timestamp: time.Now(),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal order event: %w", err)
	}

	err = s.eventWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order.ID),
		Value: eventBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("write order event: %w", err)
	}

	return order, nil
}
