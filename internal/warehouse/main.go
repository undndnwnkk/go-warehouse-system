package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	// "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/repository"
	// "github.com/undndnwnkk/go-warehouse-system/internal/warehouse/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := godotenv.Load(); err != nil {
		logger.Error(".env not found")
	}

	cfg := Config{
		DBUser:     getEnv("DATABASE_USER", "user"),
		DBPassword: getEnv("DATABASE_PASSWORD", "password"),
		DBName:     getEnv("DATABASE_NAME", "warehouse_db"),
		DBURL:      getEnv("DATABASE_URL", "postgres://user:password@localhost:5433/warehouse_db?sslmode=disable"),
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		logger.Error("pgxpool.New: ", "error", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("No connection with PostgreSQL: %v", "error", err)
	}
	logger.Info("Connection pool pgx was created")

	// warehouseRepository := repository.NewPostgresWarehouseRepository(pool)
	// warehouseService := service.NewWarehouseService(warehouseRepository)

	server := &http.Server{
		Addr: ":8081",
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP-server started on", "port", ":8080")
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
