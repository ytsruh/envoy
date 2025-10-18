package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		shouldPanic       bool
		expectedStatus    int
		expectedBodyKey   string
		expectedBodyValue string
	}{
		{
			name:              "panic_recovery",
			shouldPanic:       true,
			expectedStatus:    http.StatusInternalServerError,
			expectedBodyKey:   "error",
			expectedBodyValue: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()

			old := log.Writer()
			log.SetOutput(io.Discard)
			defer log.SetOutput(old)

			e.Use(RecoveryMiddleware())

			if tt.shouldPanic {
				e.GET("/panic", func(c echo.Context) error {
					panic("test panic")
				})
			}

			path := "/panic"
			if !tt.shouldPanic {
				e.GET("/ok", func(c echo.Context) error {
					return c.String(http.StatusOK, "OK")
				})
				path = "/ok"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.shouldPanic {
				var result map[string]string
				body, _ := io.ReadAll(rec.Body)
				if err := json.Unmarshal(body, &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
				if value, ok := result[tt.expectedBodyKey]; !ok || value != tt.expectedBodyValue {
					t.Errorf("Expected %s=%q, got %v", tt.expectedBodyKey, tt.expectedBodyValue, result)
				}
			}
		})
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requests      int
		shouldLimit   bool
		minSuccessful int
	}{
		{
			name:          "within_rate_limit",
			requests:      5,
			shouldLimit:   false,
			minSuccessful: 5,
		},
		{
			name:          "exceed_rate_limit",
			requests:      15,
			shouldLimit:   true,
			minSuccessful: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			rl := NewRateLimiter()
			e.Use(RateLimiterMiddleware(rl))

			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			successCount := 0
			limitedCount := 0

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				rec := httptest.NewRecorder()

				e.ServeHTTP(rec, req)

				if rec.Code == http.StatusOK {
					successCount++
				} else if rec.Code == http.StatusTooManyRequests {
					limitedCount++
				}
			}

			if successCount < tt.minSuccessful {
				t.Errorf("Expected at least %d successful requests, got %d", tt.minSuccessful, successCount)
			}

			if tt.shouldLimit && limitedCount == 0 {
				t.Errorf("Expected rate limit to be hit, but all requests succeeded")
			}

			if !tt.shouldLimit && limitedCount > 0 {
				t.Errorf("Expected no rate limit, but got %d limited requests", limitedCount)
			}
		})
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
