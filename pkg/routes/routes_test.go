package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	handler := RegisterRoutes()

	// Test /hello GET route
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d for /hello, got %d", http.StatusOK, rec.Code)
	}

	expectedBody := "Hello, World!\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q for /hello, got %q", expectedBody, rec.Body.String())
	}

	// Test /goodbye POST route
	req = httptest.NewRequest(http.MethodPost, "/goodbye", nil)
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d for /goodbye, got %d", http.StatusOK, rec.Code)
	}

	expectedBody = "Goodbye, World!\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q for /goodbye, got %q", expectedBody, rec.Body.String())
	}
}
