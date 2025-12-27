package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/utils"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockQueries)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockQueries) {
				m.users = make(map[string]database.User)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if resp.Token == "" {
					t.Fatal("expected non-empty token")
				}
				if resp.User.Name != "John Doe" {
					t.Fatalf("expected name 'John Doe', got '%s'", resp.User.Name)
				}
				if resp.User.Email != "john@example.com" {
					t.Fatalf("expected email 'john@example.com', got '%s'", resp.User.Email)
				}
			},
		},
		{
			name: "duplicate email",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockQueries) {
				m.users = map[string]database.User{
					"user-1": {
						ID:    "user-1",
						Email: "john@example.com",
					},
				}
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(resp.Error, "already exists") {
					t.Fatalf("expected error to contain 'already exists', got '%s'", resp.Error)
				}
			},
		},
		{
			name: "invalid email format",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMock:      func(m *MockQueries) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
			},
			setupMock:      func(m *MockQueries) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:           "missing name",
			requestBody:    RegisterRequest{Email: "john@example.com", Password: "password123"},
			setupMock:      func(m *MockQueries) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockQueries{
				projects: []database.Project{},
				users:    make(map[string]database.User),
			}
			tt.setupMock(mock)

			accessControl := utils.NewAccessControlService(mock)
			ctx := NewHandlerContext(mock, "test-secret", accessControl)

			body, _ := json.Marshal(tt.requestBody)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Register(c, ctx)
			if err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	hashedPassword, err := utils.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockQueries)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockQueries) {
				m.users = map[string]database.User{
					"user-1": {
						ID:       "user-1",
						Name:     "John Doe",
						Email:    "john@example.com",
						Password: hashedPassword,
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if resp.Token == "" {
					t.Fatal("expected non-empty token")
				}
				if resp.User.Name != "John Doe" {
					t.Fatalf("expected name 'John Doe', got '%s'", resp.User.Name)
				}
				if resp.User.Email != "john@example.com" {
					t.Fatalf("expected email 'john@example.com', got '%s'", resp.User.Email)
				}
			},
		},
		{
			name: "user not found",
			requestBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockQueries) {
				m.users = make(map[string]database.User)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(resp.Error, "invalid email or password") {
					t.Fatalf("expected error to contain 'invalid email or password', got '%s'", resp.Error)
				}
			},
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockQueries) {
				m.users = map[string]database.User{
					"user-1": {
						ID:       "user-1",
						Name:     "John Doe",
						Email:    "john@example.com",
						Password: hashedPassword,
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(resp.Error, "invalid email or password") {
					t.Fatalf("expected error to contain 'invalid email or password', got '%s'", resp.Error)
				}
			},
		},
		{
			name: "invalid email format",
			requestBody: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMock:      func(m *MockQueries) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockQueries{
				projects: []database.Project{},
				users:    make(map[string]database.User),
			}
			tt.setupMock(mock)

			accessControl := utils.NewAccessControlService(mock)
			ctx := NewHandlerContext(mock, "test-secret", accessControl)

			body, _ := json.Marshal(tt.requestBody)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Login(c, ctx)
			if err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(*echo.Context)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful profile retrieval",
			setupContext: func(c *echo.Context) {
				claims := &utils.JWTClaims{
					UserID: "user-123",
					Email:  "test@example.com",
					Iat:    time.Now().Unix(),
					Exp:    time.Now().Add(time.Hour).Unix(),
				}
				(*c).Set("user", claims)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ProfileResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if resp.UserID != "user-123" {
					t.Fatalf("expected userID 'user-123', got '%s'", resp.UserID)
				}
				if resp.Email != "test@example.com" {
					t.Fatalf("expected email 'test@example.com', got '%s'", resp.Email)
				}
			},
		},
		{
			name:           "no user in context",
			setupContext:   func(c *echo.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(resp.Error, "unauthorized") {
					t.Fatalf("expected error to contain 'unauthorized', got '%s'", resp.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockQueries{
				projects: []database.Project{},
				users:    make(map[string]database.User),
			}
			accessControl := utils.NewAccessControlService(mock)
			ctx := NewHandlerContext(mock, "test-secret", accessControl)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tt.setupContext(&c)

			err := GetProfile(c, ctx)
			if err != nil {
				t.Fatal(err)
			}

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}
