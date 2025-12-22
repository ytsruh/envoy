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
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check content type
	expectedContentType := "text/html; charset=utf-8"
	if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
	}

	// Get response body
	body := w.Body.String()

	// Test that key elements are present
	tests := []struct {
		name     string
		contains string
	}{
		{"DOCTYPE", "<!doctype html>"},
		{"HTML tag", `<html lang="en">`},
		{"Title", "<title>Envoy</title>"},
		{"Tailwind CDN", `https://cdn.tailwindcss.com`},
		{"Welcome heading", "Welcome to Envoy"},
		{"Tailwind working text", "Tailwind CSS is working!"},
		{"Blue button", "Blue Button"},
		{"Green button", "Green Button"},
		{"Purple button", "Purple Button"},
		{"Gradient background", `bg-gradient-to-br from-blue-50 to-indigo-100`},
		{"Container class", `container mx-auto px-4 py-16`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(body, tt.contains) {
				t.Errorf("Expected response body to contain %q, but it didn't", tt.contains)
			}
		})
	}

	// Test that response is not empty
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}
