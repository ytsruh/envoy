package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler_Docs(t *testing.T) {
	handler := NewDocsHandler()

	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()

	// Call the handler
	err := handler.Docs(w, req)

	// Check no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	// Check that response contains Scalar UI content
	body := w.Body.String()
	if !strings.Contains(body, "scalar") {
		t.Error("Response should contain scalar references")
	}
	if !strings.Contains(body, "createApiReference") {
		t.Error("Response should contain scalar API call")
	}
	if !strings.Contains(body, "Envoy API Documentation") {
		t.Error("Response should contain title")
	}
}

func TestDocsHandler_OpenAPI(t *testing.T) {
	handler := NewDocsHandler()

	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	// Call the handler
	err := handler.OpenAPI(w, req)

	// Check no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	// Check that response contains valid OpenAPI content
	body := w.Body.String()
	if !strings.Contains(body, "openapi") {
		t.Error("Response should contain openapi field")
	}
	if !strings.Contains(body, "Envoy API Documentation") {
		t.Error("Response should contain title")
	}
	if !strings.Contains(body, "3.0.3") {
		t.Error("Response should contain OpenAPI version")
	}
}
