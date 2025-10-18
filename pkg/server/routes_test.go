package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/utils"
)

type mockHealthHandler struct {
	called bool
}

func (m *mockHealthHandler) Health(c echo.Context) error {
	m.called = true
	return c.JSON(http.StatusOK, map[string]string{"health": "ok"})
}

type mockGreetingHandler struct {
	helloCalled   bool
	goodbyeCalled bool
}

func (m *mockGreetingHandler) Hello(c echo.Context) error {
	m.helloCalled = true
	return c.String(http.StatusOK, "Hello, World!")
}

func (m *mockGreetingHandler) Goodbye(c echo.Context) error {
	m.goodbyeCalled = true
	return c.String(http.StatusOK, "Goodbye, World!")
}

func TestRegisterHealthHandler(t *testing.T) {
	t.Parallel()
	s := &Server{
		echo:      echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	mockHandler := &mockHealthHandler{}
	s.RegisterHealthHandler(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

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
				echo:      echo.New(),
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

			s.echo.ServeHTTP(rec, req)

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
		echo:      echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
	}

	s.RegisterFaviconHandler()

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
