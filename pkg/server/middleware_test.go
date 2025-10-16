package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
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

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        SecurityHeadersConfig
		headerKey     string
		expectedValue string
	}{
		{
			name:          "xss_protection_header",
			config:        DefaultSecurityHeadersConfig(),
			headerKey:     "X-XSS-Protection",
			expectedValue: "1; mode=block",
		},
		{
			name:          "content_type_options_header",
			config:        DefaultSecurityHeadersConfig(),
			headerKey:     "X-Content-Type-Options",
			expectedValue: "nosniff",
		},
		{
			name:          "frame_options_header",
			config:        DefaultSecurityHeadersConfig(),
			headerKey:     "X-Frame-Options",
			expectedValue: "DENY",
		},
		{
			name:          "referrer_policy_header",
			config:        DefaultSecurityHeadersConfig(),
			headerKey:     "Referrer-Policy",
			expectedValue: "strict-origin-when-cross-origin",
		},
		{
			name:          "csp_header",
			config:        DefaultSecurityHeadersConfig(),
			headerKey:     "Content-Security-Policy",
			expectedValue: "default-src 'self'",
		},
		{
			name: "custom_security_config",
			config: SecurityHeadersConfig{
				XSSProtection:      "0",
				ContentTypeOptions: "custom-value",
				FrameOptions:       "SAMEORIGIN",
				ReferrerPolicy:     "no-referrer",
				CSPolicy:           "unsafe-inline",
			},
			headerKey:     "X-Frame-Options",
			expectedValue: "SAMEORIGIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			e.Use(NewSecurityHeadersMiddleware(tt.config))

			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			value := rec.Header().Get(tt.headerKey)
			if value != tt.expectedValue {
				t.Errorf("Expected header %s=%q, got %q", tt.headerKey, tt.expectedValue, value)
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
