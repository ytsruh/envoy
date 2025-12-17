package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGreetingHandlerHello(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "hello_returns_greeting",
			expectedStatus:  http.StatusOK,
			expectedMessage: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			rec := httptest.NewRecorder()

			handler := NewGreetingHandler(mockQuerier{})
			err := handler.Hello(rec, req)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal JSON response: %v", err)
			}

			if message, ok := response["message"]; !ok || message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, message)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

func TestGreetingHandlerGoodbye(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "goodbye_returns_farewell",
			expectedStatus:  http.StatusOK,
			expectedMessage: "Goodbye, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/goodbye", nil)
			rec := httptest.NewRecorder()

			handler := NewGreetingHandler(mockQuerier{})
			err := handler.Goodbye(rec, req)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal JSON response: %v", err)
			}

			if message, ok := response["message"]; !ok || message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, message)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}
