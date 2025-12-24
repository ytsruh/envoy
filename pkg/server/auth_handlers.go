package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

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

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Iat    int64  `json:"issued_at"`
	Exp    int64  `json:"expires_at"`
}

func (s *Server) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.dbService.GetQueries().GetUserByEmail(ctx, req.Email)
	if err == nil {
		return sendErrorResponse(c, http.StatusConflict, fmt.Errorf("user with this email already exists"))
	} else if err != sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check existing user"))
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to hash password"))
	}

	userID := uuid.New().String()
	now := time.Now()
	user, err := s.dbService.GetQueries().CreateUser(ctx, database.CreateUserParams{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
		DeletedAt: sql.NullTime{Valid: false},
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create user"))
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate token"))
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

func (s *Server) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.dbService.GetQueries().GetUserByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch user"))
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate token"))
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

func (s *Server) GetProfile(c echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	response := ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Iat:    claims.Iat,
		Exp:    claims.Exp,
	}

	return c.JSON(http.StatusOK, response)
}
