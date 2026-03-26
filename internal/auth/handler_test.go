package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/logging"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/metrics"
)

func TestHandler_RegisterLoginMe(t *testing.T) {
	repo := newMemoryRepo()
	svc := NewService(repo, jwt.NewManager("secret", "test", time.Hour))
	h := NewHandler(svc, logging.New("ERROR"))
	router := h.Router(metrics.New("test").Handler())

	registerBody := map[string]string{
		"email":    "demo@example.com",
		"name":     "Demo",
		"password": "123456",
	}
	regRaw, _ := json.Marshal(registerBody)

	regReq := httptest.NewRequest(http.MethodPost, "/v1/users/register", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	regRes := httptest.NewRecorder()
	router.ServeHTTP(regRes, regReq)

	if regRes.Code != http.StatusCreated {
		t.Fatalf("unexpected register status: %d", regRes.Code)
	}

	var regResp AuthResponse
	if err := json.Unmarshal(regRes.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if regResp.Token == "" {
		t.Fatal("empty register token")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+regResp.Token)
	meRes := httptest.NewRecorder()
	router.ServeHTTP(meRes, meReq)

	if meRes.Code != http.StatusOK {
		t.Fatalf("unexpected me status: %d body=%s", meRes.Code, meRes.Body.String())
	}
}
