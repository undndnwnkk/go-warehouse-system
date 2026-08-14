package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/order/service"
	http_helper "github.com/undndnwnkk/go-warehouse-system/pkg/http"
	"github.com/undndnwnkk/go-warehouse-system/pkg/jwt"
	"github.com/undndnwnkk/go-warehouse-system/pkg/logger"
	"log/slog"
	"net/http"
)

type Handler struct {
	orderService service.OrderService
}

func NewRouter(orderService service.OrderService) http.Handler {
	h := &Handler{orderService: orderService}

	jwtManager := jwt.NewManager("supersecret")

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(logger.RequestLogger(slog.Default()))
	r.Use(chimiddleware.Recoverer)

	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.NewAuthMiddleware(jwtManager))
			r.Post("/", h.createOrderHandler)
		})
	})

	return r
}

func (h *Handler) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateOrderHttpRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http_helper.WriteError(w, http.StatusBadRequest, "invalid arguments:"+err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), request.Items)
	if err != nil {
		http_helper.WriteError(w, http.StatusInternalServerError, "internal server error:"+err.Error())
		return
	}

	http_helper.WriteJSON(w, http.StatusCreated, order)
}
