package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandler_Home(t *testing.T) {
	handler := NewHomeHandler()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call the handler
	err := handler.Home(w, req)
	if err != nil {
		t.Fatalf("Home handler returned error: %v", err)
	}

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	expectedContentType := "text/html; charset=utf-8"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
	}

	// Simple test: just check HTML tag is present
	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Error("Response should contain HTML tag")
	}
}
