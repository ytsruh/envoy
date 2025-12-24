package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type CreateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required,max=255"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required,max=255"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentVariableResponse struct {
	ID            int64     `json:"id"`
	EnvironmentID int64     `json:"environment_id"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Description   *string   `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *Server) CreateEnvironmentVariable(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("environment_id"), 10, 64)
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

	var req CreateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	envVar, err := s.dbService.GetQueries().CreateEnvironmentVariable(ctx, database.CreateEnvironmentVariableParams{
		EnvironmentID: environmentID,
		Key:           strings.ToUpper(req.Key),
		Value:         req.Value,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedAt:     sql.NullTime{Time: now, Valid: true},
		UpdatedAt:     sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create environment variable"))
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return c.JSON(http.StatusCreated, response)
}

func (s *Server) GetEnvironmentVariable(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	envVarID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
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

	envVar, err := s.dbService.GetQueries().GetAccessibleEnvironmentVariable(ctx, database.GetAccessibleEnvironmentVariableParams{
		ID:      envVarID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment variable not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variable"))
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) ListEnvironmentVariables(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("environment_id"), 10, 64)
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

	envVars, err := s.dbService.GetQueries().ListEnvironmentVariablesByEnvironment(ctx, environmentID)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variables"))
	}

	var response []EnvironmentVariableResponse
	for _, envVar := range envVars {
		response = append(response, EnvironmentVariableResponse{
			ID:            envVar.ID,
			EnvironmentID: envVar.EnvironmentID,
			Key:           envVar.Key,
			Value:         envVar.Value,
			Description:   nullStringToStringPtr(envVar.Description),
			CreatedAt:     envVar.CreatedAt.Time,
			UpdatedAt:     envVar.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) UpdateEnvironmentVariable(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	envVarID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := s.accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return sendErrorResponse(c, http.StatusForbidden, err)
	}

	var req UpdateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	envVar, err := s.dbService.GetQueries().UpdateEnvironmentVariable(ctx, database.UpdateEnvironmentVariableParams{
		Key:         strings.ToUpper(req.Key),
		Value:       req.Value,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          envVarID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment variable not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update environment variable"))
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) DeleteEnvironmentVariable(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	envVarID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
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

	err = s.dbService.GetQueries().DeleteEnvironmentVariable(ctx, envVarID)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment variable"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment variable deleted successfully"})
}
