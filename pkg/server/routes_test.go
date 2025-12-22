package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
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

type mockProjectHandler struct{}

func (m *mockProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "Test Project"})
}

func (m *mockProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "Test Project"})
}

func (m *mockProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 1, "name": "Test Project"}})
}

func (m *mockProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "Updated Project"})
}

func (m *mockProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "Project deleted successfully"})
}

func TestRegisterProjectHandlers(t *testing.T) {
	t.Parallel()
	s := &Server{
		router:    echo.New(),
		dbService: &mockDBService{},
		addr:      ":8080",
		jwtSecret: "test-secret",
	}

	mockHandler := &mockProjectHandler{}
	s.RegisterProjectHandlers(mockHandler, s.jwtSecret)

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "create project route - with auth",
			method:         http.MethodPost,
			path:           "/projects",
			headers:        map[string]string{"Authorization": "Bearer valid-token"},
			expectedStatus: http.StatusUnauthorized, // JWT will fail with mock token
		},
		{
			name:           "get project route - with auth",
			method:         http.MethodGet,
			path:           "/projects/123",
			headers:        map[string]string{"Authorization": "Bearer valid-token"},
			expectedStatus: http.StatusUnauthorized, // JWT will fail with mock token
		},
		{
			name:           "list projects route - with auth",
			method:         http.MethodGet,
			path:           "/projects",
			headers:        map[string]string{"Authorization": "Bearer valid-token"},
			expectedStatus: http.StatusUnauthorized, // JWT will fail with mock token
		},
		{
			name:           "update project route - with auth",
			method:         http.MethodPut,
			path:           "/projects/123",
			headers:        map[string]string{"Authorization": "Bearer valid-token"},
			expectedStatus: http.StatusUnauthorized, // JWT will fail with mock token
		},
		{
			name:           "delete project route - with auth",
			method:         http.MethodDelete,
			path:           "/projects/123",
			headers:        map[string]string{"Authorization": "Bearer valid-token"},
			expectedStatus: http.StatusUnauthorized, // JWT will fail with mock token
		},
		{
			name:           "create project route - without auth",
			method:         http.MethodPost,
			path:           "/projects",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized, // No JWT header
		},
		{
			name:           "get project route - without auth",
			method:         http.MethodGet,
			path:           "/projects/123",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized, // No JWT header
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			s.router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
