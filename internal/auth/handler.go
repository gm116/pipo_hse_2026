package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/httpx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Router(metricsHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.Handle("/metrics", metricsHandler)
	mux.HandleFunc("/v1/users/register", h.handleRegister)
	mux.HandleFunc("/v1/users/login", h.handleLogin)
	mux.Handle("/v1/users/me", h.auth(http.HandlerFunc(h.handleMe)))
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var in RegisterInput
	if err := httpx.ReadJSON(r, &in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.service.Register(r.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			httpx.WriteError(w, http.StatusBadRequest, "invalid register payload")
		case errors.Is(err, ErrUserExists):
			httpx.WriteError(w, http.StatusConflict, "email already registered")
		default:
			h.logger.ErrorContext(r.Context(), "register failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var in LoginInput
	if err := httpx.ReadJSON(r, &in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.service.Login(r.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			httpx.WriteError(w, http.StatusBadRequest, "invalid login payload")
		case errors.Is(err, ErrInvalidCredentials):
			httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		default:
			h.logger.ErrorContext(r.Context(), "login failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.service.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		h.logger.ErrorContext(r.Context(), "me failed", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		userID, err := h.service.ParseToken(token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(httpx.ContextWithUserID(r.Context(), userID)))
	})
}

func bearerToken(raw string) (string, bool) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}
