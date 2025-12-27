package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/utils"
)

func TestDocs(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockQueries := &database.Queries{}
	accessControl := utils.NewAccessControlService(mockQueries)
	ctx := NewHandlerContext(mockQueries, "test-secret", accessControl)

	err := Docs(c, ctx)
	if err != nil {
		t.Fatalf("Docs returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html; charset=UTF-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=UTF-8', got '%s'", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html") && !strings.Contains(body, "<html") {
		t.Error("Expected HTML response body")
	}
}

func TestOpenAPI(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockQueries := &database.Queries{}
	accessControl := utils.NewAccessControlService(mockQueries)
	ctx := NewHandlerContext(mockQueries, "test-secret", accessControl)

	err := OpenAPI(c, ctx)
	if err != nil {
		t.Fatalf("OpenAPI returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	if !strings.HasPrefix(body, "{") && !strings.HasPrefix(body, "[") {
		t.Error("Expected JSON response")
	}
}
