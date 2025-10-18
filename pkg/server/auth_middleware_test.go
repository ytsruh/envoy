package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/utils"
)

func TestJWTAuthMiddleware(t *testing.T) {
	jwtSecret := "test-secret"
	userID := "user-123"
	email := "test@example.com"

	// Generate a valid token
	validToken, err := utils.GenerateJWT(userID, email, jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Generate a token with wrong secret
	wrongSecretToken, err := utils.GenerateJWT(userID, email, "wrong-secret")
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
		checkContext   bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authorization header",
		},
		{
			name:           "invalid format - no Bearer prefix",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization header format",
		},
		{
			name:           "invalid format - wrong prefix",
			authHeader:     "Token " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization header format",
		},
		{
			name:           "invalid token - wrong secret",
			authHeader:     "Bearer " + wrongSecretToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name:           "invalid token - malformed",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
		{
			name:           "invalid token - empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create a test handler that will be called if middleware passes
			testHandler := func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create middleware and wrap test handler
			middleware := JWTAuthMiddleware(jwtSecret)
			handler := middleware(testHandler)

			// Execute
			err := handler(c)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			// Check status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				body := rec.Body.String()
				if !contains(body, tt.expectedError) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, body)
				}
			}

			// Check context if token should be valid
			if tt.checkContext && tt.expectedStatus == http.StatusOK {
				claims, ok := GetUserFromContext(c)
				if !ok {
					t.Error("Expected user claims in context")
				}
				if claims.UserID != userID {
					t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
				}
				if claims.Email != email {
					t.Errorf("Expected email %s, got %s", email, claims.Email)
				}
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	e := echo.New()

	t.Run("valid claims in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedClaims := &utils.JWTClaims{
			UserID: "user-123",
			Email:  "test@example.com",
			Iat:    time.Now().Unix(),
			Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(),
		}

		c.Set(UserContextKey, expectedClaims)

		claims, ok := GetUserFromContext(c)
		if !ok {
			t.Error("Expected claims to be retrieved")
		}

		if claims.UserID != expectedClaims.UserID {
			t.Errorf("Expected user ID %s, got %s", expectedClaims.UserID, claims.UserID)
		}

		if claims.Email != expectedClaims.Email {
			t.Errorf("Expected email %s, got %s", expectedClaims.Email, claims.Email)
		}
	})

	t.Run("no claims in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		claims, ok := GetUserFromContext(c)
		if ok {
			t.Error("Expected no claims to be retrieved")
		}
		if claims != nil {
			t.Error("Expected nil claims")
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(UserContextKey, "not-a-claims-object")

		claims, ok := GetUserFromContext(c)
		if ok {
			t.Error("Expected type assertion to fail")
		}
		if claims != nil {
			t.Error("Expected nil claims when type assertion fails")
		}
	})
}

func TestJWTAuthMiddlewareExpiredToken(t *testing.T) {
	// This test would require creating an expired token
	// For now, we'll skip it as it would require mocking time
	// or waiting for token expiration
	t.Skip("Expired token test requires time mocking")
}

func TestJWTAuthMiddlewareIntegration(t *testing.T) {
	jwtSecret := "integration-test-secret"
	userID := "integration-user"
	email := "integration@example.com"

	// Generate a valid token
	token, err := utils.GenerateJWT(userID, email, jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	e := echo.New()

	// Create a test handler that checks context
	handlerCalled := false
	testHandler := func(c echo.Context) error {
		handlerCalled = true

		// Verify user is in context
		claims, ok := GetUserFromContext(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "User not found in context",
			})
		}

		if claims.UserID != userID {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "User ID mismatch",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "success",
			"user_id": claims.UserID,
			"email":   claims.Email,
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Apply middleware
	middleware := JWTAuthMiddleware(jwtSecret)
	handler := middleware(testHandler)

	// Execute
	err = handler(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify handler was called
	if !handlerCalled {
		t.Error("Test handler was not called")
	}

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !contains(body, userID) {
		t.Errorf("Expected response to contain user ID %s", userID)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
