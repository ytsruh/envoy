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

type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentResponse struct {
	ID          shared.EnvironmentID `json:"id"`
	ProjectID   shared.ProjectID     `json:"project_id"`
	Name        string               `json:"name"`
	Description *string              `json:"description"`
	CreatedAt   shared.Timestamp     `json:"created_at"`
	UpdatedAt   shared.Timestamp     `json:"updated_at"`
}

func CreateEnvironment(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	var req CreateEnvironmentRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	now := time.Now()
	environmentID := utils.GenerateUUID()
	environment, err := ctx.Queries.CreateEnvironment(dbCtx, database.CreateEnvironmentParams{
		ID:          environmentID,
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
		ID:          shared.EnvironmentID(environment.ID),
		ProjectID:   shared.ProjectID(environment.ProjectID),
		Name:        environment.Name,
		Description: shared.NullStringToStringPtr(environment.Description),
		CreatedAt:   shared.FromTime(environment.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(environment.UpdatedAt.Time),
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetEnvironment(c echo.Context, ctx *HandlerContext) error {
	id := c.Param("id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
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
		ID:          shared.EnvironmentID(environment.ID),
		ProjectID:   shared.ProjectID(environment.ProjectID),
		Name:        environment.Name,
		Description: shared.NullStringToStringPtr(environment.Description),
		CreatedAt:   shared.FromTime(environment.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(environment.UpdatedAt.Time),
	}

	return c.JSON(http.StatusOK, resp)
}

func ListEnvironments(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("project_id")

	claims, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	if err := ctx.AccessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
		return SendErrorResponse(c, http.StatusForbidden, err)
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	environments, err := ctx.Queries.ListEnvironmentsByProject(dbCtx, projectID)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch environments"))
	}

	var resp []EnvironmentResponse
	for _, env := range environments {
		resp = append(resp, EnvironmentResponse{
			ID:          shared.EnvironmentID(env.ID),
			ProjectID:   shared.ProjectID(env.ProjectID),
			Name:        env.Name,
			Description: shared.NullStringToStringPtr(env.Description),
			CreatedAt:   shared.FromTime(env.CreatedAt.Time),
			UpdatedAt:   shared.FromTime(env.UpdatedAt.Time),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func UpdateEnvironment(c echo.Context, ctx *HandlerContext) error {
	id := c.Param("id")

	_, err := GetUserOrUnauthorized(c)
	if err != nil {
		return err
	}

	var req UpdateEnvironmentRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	now := time.Now()
	updatedEnvironment, err := ctx.Queries.UpdateEnvironment(dbCtx, database.UpdateEnvironmentParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          id,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update environment"))
	}

	resp := EnvironmentResponse{
		ID:          shared.EnvironmentID(updatedEnvironment.ID),
		ProjectID:   shared.ProjectID(updatedEnvironment.ProjectID),
		Name:        updatedEnvironment.Name,
		Description: shared.NullStringToStringPtr(updatedEnvironment.Description),
		CreatedAt:   shared.FromTime(updatedEnvironment.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(updatedEnvironment.UpdatedAt.Time),
	}

	return c.JSON(http.StatusOK, resp)
}

func DeleteEnvironment(c echo.Context, ctx *HandlerContext) error {
	id := c.Param("id")

	if _, err := GetUserOrUnauthorized(c); err != nil {
		return err
	}

	dbCtx, cancel := GetDBContext()
	defer cancel()

	err := ctx.Queries.DeleteEnvironment(dbCtx, database.DeleteEnvironmentParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete environment"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Environment deleted successfully"})
}
