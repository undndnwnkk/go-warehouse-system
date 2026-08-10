package main

import (
	"context"
	"errors"
	"log/slog"
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
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found")
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
		slog.Error("failed to create pgxpool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("no connection with PostgreSQL", "error", err)
		os.Exit(1)
	}
	slog.Info("Connection pool pgx created")

	warehouseRepo := repository.NewPostgresWarehouseRepository(pool)
	warehouseSvc := service.NewWarehouseService(warehouseRepo)
	grpcServerImpl := transport.NewWarehouseGrpcServer(warehouseSvc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWarehouseServer(grpcServer, grpcServerImpl)

	go func() {
		slog.Info("gRPC server started", "addr", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("gRPC server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gRPC server...")

	grpcServer.GracefulStop()
	slog.Info("Server stopped successfully")
}
