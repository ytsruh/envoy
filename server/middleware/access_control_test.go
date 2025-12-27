package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/utils"
)

type mockAccessControl struct {
	requireOwnerErr  error
	requireEditorErr error
	requireViewerErr error
	getRoleResult    string
	getRoleErr       error
}

func (m *mockAccessControl) RequireOwner(ctx context.Context, projectID int64, userID string) error {
	return m.requireOwnerErr
}

func (m *mockAccessControl) RequireEditor(ctx context.Context, projectID int64, userID string) error {
	return m.requireEditorErr
}

func (m *mockAccessControl) RequireViewer(ctx context.Context, projectID int64, userID string) error {
	return m.requireViewerErr
}

func (m *mockAccessControl) GetRole(ctx context.Context, projectID int64, userID string) (string, error) {
	return m.getRoleResult, m.getRoleErr
}

func TestRequireProjectOwner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(*echo.Context)
		projectIDParam string
		mockAC         mockAccessControl
		expectedStatus int
	}{
		{
			name: "success - owner access",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user in context",
			setupContext:   func(c *echo.Context) {},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - invalid project ID",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "invalid",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden - not owner",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC: mockAccessControl{
				requireOwnerErr: errors.New("access denied"),
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/projects/1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.projectIDParam)
			tt.setupContext(&c)

			mw := RequireProjectOwner(&tt.mockAC)
			handler := mw(func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			})

			err := handler(c)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestRequireProjectEditor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(*echo.Context)
		projectIDParam string
		mockAC         mockAccessControl
		expectedStatus int
	}{
		{
			name: "success - editor access",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user in context",
			setupContext:   func(c *echo.Context) {},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - invalid project ID",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "abc",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden - not editor",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC: mockAccessControl{
				requireEditorErr: errors.New("access denied"),
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/projects/1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.projectIDParam)
			tt.setupContext(&c)

			mw := RequireProjectEditor(&tt.mockAC)
			handler := mw(func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			})

			err := handler(c)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestRequireProjectViewer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(*echo.Context)
		projectIDParam string
		mockAC         mockAccessControl
		expectedStatus int
	}{
		{
			name: "success - viewer access",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user in context",
			setupContext:   func(c *echo.Context) {},
			projectIDParam: "1",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - invalid project ID",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "xyz",
			mockAC:         mockAccessControl{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden - not viewer",
			setupContext: func(c *echo.Context) {
				(*c).Set(UserContextKey, &utils.JWTClaims{UserID: "user-123"})
			},
			projectIDParam: "1",
			mockAC: mockAccessControl{
				requireViewerErr: errors.New("access denied"),
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/projects/1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.projectIDParam)
			tt.setupContext(&c)

			mw := RequireProjectViewer(&tt.mockAC)
			handler := mw(func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			})

			err := handler(c)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
