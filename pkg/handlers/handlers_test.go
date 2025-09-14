package handlers

import (
	"context"
	"encoding/json"
	"io"
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

func (m mockQuerier) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m mockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return []database.User{}, nil
}

func (m mockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func TestHelloHandler(t *testing.T) {
	handler := New(mockQuerier{})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.Hello()(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expected := "Hello, World!\n"
	if string(body) != expected {
		t.Errorf("Expected %q, got %q", expected, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGoodbyeHandler(t *testing.T) {
	handler := New(mockQuerier{})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.Goodbye()(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expected := "Goodbye, World!\n"
	if string(body) != expected {
		t.Errorf("Expected %q, got %q", expected, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHealthHandler(t *testing.T) {
	handler := New(mockQuerier{})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.Health()(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if health, ok := result["health"]; !ok || health != "ok" {
		t.Errorf("Expected health status 'ok', got %v", result)
	}
}
