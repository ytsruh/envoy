package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockQueries := &database.Queries{}
	accessControl := utils.NewAccessControlService(mockQueries)
	ctx := NewHandlerContext(mockQueries, "test-secret", accessControl)

	err := Health(c, ctx)
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expectedBody := `{"health":"ok"}`
	if rec.Body.String() != expectedBody+"\n" {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
}
