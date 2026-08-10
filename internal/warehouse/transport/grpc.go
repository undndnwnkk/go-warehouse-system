package transport

import (
	"context"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WarehouseGrpcServer struct {
	pb.UnimplementedWarehouseServer
	service service.WarehouseService
}

func (s *WarehouseGrpcServer) ReserveStock(ctx context.Context, req pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	orderID := req.GetOrderId()
	if orderID == "" {
		return nil, status.Error(codes.InvalidArgument, "OrderID must be not null")
	}

	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, "client cancel request")
	}

	for _, item := range req.Items {
		err := s.service.Reserve(ctx, item.Sku, item.Quantity)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, "error while reserving")
		}
	}

	return &pb.ReserveStockResponse{
		Success: true,
		Message: "all reservings was successfull",
	}, nil
}
