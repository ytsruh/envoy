package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "hello",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
		{
			name:           "goodbye",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   "Goodbye, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			handler := NewGreetingHandler(mockQuerier{})

			var err error
			switch tt.name {
			case "hello":
				err = handler.Hello(rec, req)
			case "goodbye":
				err = handler.Goodbye(rec, req)
			}

			if err != nil {
				t.Errorf("Handler returned error: %v", err)
				return
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

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			handler := NewHealthHandler(mockQuerier{})
			err := handler.Health(rec, req)

			if err != nil {
				t.Errorf("Handler returned error: %v", err)
				return
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			body := make(map[string]string)
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
				return
			}

			if value, ok := body[tt.expectedKey]; !ok || value != tt.expectedValue {
				t.Errorf("Expected %s=%q, got %v", tt.expectedKey, tt.expectedValue, body)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}
