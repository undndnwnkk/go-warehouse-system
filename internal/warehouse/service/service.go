package service

import (
	"context"
	"errors"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/repository"
)

type WarehouseService struct {
	repo *repository.PostgresWarehouseRepository
}

func NewWarehouseService(repo *repository.PostgresWarehouseRepository) *WarehouseService {
	return &WarehouseService{repo: repo}
}

func (s *WarehouseService) CreateStock(ctx context.Context, request repository.CreateStockRequest) (*repository.Stock, error) {
	if request.Quantity <= 0 {
		return nil, ErrInvalidArgument
	}

	stock, err := s.repo.CreateStock(ctx, request)
	if err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *WarehouseService) GetStock(ctx context.Context, sku string) (*repository.Stock, error) {
	stock, err := s.repo.GetStock(ctx, sku)
	if err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *WarehouseService) Reserve(ctx context.Context, sku string, quantity int64) error {
	err := s.repo.Reserve(ctx, sku, quantity)
	if err != nil {
		return err
	}

	return nil
}

var (
	ErrInvalidArgument = errors.New("invalid argument")
)
