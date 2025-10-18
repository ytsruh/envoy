package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

// MockQuerier implements database.Querier for testing
type MockQuerier struct {
	getUserByEmailFunc func(ctx context.Context, email string) (database.User, error)
	createUserFunc     func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return database.User{}, sql.ErrNoRows
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, arg)
	}
	return database.User{}, nil
}

func (m *MockQuerier) GetUser(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil
}

func (m *MockQuerier) ListUsers(ctx context.Context) ([]database.User, error) {
	return []database.User{}, nil
}

func (m *MockQuerier) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return database.User{}, nil
}

func (m *MockQuerier) DeleteUser(ctx context.Context, arg database.DeleteUserParams) error {
	return nil
}

func (m *MockQuerier) HardDeleteUser(ctx context.Context, id string) error {
	return nil
}

func TestRegister(t *testing.T) {
	jwtSecret := "test-secret"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockQuerier    *MockQuerier
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockQuerier: &MockQuerier{
				getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
					return database.User{}, sql.ErrNoRows
				},
				createUserFunc: func(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
					return database.User{
						ID:        arg.ID,
						Name:      arg.Name,
						Email:     arg.Email,
						Password:  arg.Password,
						CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
						UpdatedAt: arg.UpdatedAt,
						DeletedAt: sql.NullTime{Valid: false},
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var resp AuthResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Token == "" {
					t.Error("Expected token in response")
				}
				if resp.User.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", resp.User.Email)
				}
				if resp.User.Name != "Test User" {
					t.Errorf("Expected name Test User, got %s", resp.User.Name)
				}
			},
		},
		{
			name: "user already exists",
			requestBody: RegisterRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockQuerier: &MockQuerier{
				getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
					return database.User{
						ID:    "existing-id",
						Email: email,
					}, nil
				},
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "User with this email already exists" {
					t.Errorf("Expected conflict error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "missing name",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Name, email, and password are required" {
					t.Errorf("Expected validation error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "missing email",
			requestBody: RegisterRequest{
				Name:     "Test User",
				Password: "password123",
			},
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Name, email, and password are required" {
					t.Errorf("Expected validation error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "short",
			},
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Password must be at least 8 characters" {
					t.Errorf("Expected password length error, got: %s", resp.Error)
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewAuthHandler(tt.mockQuerier, jwtSecret)
			err = handler.Register(c)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestLogin(t *testing.T) {
	jwtSecret := "test-secret"
	hashedPassword, _ := utils.HashPassword("password123")

	tests := []struct {
		name           string
		requestBody    interface{}
		mockQuerier    *MockQuerier
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockQuerier: &MockQuerier{
				getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
					return database.User{
						ID:        "user-123",
						Name:      "Test User",
						Email:     email,
						Password:  hashedPassword,
						CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
						UpdatedAt: time.Now(),
						DeletedAt: sql.NullTime{Valid: false},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp AuthResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Token == "" {
					t.Error("Expected token in response")
				}
				if resp.User.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", resp.User.Email)
				}
				if resp.User.ID != "user-123" {
					t.Errorf("Expected user ID user-123, got %s", resp.User.ID)
				}
			},
		},
		{
			name: "user not found",
			requestBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockQuerier: &MockQuerier{
				getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
					return database.User{}, sql.ErrNoRows
				},
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Invalid email or password" {
					t.Errorf("Expected invalid credentials error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockQuerier: &MockQuerier{
				getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
					return database.User{
						ID:       "user-123",
						Email:    email,
						Password: hashedPassword,
					}, nil
				},
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Invalid email or password" {
					t.Errorf("Expected invalid credentials error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "missing email",
			requestBody: LoginRequest{
				Password: "password123",
			},
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Email and password are required" {
					t.Errorf("Expected validation error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "missing password",
			requestBody: LoginRequest{
				Email: "test@example.com",
			},
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Email and password are required" {
					t.Errorf("Expected validation error, got: %s", resp.Error)
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockQuerier:    &MockQuerier{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewAuthHandler(tt.mockQuerier, jwtSecret)
			err = handler.Login(c)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestLoginTokenValidation(t *testing.T) {
	jwtSecret := "test-secret"
	hashedPassword, _ := utils.HashPassword("password123")

	e := echo.New()
	requestBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockQuerier := &MockQuerier{
		getUserByEmailFunc: func(ctx context.Context, email string) (database.User, error) {
			return database.User{
				ID:        "user-123",
				Name:      "Test User",
				Email:     email,
				Password:  hashedPassword,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{Valid: false},
			}, nil
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewAuthHandler(mockQuerier, jwtSecret)
	_ = handler.Login(c)

	var resp AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate the returned token
	claims, err := utils.ValidateJWT(resp.Token, jwtSecret)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Expected user ID user-123, got %s", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", claims.Email)
	}

	// Check that token expires in 7 days
	expectedExp := time.Now().Add(7 * 24 * time.Hour).Unix()
	tolerance := int64(5)
	if claims.Exp < expectedExp-tolerance || claims.Exp > expectedExp+tolerance {
		t.Errorf("Token expiration not set to 7 days")
	}
}
