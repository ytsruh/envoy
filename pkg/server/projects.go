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

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
}

type ProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Server) CreateProject(c echo.Context) error {
	var req CreateProjectRequest
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

	now := time.Now()
	project, err := s.dbService.GetQueries().CreateProject(ctx, database.CreateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		OwnerID:     claims.UserID,
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   now,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create project"))
	}

	response := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: nullStringToStringPtr(project.Description),
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.(time.Time),
	}

	return c.JSON(http.StatusCreated, response)
}

func (s *Server) GetProject(c echo.Context) error {
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

	project, err := s.dbService.GetQueries().GetAccessibleProject(ctx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found or access denied"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	response := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: nullStringToStringPtr(project.Description),
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.(time.Time),
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) ListProjects(c echo.Context) error {
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

	var response []ProjectResponse
	for _, project := range projects {
		response = append(response, ProjectResponse{
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

func (s *Server) UpdateProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return sendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	var req UpdateProjectRequest
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

	canModify, err := s.dbService.GetQueries().CanUserModifyProject(ctx, database.CanUserModifyProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check permissions"))
	}
	if canModify == 0 {
		return sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("access denied"))
	}

	originalProject, err := s.dbService.GetQueries().GetProject(ctx, projectID)
	if err == sql.ErrNoRows {
		return sendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found"))
	} else if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	now := time.Now()
	updatedProject, err := s.dbService.GetQueries().UpdateProject(ctx, database.UpdateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   now,
		ID:          projectID,
		OwnerID:     originalProject.OwnerID,
	})

	response := ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: nullStringToStringPtr(updatedProject.Description),
		OwnerID:     updatedProject.OwnerID,
		CreatedAt:   updatedProject.CreatedAt.Time,
		UpdatedAt:   updatedProject.UpdatedAt.(time.Time),
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) DeleteProject(c echo.Context) error {
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

	ownerCount, err := s.dbService.GetQueries().IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check ownership"))
	}
	if ownerCount == 0 {
		return sendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can delete projects"))
	}

	err = s.dbService.GetQueries().DeleteProject(ctx, database.DeleteProjectParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        projectID,
		OwnerID:   claims.UserID,
	})
	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete project"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}
