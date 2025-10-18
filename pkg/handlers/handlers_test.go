package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
)

type mockQuerier struct{}

func (m mockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m mockQuerier) GetUser(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return database.User{}, nil
}

func (m mockQuerier) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m mockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return []database.User{}, nil
}

func (m mockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func TestGreetingHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		handlerFunc    func(GreetingHandler) echo.HandlerFunc
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "hello",
			method:         http.MethodGet,
			handlerFunc:    func(h GreetingHandler) echo.HandlerFunc { return h.Hello },
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
		{
			name:           "goodbye",
			method:         http.MethodPost,
			handlerFunc:    func(h GreetingHandler) echo.HandlerFunc { return h.Goodbye },
			expectedStatus: http.StatusOK,
			expectedBody:   "Goodbye, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewGreetingHandler(mockQuerier{})
			err := tt.handlerFunc(handler)(c)

			if err != nil {
				t.Errorf("Handler returned error: %v", err)
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

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedKey    string
		expectedValue  string
	}{
		{
			name:           "health_ok",
			expectedStatus: http.StatusOK,
			expectedKey:    "health",
			expectedValue:  "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewHealthHandler(mockQuerier{})
			err := handler.Health(c)

			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			body, _ := io.ReadAll(rec.Body)
			var result map[string]string
			if err := json.Unmarshal(body, &result); err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
			}

			if health, ok := result[tt.expectedKey]; !ok || health != tt.expectedValue {
				t.Errorf("Expected %s=%q, got %v", tt.expectedKey, tt.expectedValue, result)
			}
		})
	}
}
