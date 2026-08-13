package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/handler"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/repository"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/service"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to warehouse gRPC", "error", err)
	}
	defer conn.Close()

	warehouseClient := pb.NewWarehouseClient(conn)
	orderDB := repository.NewPostgresOrderRepository(pool)
	orderItemsDB := repository.NewPostgresOrderItemsRepository(pool)
	orderService := service.NewOrderService(*orderDB, *orderItemsDB, warehouseClient)
	orderHandler := handler.NewRouter(*orderService)

	server := http.Server{
		Addr:         ":8082",
		Handler:      orderHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP-server started", "port", "8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", "error", err)
		}
	}()

	<-quit
	logger.Info("Ending work...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error while stopping", "error", err)
	}

	logger.Info("Server stopped")
}
