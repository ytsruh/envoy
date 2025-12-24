package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/server/middleware"
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

func Register(c echo.Context, ctx *HandlerContext) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ctx.Queries.GetUserByEmail(dbCtx, req.Email)
	if err == nil {
		return SendErrorResponse(c, http.StatusConflict, fmt.Errorf("user with this email already exists"))
	} else if err != sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check existing user"))
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to hash password"))
	}

	userID := uuid.New().String()
	now := time.Now()
	user, err := ctx.Queries.CreateUser(dbCtx, database.CreateUserParams{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
		DeletedAt: sql.NullTime{Valid: false},
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create user"))
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, ctx.JWTSecret)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate token"))
	}

	authResp := AuthResponse{
		Token: token,
		User: UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return c.JSON(http.StatusCreated, authResp)
}

func Login(c echo.Context, ctx *HandlerContext) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ctx.Queries.GetUserByEmail(dbCtx, req.Email)
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch user"))
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, ctx.JWTSecret)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate token"))
	}

	authResp := AuthResponse{
		Token: token,
		User: UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time,
		},
	}

	return c.JSON(http.StatusOK, authResp)
}

func GetProfile(c echo.Context, ctx *HandlerContext) error {
	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	profileResp := ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Iat:    claims.Iat,
		Exp:    claims.Exp,
	}

	return c.JSON(http.StatusOK, profileResp)
}
