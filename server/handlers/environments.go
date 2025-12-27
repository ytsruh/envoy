package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/middleware"
	"ytsruh.com/envoy/server/utils"
)

type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentResponse struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateEnvironment(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req CreateEnvironmentRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	environment, err := ctx.Queries.CreateEnvironment(dbCtx, database.CreateEnvironmentParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		ProjectID:   projectID,
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create environment"))
	}

	resp := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: NullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetEnvironment(c echo.Context, ctx *HandlerContext) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environment, err := ctx.Queries.GetAccessibleEnvironment(dbCtx, database.GetAccessibleEnvironmentParams{
		ID:      id,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment not found or access denied"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment"))
	}

	resp := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: NullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, resp)
}

func ListEnvironments(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := ctx.AccessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environments, err := ctx.Queries.ListEnvironmentsByProject(dbCtx, projectID)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environments"))
	}

	var resp []EnvironmentResponse
	for _, env := range environments {
		resp = append(resp, EnvironmentResponse{
			ID:          env.ID,
			ProjectID:   env.ProjectID,
			Name:        env.Name,
			Description: NullStringToStringPtr(env.Description),
			CreatedAt:   env.CreatedAt.Time,
			UpdatedAt:   env.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func UpdateEnvironment(c echo.Context, ctx *HandlerContext) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	_, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	var req UpdateEnvironmentRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	environment, err := ctx.Queries.UpdateEnvironment(dbCtx, database.UpdateEnvironmentParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          id,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update environment"))
	}

	resp := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: NullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, resp)
}

func DeleteEnvironment(c echo.Context, ctx *HandlerContext) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	_, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ctx.Queries.DeleteEnvironment(dbCtx, database.DeleteEnvironmentParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment deleted successfully"})
}
