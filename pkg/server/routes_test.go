package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type mockHealthHandler struct {
	called bool
}

func (m *mockHealthHandler) Health(w http.ResponseWriter, r *http.Request) error {
	m.called = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"health": "ok"})
}

func TestRegisterHealthHandler(t *testing.T) {
	t.Parallel()
	s := &Server{
		router:    echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	mockHandler := &mockHealthHandler{}
	s.RegisterHealthHandler(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if !mockHandler.called {
		t.Error("Expected health handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRegisterFaviconHandler(t *testing.T) {
	t.Parallel()
	s := &Server{
		router:    echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	s.RegisterFaviconHandler()

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
