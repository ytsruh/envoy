package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
	Login(w http.ResponseWriter, r *http.Request) error
	GetProfile(w http.ResponseWriter, r *http.Request) error
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

func (h *AuthHandlerImpl) Register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrorResponse{Error: "Invalid request body"}.Error, http.StatusBadRequest)
		return nil
	}

	// Basic validation
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, ErrorResponse{Error: "Name, email, and password are required"}.Error, http.StatusBadRequest)
		return nil
	}

	if len(req.Password) < 8 {
		http.Error(w, ErrorResponse{Error: "Password must be at least 8 characters"}.Error, http.StatusBadRequest)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	_, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		http.Error(w, ErrorResponse{Error: "User with this email already exists"}.Error, http.StatusConflict)
		return nil
	} else if err != sql.ErrNoRows {
		http.Error(w, ErrorResponse{Error: "Failed to check existing user"}.Error, http.StatusInternalServerError)
		return nil
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, ErrorResponse{Error: "Failed to hash password"}.Error, http.StatusInternalServerError)
		return nil
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
		http.Error(w, ErrorResponse{Error: "Failed to create user"}.Error, http.StatusInternalServerError)
		return nil
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		http.Error(w, ErrorResponse{Error: "Failed to generate token"}.Error, http.StatusInternalServerError)
		return nil
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func (h *AuthHandlerImpl) Login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrorResponse{Error: "Invalid request body"}.Error, http.StatusBadRequest)
		return nil
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		http.Error(w, ErrorResponse{Error: "Email and password are required"}.Error, http.StatusBadRequest)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user by email
	user, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, ErrorResponse{Error: "Invalid email or password"}.Error, http.StatusUnauthorized)
		return nil
	} else if err != nil {
		http.Error(w, ErrorResponse{Error: "Failed to fetch user"}.Error, http.StatusInternalServerError)
		return nil
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, ErrorResponse{Error: "Invalid email or password"}.Error, http.StatusUnauthorized)
		return nil
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		http.Error(w, ErrorResponse{Error: "Failed to generate token"}.Error, http.StatusInternalServerError)
		return nil
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Iat    int64  `json:"issued_at"`
	Exp    int64  `json:"expires_at"`
}

func (h *AuthHandlerImpl) GetProfile(w http.ResponseWriter, r *http.Request) error {
	// Get user claims from context (set by JWT middleware)
	// Note: This will need to be handled by server package middleware
	user := r.Context().Value("user")
	if user == nil {
		http.Error(w, ErrorResponse{Error: "Unauthorized"}.Error, http.StatusUnauthorized)
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		http.Error(w, ErrorResponse{Error: "Failed to parse user claims"}.Error, http.StatusInternalServerError)
		return nil
	}

	response := ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Iat:    claims.Iat,
		Exp:    claims.Exp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}
