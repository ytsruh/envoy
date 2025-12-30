package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/middleware"
	"ytsruh.com/envoy/server/utils"
	shared "ytsruh.com/envoy/shared"
)

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
	GitRepo     string `json:"git_repo" validate:"omitempty,max=500"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
	GitRepo     string `json:"git_repo" validate:"omitempty,max=500"`
}

type ProjectResponse struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	GitRepo     *string          `json:"git_repo"`
	OwnerID     shared.UserID    `json:"owner_id"`
	CreatedAt   shared.Timestamp `json:"created_at"`
	UpdatedAt   shared.Timestamp `json:"updated_at"`
}

func CreateProject(c echo.Context, ctx *HandlerContext) error {
	var req CreateProjectRequest
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

	if req.GitRepo != "" {
		_, err := ctx.Queries.GetProjectByGitRepo(dbCtx, database.GetProjectByGitRepoParams{
			OwnerID: claims.UserID,
			GitRepo: sql.NullString{String: req.GitRepo, Valid: true},
		})
		if err == nil {
			return SendErrorResponse(c, http.StatusConflict, fmt.Errorf("a project with this git repository already exists"))
		}
	}

	now := time.Now()
	projectID := utils.GenerateUUID()
	project, err := ctx.Queries.CreateProject(dbCtx, database.CreateProjectParams{
		ID:          projectID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		GitRepo:     sql.NullString{String: req.GitRepo, Valid: req.GitRepo != ""},
		OwnerID:     claims.UserID,
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   now,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create project"))
	}

	resp := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: shared.NullStringToStringPtr(project.Description),
		GitRepo:     shared.NullStringToStringPtr(project.GitRepo),
		OwnerID:     shared.UserID(project.OwnerID),
		CreatedAt:   shared.FromTime(project.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(project.UpdatedAt.(time.Time)),
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetProject(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("id")

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := ctx.Queries.GetAccessibleProject(dbCtx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found or access denied"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	resp := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: shared.NullStringToStringPtr(project.Description),
		GitRepo:     shared.NullStringToStringPtr(project.GitRepo),
		OwnerID:     shared.UserID(project.OwnerID),
		CreatedAt:   shared.FromTime(project.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(project.UpdatedAt.(time.Time)),
	}

	return c.JSON(http.StatusOK, resp)
}

func ListProjects(c echo.Context, ctx *HandlerContext) error {
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

	var resp []ProjectResponse
	for _, project := range projects {
		resp = append(resp, ProjectResponse{
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

func UpdateProject(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("id")

	var req UpdateProjectRequest
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

	canModify, err := ctx.Queries.CanUserModifyProject(dbCtx, database.CanUserModifyProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check permissions"))
	}
	if canModify == 0 {
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("access denied"))
	}

	originalProject, err := ctx.Queries.GetProject(dbCtx, projectID)
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	if req.GitRepo != "" && (NullStringToString(originalProject.GitRepo) != req.GitRepo) {
		_, err := ctx.Queries.GetProjectByGitRepo(dbCtx, database.GetProjectByGitRepoParams{
			OwnerID: originalProject.OwnerID,
			GitRepo: sql.NullString{String: req.GitRepo, Valid: true},
		})
		if err == nil {
			return SendErrorResponse(c, http.StatusConflict, fmt.Errorf("a project with this git repository already exists"))
		}
	}

	now := time.Now()
	updatedProject, err := ctx.Queries.UpdateProject(dbCtx, database.UpdateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		GitRepo:     sql.NullString{String: req.GitRepo, Valid: req.GitRepo != ""},
		UpdatedAt:   now,
		ID:          projectID,
		OwnerID:     originalProject.OwnerID,
	})

	resp := ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: shared.NullStringToStringPtr(updatedProject.Description),
		GitRepo:     shared.NullStringToStringPtr(updatedProject.GitRepo),
		OwnerID:     shared.UserID(updatedProject.OwnerID),
		CreatedAt:   shared.FromTime(updatedProject.CreatedAt.Time),
		UpdatedAt:   shared.FromTime(updatedProject.UpdatedAt.(time.Time)),
	}

	return c.JSON(http.StatusOK, resp)
}

func DeleteProject(c echo.Context, ctx *HandlerContext) error {
	projectID := c.Param("id")

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
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("only project owners can delete projects"))
	}

	err = ctx.Queries.DeleteProject(dbCtx, database.DeleteProjectParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        projectID,
		OwnerID:   claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to delete project"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
