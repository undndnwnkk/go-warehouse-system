package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"github.com/undndnwnkk/go-warehouse-system/pkg/logger"
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
	log := logger.Init("notification")
	server := http.Server{Addr: ":8083"}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAdress},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Info("Consumer started work...")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go startConsumer(ctx, &wg, reader, log)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Info("Error while reading message: " + err.Error())
			break
		}

		log.Info("got: " + string(msg.Value))
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

	if err := reader.Close(); err != nil {
		log.Error("error closing kafka reader", "error", err)
	}
	wg.Wait()
	log.Info("Consumer has stopped")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Error while stopping", "error", err)
	}

	log.Info("Server stopped")
}

func startConsumer(ctx context.Context, wg *sync.WaitGroup, reader *kafka.Reader, log *slog.Logger) {
	defer wg.Done()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, kafka.ErrGroupClosed) {
				log.Info("[Consumer] Ending read loop...")
				return
			}
			log.Error("[Consumer] Error while reading", "error", err)

			time.Sleep(1 * time.Second)
			continue
		}

		processMessage(msg, log)
	}
}

func processMessage(msg kafka.Message, log *slog.Logger) {
	log.Info("[Consumer] Message delivered", "Offset", msg.Offset, "Key", string(msg.Key), "Value", string(msg.Value))
}
