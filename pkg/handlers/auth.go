package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type AuthHandler interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	GetProfile(c echo.Context) error
}

type AuthHandlerImpl struct {
	queries   database.Querier
	jwtSecret string
}

func NewAuthHandler(queries database.Querier, jwtSecret string) AuthHandler {
	return &AuthHandlerImpl{
		queries:   queries,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  UserData `json:"user"`
}

type UserData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandlerImpl) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Basic validation
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Name, email, and password are required"})
	}

	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	_, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "User with this email already exists"})
	} else if err != sql.ErrNoRows {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check existing user"})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to hash password"})
	}

	// Create user
	userID := uuid.New().String()
	now := time.Now()
	user, err := h.queries.CreateUser(ctx, database.CreateUserParams{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
		DeletedAt: sql.NullTime{Valid: false},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
	}

	response := AuthResponse{
		Token: token,
		User: UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *AuthHandlerImpl) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user by email
	user, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
	}

	response := AuthResponse{
		Token: token,
		User: UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return c.JSON(http.StatusOK, response)
}

type ProfileHandler interface {
	GetProfile(c echo.Context) error
}

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Iat    int64  `json:"issued_at"`
	Exp    int64  `json:"expires_at"`
}

func (h *AuthHandlerImpl) GetProfile(c echo.Context) error {
	// Get user claims from context (set by JWT middleware)
	user := c.Get("user")
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized",
		})
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to parse user claims",
		})
	}

	response := ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Iat:    claims.Iat,
		Exp:    claims.Exp,
	}

	return c.JSON(http.StatusOK, response)
}
