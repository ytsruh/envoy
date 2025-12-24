package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHealthHandler(t *testing.T) {
	t.Parallel()

	s := New(":8080", &mockDBService{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("expected body to contain 'ok', got %s", rec.Body.String())
	}
}

func TestRegisterDocsHandlers(t *testing.T) {
	t.Parallel()

	s := New(":8080", &mockDBService{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRegisterFaviconHandler(t *testing.T) {
	t.Parallel()

	s := New(":8080", &mockDBService{}, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
