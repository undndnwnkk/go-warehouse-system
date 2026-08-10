package service

import (
	"context"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/repository"
)

type WarehouseService struct {
	repo *repository.PostgresWarehouseRepository
}

func NewWarehouseService(repo *repository.PostgresWarehouseRepository) *WarehouseService {
	return &WarehouseService{repo: repo}
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
