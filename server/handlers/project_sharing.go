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
	shared "ytsruh.com/envoy/shared"
)

type AddUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=viewer editor"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=viewer editor"`
}

type ProjectUserResponse struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func AddUserToProject(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	var req AddUserRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := ctx.Queries.IsProjectOwner(dbCtx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can add users"))
	}

	_, err = ctx.Queries.GetUser(dbCtx, req.UserID)
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("user not found"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch user"))
	}

	projectUser, err := ctx.Queries.AddUserToProject(dbCtx, database.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to add user to project"))
	}

	resp := ProjectUserResponse{
		UserID:    projectUser.UserID,
		Role:      projectUser.Role,
		CreatedAt: projectUser.CreatedAt.Time,
	}

	return c.JSON(http.StatusCreated, resp)
}

func RemoveUserFromProject(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	userID := c.Param("user_id")
	if userID == "" {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := ctx.Queries.IsProjectOwner(dbCtx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can remove users"))
	}

	err = ctx.Queries.RemoveUserFromProject(dbCtx, database.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to remove user from project"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User removed successfully"})
}

func UpdateUserRole(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	userID := c.Param("user_id")
	if userID == "" {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := ctx.Queries.IsProjectOwner(dbCtx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can update user roles"))
	}

	err = ctx.Queries.UpdateUserRole(dbCtx, database.UpdateUserRoleParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      req.Role,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update user role"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User role updated successfully"})
}

func GetProjectUsers(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	_, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectUsers, err := ctx.Queries.GetProjectUsers(dbCtx, projectID)
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project users"))
	}

	var resp []ProjectUserResponse
	for _, pu := range projectUsers {
		resp = append(resp, ProjectUserResponse{
			UserID:    pu.UserID,
			Role:      pu.Role,
			CreatedAt: pu.CreatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func ListUserProjects(c echo.Context, ctx *HandlerContext) error {
	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projects, err := ctx.Queries.GetUserProjects(dbCtx, database.GetUserProjectsParams{
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch projects"))
	}

	var resp []shared.UserProjectResponse
	for _, project := range projects {
		resp = append(resp, shared.UserProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: shared.NullStringToStringPtr(project.Description),
			GitRepo:     shared.NullStringToStringPtr(project.GitRepo),
			OwnerID:     shared.UserID(project.OwnerID),
			CreatedAt:   shared.FromTime(project.CreatedAt.Time),
			UpdatedAt:   shared.FromTime(project.UpdatedAt.(time.Time)),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
