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

type UserProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Server) AddUserToProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	var req AddUserRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := s.dbService.GetQueries().IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can add users"))
	}

	_, err = s.dbService.GetQueries().GetUser(ctx, req.UserID)
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("user not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to verify user"))
	}

	_, err = s.dbService.GetQueries().GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    req.UserID,
	})
	if err == nil {
		return sendErrorResponse(c, http.StatusConflict, fmt.Errorf("user is already a member of this project"))
	} else if err != sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check membership"))
	}

	now := time.Now()
	membership, err := s.dbService.GetQueries().AddUserToProject(ctx, database.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to add user to project"))
	}

	response := ProjectUserResponse{
		UserID:    membership.UserID,
		Role:      membership.Role,
		CreatedAt: membership.CreatedAt.Time,
	}

	return c.JSON(http.StatusCreated, response)
}

func (s *Server) RemoveUserFromProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	targetUserID := c.Param("user_id")

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := s.dbService.GetQueries().IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can remove users"))
	}

	_, err = s.dbService.GetQueries().GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("user is not a member of this project"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check membership"))
	}

	err = s.dbService.GetQueries().RemoveUserFromProject(ctx, database.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to remove user from project"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User removed from project successfully"})
}

func (s *Server) UpdateUserRole(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	targetUserID := c.Param("user_id")

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerCount, err := s.dbService.GetQueries().IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can update user roles"))
	}

	membership, err := s.dbService.GetQueries().GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("user is not a member of this project"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check membership"))
	}

	if membership.Role == req.Role {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("user already has this role"))
	}

	err = s.dbService.GetQueries().UpdateUserRole(ctx, database.UpdateUserRoleParams{
		Role:      req.Role,
		UpdatedAt: time.Now(),
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to update user role"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User role updated successfully"})
}

func (s *Server) GetProjectUsers(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.dbService.GetQueries().GetAccessibleProject(ctx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found or access denied"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	projectUsers, err := s.dbService.GetQueries().GetProjectUsers(ctx, projectID)
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project users"))
	}

	var response []ProjectUserResponse
	for _, user := range projectUsers {
		response = append(response, ProjectUserResponse{
			UserID:    user.UserID,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) ListUserProjects(c echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return sendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projects, err := s.dbService.GetQueries().GetUserProjects(ctx, database.GetUserProjectsParams{
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch projects"))
	}

	var response []UserProjectResponse
	for _, project := range projects {
		response = append(response, UserProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: nullStringToStringPtr(project.Description),
			OwnerID:     project.OwnerID,
			CreatedAt:   project.CreatedAt.Time,
			UpdatedAt:   project.UpdatedAt.(time.Time),
		})
	}

	return c.JSON(http.StatusOK, response)
}
