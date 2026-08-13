package service

import (
	"context"
	"errors"
	"time"

	"github.com/undndnwnkk/go-warehouse-system/internal/order/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/repository"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
)

type OrderService struct {
	ordersDB        repository.PostgresOrderRepository
	orderItemsDB    repository.PostgresOrderItemsRepository
	warehouseClient pb.WarehouseClient
}

func NewOrderService(ordersDB repository.PostgresOrderRepository, orderItemsDB repository.PostgresOrderItemsRepository, wc pb.WarehouseClient) *OrderService {
	return &OrderService{ordersDB: ordersDB, orderItemsDB: orderItemsDB, warehouseClient: wc}
}

func (s *OrderService) CreateOrder(ctx context.Context, items []model.CreateOrderItemRequest) (*model.Order, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized: missing user id")
	}

	orderToSave := model.CreateOrderRequest{UserID: userID, Status: model.StatusPending}
	order, err := s.ordersDB.CreateOrder(ctx, orderToSave)
	if err != nil {
		return nil, err
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
		_ = s.ordersDB.ChangeOrderStatus(ctx, order.ID, model.StatusFailed)
		return order, errors.New("failed to reserve stock in warehouse")
	}

	for _, item := range items {
		_, err = s.orderItemsDB.CreateOrderItem(ctx, model.SaveOrderItemRequest{
			OrderID:  order.ID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		})
		if err != nil {
			return nil, err
		}
	}

	order.Status = model.StatusPlaced
	if err := s.ordersDB.ChangeOrderStatus(ctx, order.ID, model.StatusPlaced); err != nil {
		return nil, err
	}

	// TODO: Kafka message sending

	return order, nil
}
