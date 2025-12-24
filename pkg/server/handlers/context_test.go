package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

func TestNewHandlerContext(t *testing.T) {
	t.Parallel()

	mockQueries := &database.Queries{}
	jwtSecret := "test-secret"
	accessControl := utils.NewAccessControlService(mockQueries)

	ctx := NewHandlerContext(mockQueries, jwtSecret, accessControl)

	if ctx.Queries != mockQueries {
		t.Error("Expected queries to match mock queries")
	}

	if ctx.JWTSecret != jwtSecret {
		t.Error("Expected JWT secret to match provided secret")
	}

	if ctx.AccessControl != accessControl {
		t.Error("Expected access control to match provided service")
	}
}

func TestSendErrorResponse(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testErr := errors.New("test error")
	err := SendErrorResponse(c, http.StatusBadRequest, testErr)
	if err != nil {
		t.Fatalf("SendErrorResponse returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", resp.Error)
	}
}

func TestNullStringToStringPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected *string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "string input",
			input:    "test string",
			expected: func() *string { s := "test string"; return &s }(),
		},
		{
			name:     "int input",
			input:    123,
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: func() *string { s := ""; return &s }(),
		},
		{
			name:     "bool input",
			input:    true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullStringToStringPtr(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", *result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %v, got nil", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("Expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}
