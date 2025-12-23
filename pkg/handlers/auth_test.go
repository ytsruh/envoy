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

func (m *MockQuerier) CreateProject(ctx context.Context, arg database.CreateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m *MockQuerier) GetProject(ctx context.Context, id int64) (database.Project, error) {
	return database.Project{}, nil
}

func (m *MockQuerier) ListProjectsByOwner(ctx context.Context, ownerID string) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m *MockQuerier) UpdateProject(ctx context.Context, arg database.UpdateProjectParams) (database.Project, error) {
	return database.Project{}, nil
}

func (m *MockQuerier) DeleteProject(ctx context.Context, arg database.DeleteProjectParams) error {
	return nil
}

// Project sharing methods
func (m *MockQuerier) AddUserToProject(ctx context.Context, arg database.AddUserToProjectParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, nil
}

func (m *MockQuerier) RemoveUserFromProject(ctx context.Context, arg database.RemoveUserFromProjectParams) error {
	return nil
}

func (m *MockQuerier) UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error {
	return nil
}

func (m *MockQuerier) GetProjectUsers(ctx context.Context, projectID int64) ([]database.ProjectUser, error) {
	return []database.ProjectUser{}, nil
}

func (m *MockQuerier) GetUserProjects(ctx context.Context, arg database.GetUserProjectsParams) ([]database.Project, error) {
	return []database.Project{}, nil
}

func (m *MockQuerier) GetProjectMembership(ctx context.Context, arg database.GetProjectMembershipParams) (database.ProjectUser, error) {
	return database.ProjectUser{}, sql.ErrNoRows
}

func (m *MockQuerier) IsProjectOwner(ctx context.Context, arg database.IsProjectOwnerParams) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) GetAccessibleProject(ctx context.Context, arg database.GetAccessibleProjectParams) (database.Project, error) {
	return database.Project{}, sql.ErrNoRows
}

func (m *MockQuerier) CanUserModifyProject(ctx context.Context, arg database.CanUserModifyProjectParams) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) GetProjectMemberRole(ctx context.Context, arg database.GetProjectMemberRoleParams) (string, error) {
	return "", sql.ErrNoRows
}

// Environment methods
func (m *MockQuerier) CreateEnvironment(ctx context.Context, arg database.CreateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQuerier) GetEnvironment(ctx context.Context, id int64) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQuerier) ListEnvironmentsByProject(ctx context.Context, projectID int64) ([]database.Environment, error) {
	return []database.Environment{}, nil
}

func (m *MockQuerier) UpdateEnvironment(ctx context.Context, arg database.UpdateEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQuerier) DeleteEnvironment(ctx context.Context, arg database.DeleteEnvironmentParams) error {
	return nil
}

func (m *MockQuerier) GetAccessibleEnvironment(ctx context.Context, arg database.GetAccessibleEnvironmentParams) (database.Environment, error) {
	return database.Environment{}, nil
}

func (m *MockQuerier) CanUserModifyEnvironment(ctx context.Context, arg database.CanUserModifyEnvironmentParams) (int64, error) {
	return 0, nil
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
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := NewAuthHandler(tt.mockQuerier, jwtSecret)
			err = handler.Register(rec, req)

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
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := NewAuthHandler(tt.mockQuerier, jwtSecret)
			err = handler.Login(rec, req)

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
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := NewAuthHandler(mockQuerier, jwtSecret)
	_ = handler.Login(rec, req)

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

func TestGetProfile(t *testing.T) {
	jwtSecret := "test-secret"
	hashedPassword, _ := utils.HashPassword("password123")
	tests := []struct {
		name           string
		mockQuerier    *MockQuerier
		setupContext   func(*http.Request)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful profile retrieval",
			setupContext: func(r *http.Request) {
				claims := &utils.JWTClaims{
					UserID: "user-123",
					Email:  "test@example.com",
					Iat:    time.Now().Unix(),
					Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(),
				}
				ctx := context.WithValue(r.Context(), "user", claims)
				*r = *r.WithContext(ctx)
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
				var resp ProfileResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.UserID != "user-123" {
					t.Errorf("Expected user ID user-123, got %s", resp.UserID)
				}
				if resp.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", resp.Email)
				}
				if resp.Iat == 0 {
					t.Error("Expected issued_at timestamp to be set")
				}
				if resp.Exp == 0 {
					t.Error("Expected expires_at timestamp to be set")
				}
			},
		},
		{
			name: "missing user in context",
			setupContext: func(r *http.Request) {
				// Don't set user in context
			},
			expectedStatus: http.StatusUnauthorized,
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
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Unauthorized" {
					t.Errorf("Expected Unauthorized error, got: %s", resp.Error)
				}
			},
		},
		{
			name: "invalid type in context",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user", "not-a-claims-object")
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusInternalServerError,
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
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Error != "Failed to parse user claims" {
					t.Errorf("Expected parse error, got: %s", resp.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			rec := httptest.NewRecorder()

			// Setup context
			if tt.setupContext != nil {
				tt.setupContext(req)
			}

			handler := NewAuthHandler(tt.mockQuerier, jwtSecret)
			err := handler.GetProfile(rec, req)

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

func TestProfileResponseStructure(t *testing.T) {
	jwtSecret := "test-secret"
	hashedPassword, _ := utils.HashPassword("password123")

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rec := httptest.NewRecorder()

	now := time.Now().Unix()
	exp := time.Now().Add(7 * 24 * time.Hour).Unix()

	claims := &utils.JWTClaims{
		UserID: "test-user-id",
		Email:  "user@example.com",
		Iat:    now,
		Exp:    exp,
	}
	ctx := context.WithValue(req.Context(), "user", claims)
	req = req.WithContext(ctx)

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
	handler := NewAuthHandler(mockQuerier, jwtSecret)
	err := handler.GetProfile(rec, req)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp ProfileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify all fields are present
	if resp.UserID != "test-user-id" {
		t.Errorf("UserID = %s, want test-user-id", resp.UserID)
	}

	if resp.Email != "user@example.com" {
		t.Errorf("Email = %s, want user@example.com", resp.Email)
	}

	if resp.Iat != now {
		t.Errorf("Iat = %d, want %d", resp.Iat, now)
	}

	if resp.Exp != exp {
		t.Errorf("Exp = %d, want %d", resp.Exp, exp)
	}
}

func TestProfileHandlerWithDifferentUserData(t *testing.T) {
	jwtSecret := "test-secret"
	hashedPassword, _ := utils.HashPassword("password123")

	testCases := []struct {
		name        string
		userID      string
		email       string
		mockQuerier *MockQuerier
	}{
		{
			name:   "user with UUID",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			email:  "uuid@example.com",
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
		},
		{
			name:   "user with special email",
			userID: "user-456",
			email:  "user+tag@example.com",
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
		},
		{
			name:   "user with long ID",
			userID: "very-long-user-id-with-many-characters-1234567890",
			email:  "longid@example.com",
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			rec := httptest.NewRecorder()

			claims := &utils.JWTClaims{
				UserID: tc.userID,
				Email:  tc.email,
				Iat:    time.Now().Unix(),
				Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(),
			}
			ctx := context.WithValue(req.Context(), "user", claims)
			req = req.WithContext(ctx)

			handler := NewAuthHandler(tc.mockQuerier, jwtSecret)
			err := handler.GetProfile(rec, req)

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			var resp ProfileResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if resp.UserID != tc.userID {
				t.Errorf("Expected user ID %s, got %s", tc.userID, resp.UserID)
			}

			if resp.Email != tc.email {
				t.Errorf("Expected email %s, got %s", tc.email, resp.Email)
			}
		})
	}
}
