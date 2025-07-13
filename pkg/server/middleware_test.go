package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecoveryMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create a new recorder to capture the response
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Apply the middleware
	RecoveryMiddleware(handler).ServeHTTP(rec, req)

	// Check the status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Check the response body
	expectedBody := "Internal Server Error"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	LoggingMiddleware(handler).ServeHTTP(rec, req)

	// We are not checking log output directly in unit tests to avoid global state issues.
	// The primary purpose of this test is to ensure the middleware executes without error
	// and passes control to the next handler.

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	// Test case 1: Handler completes within timeout
	handlerFast := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate fast operation
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	recFast := httptest.NewRecorder()
	reqFast := httptest.NewRequest("GET", "/", nil)
	TimeoutMiddleware(handlerFast).ServeHTTP(recFast, reqFast)

	if recFast.Code != http.StatusOK {
		t.Errorf("Expected status %d for fast handler, got %d", http.StatusOK, recFast.Code)
	}
	if recFast.Body.String() != "OK" {
		t.Errorf("Expected body %q for fast handler, got %q", "OK", recFast.Body.String())
	}

	// Test case 2: Handler exceeds timeout
	handlerSlow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second) // Simulate slow operation (longer than 30s timeout)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	recSlow := httptest.NewRecorder()
	reqSlow := httptest.NewRequest("GET", "/", nil)
	TimeoutMiddleware(handlerSlow).ServeHTTP(recSlow, reqSlow)

	if recSlow.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for slow handler, got %d", http.StatusServiceUnavailable, recSlow.Code)
	}
	if recSlow.Body.String() != "Request timed out" {
		t.Errorf("Expected body %q for slow handler, got %q", "Request timed out", recSlow.Body.String())
	}
}

func TestCorsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test OPTIONS request
	recOptions := httptest.NewRecorder()
	reqOptions := httptest.NewRequest("OPTIONS", "/", nil)
	CorsMiddleware(handler).ServeHTTP(recOptions, reqOptions)

	if recOptions.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, recOptions.Code)
	}
	if recOptions.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", recOptions.Header().Get("Access-Control-Allow-Origin"))
	}
	if recOptions.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods header not set")
	}
	if recOptions.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Access-Control-Allow-Headers header not set")
	}

	// Test GET request
	recGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest("GET", "/", nil)
	CorsMiddleware(handler).ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Errorf("Expected status %d for GET, got %d", http.StatusOK, recGet.Code)
	}
	if recGet.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", recGet.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	SecurityHeadersMiddleware(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection: 1; mode=block, got %s", rec.Header().Get("X-XSS-Protection"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options: nosniff, got %s", rec.Header().Get("X-Content-Type-Options"))
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options: DENY, got %s", rec.Header().Get("X-Frame-Options"))
	}
	if rec.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy: strict-origin-when-cross-origin, got %s", rec.Header().Get("Referrer-Policy"))
	}
	if rec.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("Expected Content-Security-Policy: default-src 'self', got %s", rec.Header().Get("Content-Security-Policy"))
	}
}

func TestApplyMiddleware(t *testing.T) {
	// Middleware that adds a header
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Middleware that adds another header
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final Handler"))
	})

	// Apply middlewares
	chainedHandler := ApplyMiddleware(finalHandler, middleware1, middleware2)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	chainedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("X-Middleware-1") != "true" {
		t.Error("Middleware 1 header not set")
	}
	if rec.Header().Get("X-Middleware-2") != "true" {
		t.Error("Middleware 2 header not set")
	}
	if rec.Body.String() != "Final Handler" {
		t.Errorf("Expected body %q, got %q", "Final Handler", rec.Body.String())
	}
}

func TestRegisterMiddleware(t *testing.T) {
	// This test verifies that registerMiddleware returns a handler that applies
	// the expected behavior of the chained middlewares.
	// We'll check for a few headers that are set by the default middlewares.

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	chainedHandler := registerMiddleware(finalHandler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	chainedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check for headers set by SecurityHeadersMiddleware
	if rec.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("SecurityHeadersMiddleware header X-XSS-Protection not set correctly")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("SecurityHeadersMiddleware header X-Content-Type-Options not set correctly")
	}

	// Check for headers set by CorsMiddleware
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CorsMiddleware header Access-Control-Allow-Origin not set correctly")
	}
}
