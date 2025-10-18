package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TestRecoverMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Echo's Recover middleware returns 500 by default
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(2), // Very low limit for testing
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("First request should succeed, got %d", rec1.Code)
	}

	// Second request should succeed
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Second request should succeed, got %d", rec2.Code)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got %d", rec3.Code)
	}
}

func TestGzipMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		acceptEncoding string
		expectGzip     bool
	}{
		{
			name:           "with_gzip_support",
			acceptEncoding: "gzip",
			expectGzip:     true,
		},
		{
			name:           "without_gzip_support",
			acceptEncoding: "",
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			e.Use(middleware.Gzip())

			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, strings.Repeat("Hello World ", 100))
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			hasGzipEncoding := rec.Header().Get("Content-Encoding") == "gzip"
			if tt.expectGzip && !hasGzipEncoding {
				t.Error("Expected gzip encoding, but not found")
			}
			if !tt.expectGzip && hasGzipEncoding {
				t.Error("Expected no gzip encoding, but found it")
			}
		})
	}
}

func TestSecureMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.Secure())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Test that Secure middleware adds expected built-in headers
	// Note: These are the actual defaults from Echo's Secure middleware
	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "SAMEORIGIN", // Echo's default is SAMEORIGIN, not DENY
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := rec.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s=%q, got %q", header, expectedValue, actualValue)
		}
	}

	// Strict-Transport-Security is only set for HTTPS requests
	// so we don't test it here since we're using HTTP
}

func TestBodyLimitMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		bodySize       int
		expectedStatus int
	}{
		{
			name:           "within_limit",
			bodySize:       1024, // 1KB
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exceeds_limit",
			bodySize:       3 * 1024 * 1024, // 3MB (exceeds 2MB limit)
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			e.Use(middleware.BodyLimit("2M"))

			e.POST("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			body := bytes.Repeat([]byte("a"), tt.bodySize)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/octet-stream")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestDecompressMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		compressBody bool
		expectStatus int
	}{
		{
			name:         "gzip_compressed_body",
			compressBody: true,
			expectStatus: http.StatusOK,
		},
		{
			name:         "uncompressed_body",
			compressBody: false,
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			e.Use(middleware.Decompress())

			e.POST("/test", func(c echo.Context) error {
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				return c.String(http.StatusOK, string(body))
			})

			var body io.Reader
			req := httptest.NewRequest(http.MethodPost, "/test", nil)

			if tt.compressBody {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				gz.Write([]byte("Hello World"))
				gz.Close()
				body = &buf
				req.Header.Set("Content-Encoding", "gzip")
			} else {
				body = strings.NewReader("Hello World")
			}

			req.Body = io.NopCloser(body)
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				responseBody := rec.Body.String()
				if responseBody != "Hello World" {
					t.Errorf("Expected response body 'Hello World', got %q", responseBody)
				}
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.RequestID())

	e.GET("/test", func(c echo.Context) error {
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)
		if requestID == "" {
			return c.String(http.StatusInternalServerError, "no request ID")
		}
		return c.String(http.StatusOK, requestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that request ID was added to response headers
	requestID := rec.Header().Get(echo.HeaderXRequestID)
	if requestID == "" {
		t.Error("Expected request ID header to be set")
	}

	// Check that request ID was returned in response body
	if rec.Body.String() != requestID {
		t.Errorf("Expected response body to contain request ID %q, got %q", requestID, rec.Body.String())
	}
}

func TestRemoveTrailingSlashMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test without trailing slash - should work normally
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 for URL without trailing slash, got %d", rec.Code)
	}

	// Test with trailing slash - should redirect
	req2 := httptest.NewRequest(http.MethodGet, "/test/", nil)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusMovedPermanently {
		t.Errorf("Expected redirect status 301, got %d", rec2.Code)
	}

	location := rec2.Header().Get("Location")
	if location != "/test" {
		t.Errorf("Expected redirect to /test, got %q", location)
	}
}
