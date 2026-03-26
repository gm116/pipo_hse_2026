package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProxy_StripsPrefix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer upstream.Close()

	proxy, err := newProxy(upstream.URL, "/api/auth")
	if err != nil {
		t.Fatalf("newProxy error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/v1/users/login", nil)
	res := httptest.NewRecorder()
	proxy.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d", res.Code)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "/v1/users/login" {
		t.Fatalf("unexpected upstream path: %s", string(body))
	}
}
