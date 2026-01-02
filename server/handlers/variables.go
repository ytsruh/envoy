package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/utils"
	shared "ytsruh.com/envoy/shared"
)

type CreateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentVariableResponse struct {
	ID            shared.EnvironmentVariableID `json:"id"`
	EnvironmentID shared.EnvironmentID         `json:"environment_id"`
	Key           string                       `json:"key"`
	Value         string                       `json:"value"`
	Description   *string                      `json:"description"`
	CreatedAt     shared.Timestamp             `json:"created_at"`
	UpdatedAt     shared.Timestamp             `json:"updated_at"`
}

func CreateEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")
	environmentID := c.Param("environment_id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req CreateEnvironmentVariableRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	now := time.Now()
	variableID := utils.GenerateUUID()
	variable, err := ctx.Queries.CreateEnvironmentVariable(dbCtx, database.CreateEnvironmentVariableParams{
		ID:            variableID,
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
		ID:            shared.EnvironmentVariableID(variable.ID),
		EnvironmentID: shared.EnvironmentID(variable.EnvironmentID),
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   shared.NullStringToStringPtr(variable.Description),
		CreatedAt:     shared.FromTime(variable.CreatedAt.Time),
		UpdatedAt:     shared.FromTime(variable.UpdatedAt.Time),
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	id := c.Param("id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
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
		ID:            shared.EnvironmentVariableID(variable.ID),
		EnvironmentID: shared.EnvironmentID(variable.EnvironmentID),
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   shared.NullStringToStringPtr(variable.Description),
		CreatedAt:     shared.FromTime(variable.CreatedAt.Time),
		UpdatedAt:     shared.FromTime(variable.UpdatedAt.Time),
	}

	return c.JSON(http.StatusOK, resp)
}

func ListEnvironmentVariables(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")
	environmentID := c.Param("environment_id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	variables, err := ctx.Queries.ListEnvironmentVariablesByEnvironment(dbCtx, environmentID)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environment variables"))
	}

	var resp []EnvironmentVariableResponse
	for _, v := range variables {
		resp = append(resp, EnvironmentVariableResponse{
			ID:            shared.EnvironmentVariableID(v.ID),
			EnvironmentID: shared.EnvironmentID(v.EnvironmentID),
			Key:           v.Key,
			Value:         v.Value,
			Description:   shared.NullStringToStringPtr(v.Description),
			CreatedAt:     shared.FromTime(v.CreatedAt.Time),
			UpdatedAt:     shared.FromTime(v.UpdatedAt.Time),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func UpdateEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")
	id := c.Param("id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req UpdateEnvironmentVariableRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
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
		ID:            shared.EnvironmentVariableID(variable.ID),
		EnvironmentID: shared.EnvironmentID(variable.EnvironmentID),
		Key:           variable.Key,
		Value:         variable.Value,
		Description:   shared.NullStringToStringPtr(variable.Description),
		CreatedAt:     shared.FromTime(variable.CreatedAt.Time),
		UpdatedAt:     shared.FromTime(variable.UpdatedAt.Time),
	}

	return c.JSON(http.StatusOK, resp)
}

func DeleteEnvironmentVariable(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")
	id := c.Param("id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	err = ctx.Queries.DeleteEnvironmentVariable(dbCtx, id)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment variable"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment variable deleted successfully"})
}
