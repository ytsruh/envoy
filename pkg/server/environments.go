package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
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

func (s *Server) CreateEnvironment(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	var req CreateEnvironmentRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	environment, err := s.dbService.GetQueries().CreateEnvironment(ctx, database.CreateEnvironmentParams{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create environment"))
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusCreated, response)
}

func (s *Server) GetEnvironment(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environment, err := s.dbService.GetQueries().GetAccessibleEnvironment(ctx, database.GetAccessibleEnvironmentParams{
		ID:      environmentID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment"))
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) ListEnvironments(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environments, err := s.dbService.GetQueries().ListEnvironmentsByProject(ctx, projectID)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environments"))
	}

	var response []EnvironmentResponse
	for _, environment := range environments {
		response = append(response, EnvironmentResponse{
			ID:          environment.ID,
			ProjectID:   environment.ProjectID,
			Name:        environment.Name,
			Description: nullStringToStringPtr(environment.Description),
			CreatedAt:   environment.CreatedAt.Time,
			UpdatedAt:   environment.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) UpdateEnvironment(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	var req UpdateEnvironmentRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	environment, err := s.dbService.GetQueries().UpdateEnvironment(ctx, database.UpdateEnvironmentParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          environmentID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update environment"))
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) DeleteEnvironment(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.dbService.GetQueries().DeleteEnvironment(ctx, database.DeleteEnvironmentParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        environmentID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment deleted successfully"})
}
