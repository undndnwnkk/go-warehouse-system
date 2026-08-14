package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	slog.Info("Consumer started work...")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go startConsumer(ctx, &wg, reader)

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

	if err := reader.Close(); err != nil {
		slog.Error("error closing kafka reader", "error", err)
	}
	wg.Wait()
	slog.Info("Consumer has stopped")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error while stopping", "error", err)
	}

	slog.Info("Server stopped")
}

func startConsumer(ctx context.Context, wg *sync.WaitGroup, reader *kafka.Reader) {
	defer wg.Done()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, kafka.ErrGroupClosed) {
				slog.Info("[Consumer] Ending read loop...")
				return
			}
			slog.Error("[Consumer] Error while reading", "error", err)

			time.Sleep(1 * time.Second)
			continue
		}

		processMessage(msg)
	}
}

func processMessage(msg kafka.Message) {
	slog.Info("[Consumer] Message delivered", "Offset", msg.Offset, "Key", string(msg.Key), "Value", string(msg.Value))
}
