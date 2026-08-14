package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/repository"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/service"
	"github.com/undndnwnkk/go-warehouse-system/internal/warehouse/transport"
	"github.com/undndnwnkk/go-warehouse-system/pkg/logger"
)

func main() {
	log := logger.Init("warehouse")
	if err := godotenv.Load(); err != nil {
		log.Warn(".env file not found")
	}

	cfg := Config{
		DBUser:     getEnv("DATABASE_USER", "user"),
		DBPassword: getEnv("DATABASE_PASSWORD", "password"),
		DBName:     getEnv("DATABASE_NAME", "warehouse_db"),
		DBURL:      getEnv("DATABASE_URL", "postgres://user:password@localhost:5433/warehouse_db?sslmode=disable"),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Error("failed to create pgxpool", "error", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("no connection with PostgreSQL", "error", err)
		return
	}
	log.Info("Connection pool pgx created")

	warehouseRepo := repository.NewPostgresWarehouseRepository(pool)
	warehouseSvc := service.NewWarehouseService(warehouseRepo)
	grpcServerImpl := transport.NewWarehouseGrpcServer(warehouseSvc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Error("failed to listen", "error", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWarehouseServer(grpcServer, grpcServerImpl)

	go func() {
		log.Info("gRPC server started", "addr", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error("gRPC server error", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down gRPC server...")

	grpcServer.GracefulStop()
	log.Info("Server stopped successfully")
}
