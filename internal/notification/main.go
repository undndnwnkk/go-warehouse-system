package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	topic        = "order_events"
	brokerAdress = "localhost:29092"
	groupID      = "notification"
)

func main() {
	server := http.Server{Addr: ":8083"}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAdress},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	slog.Info("Consumer started work...")

	ctx := context.Background()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			slog.Info("Error while reading message: " + err.Error())
			break
		}

		slog.Info("got: " + string(msg.Value))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("HTTP-server started", "port", "8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	<-quit
	slog.Info("Ending work...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error while stopping", "error", err)
	}

	slog.Info("Server stopped")
}
