package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerResponseFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedKey    string
		expectedValue  string
	}{
		{
			name:           "health_endpoint_returns_ok",
			expectedStatus: http.StatusOK,
			expectedKey:    "health",
			expectedValue:  "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			handler := NewHealthHandler(mockQuerier{})
			err := handler.Health(rec, req)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			body, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}

			var result map[string]string
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if value, ok := result[tt.expectedKey]; !ok || value != tt.expectedValue {
				t.Errorf("Expected %s=%q, got %v", tt.expectedKey, tt.expectedValue, result)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}
