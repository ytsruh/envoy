package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockRouter implements the Router interface for testing.
type mockRouter struct {
	routes map[string]http.HandlerFunc
}

func newMockRouter() *mockRouter {
	return &mockRouter{
		routes: make(map[string]http.HandlerFunc),
	}
}

func (m *mockRouter) Get(path string, handler http.HandlerFunc) {
	m.routes["GET "+path] = handler
}

func (m *mockRouter) Post(path string, handler http.HandlerFunc) {
	m.routes["POST "+path] = handler
}

func (m *mockRouter) Put(pattern string, handler http.HandlerFunc) {
	m.routes["PUT "+pattern] = handler
}

func (m *mockRouter) Patch(pattern string, handler http.HandlerFunc) {
	m.routes["PATCH "+pattern] = handler
}

func (m *mockRouter) Delete(pattern string, handler http.HandlerFunc) {
	m.routes["DELETE "+pattern] = handler
}

func (m *mockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + " " + r.URL.Path
	if handler, ok := m.routes[key]; ok {
		handler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// mockDBService implements the DBService interface for testing.
type mockDBService struct{}

func (m *mockDBService) Health() map[string]string {
	return map[string]string{"status": "ok", "message": "mock health check"}
}

func TestRegisterRoutes(t *testing.T) {
	router := newMockRouter()
	RegisterRoutes(router, &mockDBService{})

	t.Run("GET /hello", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /hello, got %d", http.StatusOK, rec.Code)
		}

		expectedBody := "Hello, World!\n"
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected body %q for /hello, got %q", expectedBody, rec.Body.String())
		}
	})

	t.Run("POST /goodbye", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/goodbye", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /goodbye, got %d", http.StatusOK, rec.Code)
		}

		expectedBody := "Goodbye, World!\n"
		if rec.Body.String() != expectedBody {
			t.Errorf("Expected body %q for /goodbye, got %q", expectedBody, rec.Body.String())
		}
	})

	t.Run("GET /health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /health, got %d", http.StatusOK, rec.Code)
		}

		expectedHealth := map[string]string{"status": "ok", "message": "mock health check"}
		var actualHealth map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &actualHealth)
		if err != nil {
			t.Fatalf("Failed to unmarshal health response: %v", err)
		}

		if actualHealth["status"] != expectedHealth["status"] || actualHealth["message"] != expectedHealth["message"] {
			t.Errorf("Expected health %v, got %v", expectedHealth, actualHealth)
		}
	})
}
