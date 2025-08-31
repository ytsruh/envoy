package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryMiddleware(t *testing.T) {
	// explicit handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Silence global logger for this test
	old := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(old)

	// Create a new recorder to capture the response
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Apply the middleware directly to the handler under test
	wrapped := RecoveryMiddleware(handler)
	wrapped.ServeHTTP(rec, req)

	// Check the status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Check the response body
	expectedBody := "Internal Server Error"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
}
