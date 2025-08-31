package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMethods(t *testing.T) {
	router := NewRouter()
	testPath := "/test"
	testBody := "Test Handler"

	// Define a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testBody))
	})

	// Test GET
	router.Get(testPath, testHandler)
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != testBody {
		t.Errorf("GET %s failed: expected %d %q, got %d %q", testPath, http.StatusOK, testBody, rec.Code, rec.Body.String())
	}

	// Test POST
	router.Post(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPost, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != testBody {
		t.Errorf("POST %s failed: expected %d %q, got %d %q", testPath, http.StatusOK, testBody, rec.Code, rec.Body.String())
	}

	// Test PUT
	router.Put(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPut, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != testBody {
		t.Errorf("PUT %s failed: expected %d %q, got %d %q", testPath, http.StatusOK, testBody, rec.Code, rec.Body.String())
	}

	// Test PATCH
	router.Patch(testPath, testHandler)
	req = httptest.NewRequest(http.MethodPatch, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != testBody {
		t.Errorf("PATCH %s failed: expected %d %q, got %d %q", testPath, http.StatusOK, testBody, rec.Code, rec.Body.String())
	}

	// Test DELETE
	router.Delete(testPath, testHandler)
	req = httptest.NewRequest(http.MethodDelete, testPath, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != testBody {
		t.Errorf("DELETE %s failed: expected %d %q, got %d %q", testPath, http.StatusOK, testBody, rec.Code, rec.Body.String())
	}
}

func TestRouterServeHTTP(t *testing.T) {
	router := NewRouter()
	path := "/serve"
	expectedBody := "ServeHTTP Test"

	router.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}

	// Test a non-existent route
	req = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-existent route, got %d", http.StatusNotFound, rec.Code)
	}
}
