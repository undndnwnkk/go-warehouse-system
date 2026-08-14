package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	kafkaevents "github.com/undndnwnkk/go-warehouse-system/pkg/kafka"
	"google.golang.org/grpc"
)

func TestCreateOrder_WhenWarehouseReturnsError_MarksOrderAsFailed(t *testing.T) {
	store := &fakeOrderStore{
		order: &model.Order{
			ID:        "order-1",
			UserID:    "user-1",
			Status:    model.StatusPending,
			CreatedAt: time.Now(),
		},
	}
	warehouse := &fakeWarehouseClient{err: errors.New("warehouse unavailable")}
	writer := &fakeEventWriter{}

	svc := NewOrderService(store, warehouse, writer)

	order, err := svc.CreateOrder(contextWithUserID(), []model.CreateOrderItemRequest{
		{SKU: "SKU-001", Quantity: 2},
	})

	if err == nil {
		t.Fatal("expected warehouse error")
	}
	if order == nil {
		t.Fatal("expected failed order to be returned")
	}
	if order.Status != model.StatusFailed {
		t.Fatalf("expected returned order status %q, got %q", model.StatusFailed, order.Status)
	}
	if len(store.statusChanges) != 1 {
		t.Fatalf("expected 1 status change, got %d", len(store.statusChanges))
	}
	if store.statusChanges[0].status != model.StatusFailed {
		t.Fatalf("expected stored status %q, got %q", model.StatusFailed, store.statusChanges[0].status)
	}
	if len(writer.messages) != 0 {
		t.Fatalf("expected no kafka messages, got %d", len(writer.messages))
	}
}

func TestCreateOrder_WhenWarehouseSucceeds_SendsKafkaMessage(t *testing.T) {
	store := &fakeOrderStore{
		order: &model.Order{
			ID:        "order-1",
			UserID:    "user-1",
			Status:    model.StatusPending,
			CreatedAt: time.Now(),
		},
	}
	warehouse := &fakeWarehouseClient{
		response: &pb.ReserveStockResponse{Success: true},
	}
	writer := &fakeEventWriter{}

	svc := NewOrderService(store, warehouse, writer)

	order, err := svc.CreateOrder(contextWithUserID(), []model.CreateOrderItemRequest{
		{SKU: "SKU-001", Quantity: 2},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Status != model.StatusPlaced {
		t.Fatalf("expected returned order status %q, got %q", model.StatusPlaced, order.Status)
	}
	if len(store.statusChanges) != 1 {
		t.Fatalf("expected 1 status change, got %d", len(store.statusChanges))
	}
	if store.statusChanges[0].status != model.StatusPlaced {
		t.Fatalf("expected stored status %q, got %q", model.StatusPlaced, store.statusChanges[0].status)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 kafka message, got %d", len(writer.messages))
	}

	var event kafkaevents.KafkaOrderEvent
	if err := json.Unmarshal(writer.messages[0].Value, &event); err != nil {
		t.Fatalf("unmarshal kafka event: %v", err)
	}
	if event.OrderID != order.ID {
		t.Fatalf("expected event order id %q, got %q", order.ID, event.OrderID)
	}
	if event.UserID != order.UserID {
		t.Fatalf("expected event user id %q, got %q", order.UserID, event.UserID)
	}
	if event.Status != string(model.StatusPlaced) {
		t.Fatalf("expected event status %q, got %q", model.StatusPlaced, event.Status)
	}
}

func contextWithUserID() context.Context {
	return middleware.WithUserID(context.Background(), "user-1")
}

type fakeOrderStore struct {
	order         *model.Order
	createErr     error
	statusErr     error
	statusChanges []statusChange
}

type statusChange struct {
	orderID string
	status  model.OrderStatus
}

func (s *fakeOrderStore) CreateOrderWithItems(
	ctx context.Context,
	orderRequest model.CreateOrderRequest,
	items []model.CreateOrderItemRequest,
) (*model.Order, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	return s.order, nil
}

func (s *fakeOrderStore) ChangeOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	s.statusChanges = append(s.statusChanges, statusChange{orderID: orderID, status: status})
	if s.statusErr != nil {
		return s.statusErr
	}
	return nil
}

type fakeWarehouseClient struct {
	response *pb.ReserveStockResponse
	err      error
}

func (c *fakeWarehouseClient) ReserveStock(
	ctx context.Context,
	req *pb.ReserveStockRequest,
	opts ...grpc.CallOption,
) (*pb.ReserveStockResponse, error) {
	return c.response, c.err
}

type fakeEventWriter struct {
	messages []kafka.Message
	err      error
}

func (w *fakeEventWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.messages = append(w.messages, msgs...)
	return w.err
}
