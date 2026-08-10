package transport

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/repository"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/service"
)

type WarehouseGrpcServer struct {
	pb.UnimplementedWarehouseServer
	service *service.WarehouseService
}

func NewWarehouseGrpcServer(svc *service.WarehouseService) *WarehouseGrpcServer {
	return &WarehouseGrpcServer{service: svc}
}

func (s *WarehouseGrpcServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	orderID := req.GetOrderId()
	if orderID == "" {
		return nil, status.Error(codes.InvalidArgument, "OrderID must be provided")
	}

	for _, item := range req.GetItems() {
		err := s.service.Reserve(ctx, item.GetSku(), item.GetQuantity())
		if err != nil {
			if errors.Is(err, repository.ErrNotEnough) {
				return nil, status.Error(codes.FailedPrecondition, "not enough stock available")
			}
			return nil, status.Error(codes.Internal, "failed to reserve stock")
		}
	}

	return &pb.ReserveStockResponse{
		Success: true,
		Message: "all reservings were successful",
	}, nil
}
