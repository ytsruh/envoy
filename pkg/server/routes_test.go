package server

import (
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

// testHandlers implements Handlers interface for testing route registration
type testHandlers struct {
	hello   func(http.ResponseWriter, *http.Request)
	goodbye func(http.ResponseWriter, *http.Request)
	health  func(http.ResponseWriter, *http.Request)
}

func (t *testHandlers) Hello() http.HandlerFunc {
	return http.HandlerFunc(t.hello)
}

func (t *testHandlers) Goodbye() http.HandlerFunc {
	return http.HandlerFunc(t.goodbye)
}

func (t *testHandlers) Health() http.HandlerFunc {
	return http.HandlerFunc(t.health)
}

func TestRegisterRoutes(t *testing.T) {
	// create a server instance with a mock DB to call the method
	dbservice := &mockDBService{}
	router := NewRouter()
	s := &Server{
		dbService: dbservice,
		router:    &router,
	}

	// Create simple test handlers that just set flags
	helloCalled := false
	goodbyeCalled := false
	healthCalled := false

	testHandler := &testHandlers{
		hello: func(w http.ResponseWriter, r *http.Request) {
			helloCalled = true
			w.WriteHeader(http.StatusOK)
		},
		goodbye: func(w http.ResponseWriter, r *http.Request) {
			goodbyeCalled = true
			w.WriteHeader(http.StatusOK)
		},
		health: func(w http.ResponseWriter, r *http.Request) {
			healthCalled = true
			w.WriteHeader(http.StatusOK)
		},
	}

	s.RegisterRoutes(testHandler)

	t.Run("GET /hello", func(t *testing.T) {
		helloCalled = false
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()

		s.router.ServeHTTP(rec, req)

		if !helloCalled {
			t.Errorf("Expected hello handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /hello, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("POST /goodbye", func(t *testing.T) {
		goodbyeCalled = false
		req := httptest.NewRequest(http.MethodPost, "/goodbye", nil)
		rec := httptest.NewRecorder()

		s.router.ServeHTTP(rec, req)

		if !goodbyeCalled {
			t.Errorf("Expected goodbye handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /goodbye, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("GET /health", func(t *testing.T) {
		healthCalled = false
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		s.router.ServeHTTP(rec, req)

		if !healthCalled {
			t.Errorf("Expected health handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d for /health, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("GET /favicon.ico", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		rec := httptest.NewRecorder()

		s.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status %d for /favicon.ico, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
