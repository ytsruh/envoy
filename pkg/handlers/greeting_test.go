package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGreetingHandlerHello(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "hello_returns_greeting",
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewGreetingHandler(mockQuerier{})
			err := handler.Hello(c)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Body.String() != tt.expectedBody {
				t.Errorf("Expected %q, got %q", tt.expectedBody, rec.Body.String())
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "text/plain; charset=utf-8" && contentType != "text/plain" {
				t.Logf("Content-Type: %s (not validated strictly for text responses)", contentType)
			}
		})
	}
}

func TestGreetingHandlerGoodbye(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "goodbye_returns_farewell",
			expectedStatus: http.StatusOK,
			expectedBody:   "Goodbye, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/goodbye", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewGreetingHandler(mockQuerier{})
			err := handler.Goodbye(c)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Body.String() != tt.expectedBody {
				t.Errorf("Expected %q, got %q", tt.expectedBody, rec.Body.String())
			}
		})
	}
}
