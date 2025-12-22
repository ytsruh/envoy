package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

// URL path parameter extraction helper
func getProjectIDFromPath(r *http.Request) (int64, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		return 0, fmt.Errorf("project ID is required")
	}
	return strconv.ParseInt(pathParts[1], 10, 64)
}

type ProjectHandler interface {
	CreateProject(w http.ResponseWriter, r *http.Request) error
	GetProject(w http.ResponseWriter, r *http.Request) error
	ListProjects(w http.ResponseWriter, r *http.Request) error
	UpdateProject(w http.ResponseWriter, r *http.Request) error
	DeleteProject(w http.ResponseWriter, r *http.Request) error
}

type ProjectHandlerImpl struct {
	queries database.Querier
}

func NewProjectHandler(queries database.Querier) ProjectHandler {
	return &ProjectHandlerImpl{
		queries: queries,
	}
}

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

func (h *ProjectHandlerImpl) CreateProject(w http.ResponseWriter, r *http.Request) error {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return nil
	}

	if err := utils.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	project, err := h.queries.CreateProject(ctx, database.CreateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		OwnerID:     claims.UserID,
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   now,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create project"})
		return nil
	}

	response := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: nullStringToStringPtr(project.Description),
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.(time.Time),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandlerImpl) GetProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := h.queries.GetAccessibleProject(ctx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Project not found or access denied"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch project"})
		return nil
	}

	response := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: nullStringToStringPtr(project.Description),
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.(time.Time),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandlerImpl) ListProjects(w http.ResponseWriter, r *http.Request) error {
	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projects, err := h.queries.GetUserProjects(ctx, database.GetUserProjectsParams{
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch projects"})
		return nil
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandlerImpl) UpdateProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return nil
	}

	if err := utils.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user can modify the project
	canModify, err := h.queries.CanUserModifyProject(ctx, database.CanUserModifyProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check permissions"})
		return nil
	}
	if canModify == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Access denied"})
		return nil
	}

	// Get the original project for the update operation
	originalProject, err := h.queries.GetProject(ctx, projectID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Project not found"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch project"})
		return nil
	}

	now := time.Now()
	updatedProject, err := h.queries.UpdateProject(ctx, database.UpdateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   now,
		ID:          projectID,
		OwnerID:     originalProject.OwnerID, // Use original owner ID for update
	})

	response := ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: nullStringToStringPtr(updatedProject.Description),
		OwnerID:     updatedProject.OwnerID,
		CreatedAt:   updatedProject.CreatedAt.Time,
		UpdatedAt:   updatedProject.UpdatedAt.(time.Time),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandlerImpl) DeleteProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user is the project owner
	ownerCount, err := h.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check ownership"})
		return nil
	}
	if ownerCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only project owners can delete projects"})
		return nil
	}

	err = h.queries.DeleteProject(ctx, database.DeleteProjectParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        projectID,
		OwnerID:   claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete project"})
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "Project deleted successfully"})
}

func nullStringToStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
