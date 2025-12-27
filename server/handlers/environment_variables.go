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

func CreateEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("environment_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req CreateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	variable, err := ctx.Queries.CreateEnvironmentVariable(dbCtx, database.CreateEnvironmentVariableParams{
		Key:           req.Key,
		Value:         req.Value,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		EnvironmentID: environmentID,
		CreatedAt:     sql.NullTime{Time: now, Valid: true},
		UpdatedAt:     sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create environment variable"))
	}

	resp := EnvironmentVariableResponse{
		ID:            variable.ID,
		EnvironmentID: variable.EnvironmentID,
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   NullStringToStringPtr(variable.Description),
		CreatedAt:     variable.CreatedAt.Time,
		UpdatedAt:     variable.UpdatedAt.Time,
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	variable, err := ctx.Queries.GetAccessibleEnvironmentVariable(dbCtx, database.GetAccessibleEnvironmentVariableParams{
		ID:      id,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("environment variable not found or access denied"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variable"))
	}

	resp := EnvironmentVariableResponse{
		ID:            variable.ID,
		EnvironmentID: variable.EnvironmentID,
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   NullStringToStringPtr(variable.Description),
		CreatedAt:     variable.CreatedAt.Time,
		UpdatedAt:     variable.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, resp)
}

func ListEnvironmentVariables(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	environmentID, err := strconv.ParseInt(c.Param("environment_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment ID"))
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

	variables, err := ctx.Queries.ListEnvironmentVariablesByEnvironment(dbCtx, environmentID)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variables"))
	}

	var resp []EnvironmentVariableResponse
	for _, v := range variables {
		resp = append(resp, EnvironmentVariableResponse{
			ID:            v.ID,
			EnvironmentID: v.EnvironmentID,
			Key:           v.Key,
			Value:         v.Value,
			Description:   NullStringToStringPtr(v.Description),
			CreatedAt:     v.CreatedAt.Time,
			UpdatedAt:     v.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func UpdateEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req UpdateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	variable, err := ctx.Queries.UpdateEnvironmentVariable(dbCtx, database.UpdateEnvironmentVariableParams{
		Key:         req.Key,
		Value:       req.Value,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          id,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update environment variable"))
	}

	resp := EnvironmentVariableResponse{
		ID:            variable.ID,
		EnvironmentID: variable.EnvironmentID,
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   NullStringToStringPtr(variable.Description),
		CreatedAt:     variable.CreatedAt.Time,
		UpdatedAt:     variable.UpdatedAt.Time,
	}

	return c.JSON(http.StatusOK, resp)
}

func DeleteEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid environment variable ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ctx.Queries.DeleteEnvironmentVariable(dbCtx, id)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment variable"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment variable deleted successfully"})
}
