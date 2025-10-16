package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestBuilderDefaults(t *testing.T) {
	t.Parallel()

	builder := NewBuilder(":8080", &mockDBService{})

	if builder.addr != ":8080" {
		t.Errorf("Expected address :8080, got %s", builder.addr)
	}

	if builder.echo == nil {
		t.Error("Expected echo instance to be initialized")
	}

	if builder.timeoutDuration != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", builder.timeoutDuration)
	}

	if builder.securityConfig.XSSProtection != "1; mode=block" {
		t.Error("Expected default security headers config")
	}
}

func TestBuilderWithSecurityHeaders(t *testing.T) {
	t.Parallel()

	customConfig := SecurityHeadersConfig{
		XSSProtection:      "0",
		ContentTypeOptions: "custom",
		FrameOptions:       "SAMEORIGIN",
		ReferrerPolicy:     "no-referrer",
		CSPolicy:           "unsafe-inline",
	}

	builder := NewBuilder(":8080", &mockDBService{}).
		WithSecurityHeaders(customConfig)

	if builder.securityConfig.XSSProtection != "0" {
		t.Error("Expected custom XSSProtection")
	}

	if builder.securityConfig.FrameOptions != "SAMEORIGIN" {
		t.Error("Expected custom FrameOptions")
	}
}

func TestBuilderWithTimeout(t *testing.T) {
	t.Parallel()

	timeout := 60 * time.Second
	builder := NewBuilder(":8080", &mockDBService{}).
		WithTimeout(timeout)

	if builder.timeoutDuration != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, builder.timeoutDuration)
	}
}

func TestBuilderWithMiddleware(t *testing.T) {
	t.Parallel()

	mw1 := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	mw2 := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}

	builder := NewBuilder(":8080", &mockDBService{}).
		WithMiddleware(mw1).
		WithMiddleware(mw2)

	if len(builder.middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(builder.middlewares))
	}
}

func TestBuilderChaining(t *testing.T) {
	t.Parallel()

	builder := NewBuilder(":8080", &mockDBService{}).
		WithTimeout(60 * time.Second).
		WithSecurityHeaders(CustomSecurityHeadersConfig()).
		WithMiddleware(func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		})

	if builder.timeoutDuration != 60*time.Second {
		t.Error("Method chaining failed for timeout")
	}

	if len(builder.middlewares) != 1 {
		t.Error("Method chaining failed for middleware")
	}
}

func TestBuilderBuild(t *testing.T) {
	t.Parallel()

	builder := NewBuilder(":8080", &mockDBService{})
	server, err := builder.Build()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if server == nil {
		t.Error("Expected server instance")
	}

	if server.addr != ":8080" {
		t.Errorf("Expected address :8080, got %s", server.addr)
	}

	if server.echo == nil {
		t.Error("Expected echo instance in server")
	}
}

func TestBuilderBuildWithCustomConfiguration(t *testing.T) {
	t.Parallel()

	customConfig := CustomSecurityHeadersConfig()
	builder := NewBuilder(":9000", &mockDBService{}).
		WithSecurityHeaders(customConfig).
		WithTimeout(45 * time.Second)

	server, err := builder.Build()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if server.addr != ":9000" {
		t.Errorf("Expected address :9000, got %s", server.addr)
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	server.echo.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("Routes should be registered")
	}
}

func CustomSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		XSSProtection:      "0",
		ContentTypeOptions: "custom",
		FrameOptions:       "SAMEORIGIN",
		ReferrerPolicy:     "no-referrer",
		CSPolicy:           "unsafe-inline",
	}
}
