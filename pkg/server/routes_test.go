package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/utils"
)

type mockHealthHandler struct {
	called bool
}

func (m *mockHealthHandler) Health(w http.ResponseWriter, r *http.Request) error {
	m.called = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"health": "ok"})
}

type mockGreetingHandler struct {
	helloCalled   bool
	goodbyeCalled bool
}

func (m *mockGreetingHandler) Hello(w http.ResponseWriter, r *http.Request) error {
	m.helloCalled = true
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
	return nil
}

func (m *mockGreetingHandler) Goodbye(w http.ResponseWriter, r *http.Request) error {
	m.goodbyeCalled = true
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Goodbye, World!"))
	return nil
}

func TestRegisterHealthHandler(t *testing.T) {
	t.Parallel()
	s := &Server{
		router:    echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	mockHandler := &mockHealthHandler{}
	s.RegisterHealthHandler(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if !mockHandler.called {
		t.Error("Expected health handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRegisterGreetingHandlers(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-greeting-handlers"
	userID := "test-user-123"
	email := "test@example.com"

	// Generate a valid JWT token
	token, err := utils.GenerateJWT(userID, email, jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		handlerCheck   func(*mockGreetingHandler) bool
	}{
		{
			name:           "hello",
			method:         http.MethodGet,
			path:           "/hello",
			expectedStatus: http.StatusOK,
			handlerCheck:   func(m *mockGreetingHandler) bool { return m.helloCalled },
		},
		{
			name:           "goodbye",
			method:         http.MethodPost,
			path:           "/goodbye",
			expectedStatus: http.StatusOK,
			handlerCheck:   func(m *mockGreetingHandler) bool { return m.goodbyeCalled },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Server{
				router:    echo.New(),
				dbService: &mockDBService{},
				addr:      ":8080",
				jwtSecret: jwtSecret,
			}

			mockHandler := &mockGreetingHandler{}
			s.RegisterGreetingHandlers(mockHandler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			// Add JWT token to Authorization header
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()

			s.router.ServeHTTP(rec, req)

			if !tt.handlerCheck(mockHandler) {
				t.Error("Expected handler to be called")
			}
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestRegisterFaviconHandler(t *testing.T) {
	t.Parallel()
	s := &Server{
		router:    echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	s.RegisterFaviconHandler()

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
