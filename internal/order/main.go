package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/handler"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/repository"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/service"
	pb "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/pb"
	"github.com/undndnwnkk/go-warehouse-system/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"
)

const (
	topic        = "order_events"
	brokerAdress = "localhost:29092"
)

func main() {
	log := logger.Init("order")
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

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to warehouse gRPC", "error", err)
	}
	defer conn.Close()

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokerAdress),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	warehouseClient := pb.NewWarehouseClient(conn)
	orderStore := repository.NewPostgresOrderStore(pool)
	orderService := service.NewOrderService(orderStore, warehouseClient, writer)
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
		log.Info("HTTP-server started", "port", "8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server error", "error", err)
		}
	}()

	<-quit
	log.Info("Ending work...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Error while stopping", "error", err)
	}

	log.Info("Server stopped")
}
