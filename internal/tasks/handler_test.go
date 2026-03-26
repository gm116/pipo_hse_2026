package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/logging"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/metrics"
)

func TestHandler_CRUDFlow(t *testing.T) {
	repo := newMemoryTaskRepo()
	tokens := jwt.NewManager("secret", "auth-service", time.Hour)
	svc := NewService(repo)
	h := NewHandler(svc, tokens, logging.New("ERROR"))
	router := h.Router(metrics.New("test").Handler())
	token, _ := tokens.Issue(7)

	createBody := map[string]any{"title": "Plan sprint", "description": "backend work", "status": "todo"}
	createRaw, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(createRaw))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRes.Code, createRes.Body.String())
	}

	var created Task
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRes.Code, listRes.Body.String())
	}

	updateBody := map[string]any{"title": "Plan sprint updated", "description": "done", "status": "done"}
	updateRaw, _ := json.Marshal(updateBody)
	updatePath := "/v1/tasks/" + strconv.FormatInt(created.ID, 10)
	updateReq := httptest.NewRequest(http.MethodPut, updatePath, bytes.NewReader(updateRaw))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updateRes.Code, updateRes.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, updatePath, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRes := httptest.NewRecorder()
	router.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleteRes.Code, deleteRes.Body.String())
	}
}
