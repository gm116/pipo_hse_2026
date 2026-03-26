package tasks

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/httpx"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
)

type Handler struct {
	service *Service
	tokens  *jwt.Manager
	logger  *slog.Logger
}

func NewHandler(service *Service, tokens *jwt.Manager, logger *slog.Logger) *Handler {
	return &Handler{service: service, tokens: tokens, logger: logger}
}

func (h *Handler) Router(metricsHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.Handle("/metrics", metricsHandler)
	mux.Handle("/v1/tasks", h.auth(http.HandlerFunc(h.handleTasksCollection)))
	mux.Handle("/v1/tasks/", h.auth(http.HandlerFunc(h.handleTaskItem)))
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleTasksCollection(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		tasks, err := h.service.ListTasks(r.Context(), userID)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "list tasks failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": tasks})
	case http.MethodPost:
		var in CreateTaskInput
		if err := httpx.ReadJSON(r, &in); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		task, err := h.service.CreateTask(r.Context(), userID, in)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidInput):
				httpx.WriteError(w, http.StatusBadRequest, "invalid task payload")
			default:
				h.logger.ErrorContext(r.Context(), "create task failed", "err", err)
				httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, task)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleTaskItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, ok := parseTaskID(r.URL.Path)
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := h.service.GetTask(r.Context(), userID, taskID)
		if err != nil {
			h.handleTaskError(w, r, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, task)
	case http.MethodPut:
		var in UpdateTaskInput
		if err := httpx.ReadJSON(r, &in); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		task, err := h.service.UpdateTask(r.Context(), userID, taskID, in)
		if err != nil {
			h.handleTaskError(w, r, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, task)
	case http.MethodDelete:
		err := h.service.DeleteTask(r.Context(), userID, taskID)
		if err != nil {
			h.handleTaskError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleTaskError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.WriteError(w, http.StatusBadRequest, "invalid task payload")
	case errors.Is(err, ErrTaskNotFound):
		httpx.WriteError(w, http.StatusNotFound, "task not found")
	default:
		h.logger.ErrorContext(r.Context(), "task request failed", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		claims, err := h.tokens.Parse(token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(httpx.ContextWithUserID(r.Context(), claims.Subject)))
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

func parseTaskID(path string) (int64, bool) {
	trimmed := strings.TrimPrefix(path, "/v1/tasks/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
