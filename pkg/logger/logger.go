package logger

import (
	"context"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type ctxKey struct{}

func Init(service string) *slog.Logger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("service", service)

	slog.SetDefault(l)
	return l
}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}

	return slog.Default()
}

func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			requestID := chimiddleware.GetReqID(r.Context())

			l := base.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			ctx := WithContext(r.Context(), l)

			next.ServeHTTP(ww, r.WithContext(ctx))

			l.InfoContext(ctx, "http request completed",
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
