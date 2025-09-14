package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMethods(t *testing.T) {
	router := NewRouter()
	testPath := "/test"
	called := false

	// Define a test handler that only sets a flag
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Test GET
	called = false
	router.Get(testPath, testHandler)
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Errorf("GET %s failed: handler not called or wrong status", testPath)
	}

	// Test POST
	called = false
	router.Post(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPost, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Errorf("POST %s failed: handler not called or wrong status", testPath)
	}

	// Test PUT
	called = false
	router.Put(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPut, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Errorf("PUT %s failed: handler not called or wrong status", testPath)
	}

	// Test PATCH
	called = false
	router.Patch(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPatch, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Errorf("PATCH %s failed: handler not called or wrong status", testPath)
	}

	// Test DELETE
	called = false
	router.Delete(testPath, testHandler)
	req = httptest.NewRequest(http.MethodDelete, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Errorf("DELETE %s failed: handler not called or wrong status", testPath)
	}
}

func TestRouterServeHTTP(t *testing.T) {
	router := NewRouter()
	path := "/serve"
	called := false

	router.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !called {
		t.Errorf("Expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test a non-existent route
	called = false
	req = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-existent route, got %d", http.StatusNotFound, rec.Code)
	}
	if called {
		t.Errorf("Handler should not be called for non-existent route")
	}
}
