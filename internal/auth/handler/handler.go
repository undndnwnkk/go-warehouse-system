package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/service"
	http_helper "github.com/undndnwnkk/go-warehouse-system/pkg/http"
	"net/http"
)

type Handler struct {
	authService service.AuthService
}

func NewRouter(authService service.AuthService) http.Handler {
	h := &Handler{authService: authService}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http_helper.WriteError(w, http.StatusNotFound, "not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http_helper.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.createUserHandler)
		r.Post("/login", h.loginUserHandler)
		r.Post("/refresh", h.refreshTokenHandler)
	})

	return r
}

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http_helper.WriteError(w, http.StatusBadRequest, "incorrect json: "+err.Error())
		return
	}

	tokens, err := h.authService.Register(r.Context(), request)
	if err != nil {
		// TODO normal error handling
		http_helper.WriteError(w, http.StatusInternalServerError, err.Error())
	}

	http_helper.WriteJSON(w, http.StatusCreated, tokens)
}

func (h *Handler) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var request model.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http_helper.WriteError(w, http.StatusBadRequest, "invalid request:"+err.Error())
		return
	}

	tokens, err := h.authService.Login(r.Context(), request)
	if err != nil {
		// TODO normalize error handling
		http_helper.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http_helper.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var request model.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http_helper.WriteError(w, http.StatusBadRequest, "invalid request:"+err.Error())
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), request.UserID, request.RawRefreshToken)
	if err != nil {
		// TODO normalize error handling
		http_helper.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http_helper.WriteJSON(w, http.StatusOK, tokens)
}
