package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/service"
	http_helper "github.com/undndnwnkk/go-warehouse-system/pkg/http"
	"github.com/undndnwnkk/go-warehouse-system/pkg/logger"
	"log/slog"
	"net/http"
)

type Handler struct {
	authService service.AuthService
}

func NewRouter(authService service.AuthService) http.Handler {
	h := &Handler{authService: authService}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(logger.RequestLogger(slog.Default()))
	r.Use(chimiddleware.Recoverer)

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

// createUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Creates a new user and returns an access token and refresh token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.CreateUserRequest	true	"Registration payload"
//	@Success		201		{object}	model.TokenPair
//	@Failure		400		{object}	http_helper.ErrorResponse
//	@Failure		409		{object}	http_helper.ErrorResponse
//	@Failure		500		{object}	http_helper.ErrorResponse
//	@Router			/api/v1/auth/register [post]
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

// loginUserHandler godoc
//
//	@Summary		Log in
//	@Description	Authenticates a user and returns an access token and refresh token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.CreateUserRequest	true	"Login payload"
//	@Success		200		{object}	model.TokenPair
//	@Failure		400		{object}	http_helper.ErrorResponse
//	@Failure		401		{object}	http_helper.ErrorResponse
//	@Failure		500		{object}	http_helper.ErrorResponse
//	@Router			/api/v1/auth/login [post]
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

// refreshTokenHandler godoc
//
//	@Summary		Refresh tokens
//	@Description	Revokes the current refresh token and returns a new token pair.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.RefreshTokenRequest	true	"Refresh token payload"
//	@Success		200		{object}	model.TokenPair
//	@Failure		400		{object}	http_helper.ErrorResponse
//	@Failure		401		{object}	http_helper.ErrorResponse
//	@Failure		500		{object}	http_helper.ErrorResponse
//	@Router			/api/v1/auth/refresh [post]
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
