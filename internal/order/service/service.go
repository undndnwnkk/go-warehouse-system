package service

import (
	"context"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/repository"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"time"
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
	userID := ctx.Value("userID").(string)
	orderToSave := model.CreateOrderRequest{UserID: userID, Status: model.StatusPending}
	order, err := s.ordersDB.CreateOrder(ctx, orderToSave)
	if err != nil {
		return nil, err
	}

	grpcItems := make([]*pb.OrderItem, len(items))
	for i := range items {
		current := &pb.OrderItem{Sku: items[i].SKU, Quantity: items[i].Quantity}
		grpcItems[i] = current
	}

	grpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	grpcResponse, err := s.warehouseClient.ReserveStock(
		grpcCtx,
		&pb.ReserveStockRequest{OrderId: order.ID, Items: grpcItems})

	if err != nil || !grpcResponse.Success {
		order.Status = model.StatusFailed
		err = s.ordersDB.ChangeOrderStatus(ctx, order.ID, model.StatusFailed)
		if err != nil {
			return nil, err
		}
	} else {
		order.Status = model.StatusPlaced
		err = s.ordersDB.ChangeOrderStatus(ctx, order.ID, model.StatusPlaced)
		if err != nil {
			return nil, err
		}
		// Kafka sending
	}
	return order, nil

}
